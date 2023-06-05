package leaderelection


import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sync"

	"time"

	"github.com/go-zookeeper/zk"
)

const (
	Follower = iota
	Leader
)

const electionOver = "DONE"

type (
	Role int
	Election struct {
		status chan Status
		resign chan struct{}
		end chan struct{}
		zkEvents <-chan zk.Event
		triggerElection chan error
		stopZKWatch chan struct{}
		zkConn           *zk.Conn 
		ElectionResource string   
		clientName       string 
		candidateID string 
		once sync.Once
		doneChl <-chan zk.Event
	}
	Status struct {
		CandidateID string
		Err         error
		NowFollowing string
		Role         Role  
		WasFollowing string 
	}
)
func (status Status) String() string {
	var statusErr string
	if status.Err == nil {
		statusErr = "<nil>"
	} else {
		statusErr = status.Err.Error()
	}
	return "Status:" +
		"    CandidateID:" + status.CandidateID +
		"    NowFollowing:" + status.NowFollowing +
		"    Role:" + strconv.Itoa(int(status.Role)) +
		"    WasFollowing:" + status.WasFollowing +
		"    Err:" + statusErr
}
func NewElection(zkConn *zk.Conn, electionResource string, clientName string) (*Election, error) {
	if zkConn == nil {
		return nil, errors.New("zkConn must not be nil")
	}
	exists, _, err := zkConn.Exists(electionResource)
	if err != nil {
		return nil, fmt.Errorf("%s Error validating electionResource (%v). Error (%v) returned",
			"NewElection:", electionResource, err)
	} else if !exists {
		return nil, fmt.Errorf("%s Provided electionResource: <%q>: does not exist", "NewElection:", electionResource)
	}

	_, _, electionDoneChl, err := zkConn.ExistsW(strings.Join([]string{electionResource, electionOver}, "/"))
	if err != nil {
		return nil, fmt.Errorf("%s Error setting watch on DONE node", "NewElection:")
	}

	if clientName == "" {
		clientName = "name-not-provided"
	}
	return &Election{
		status: make(chan Status, 1),
		resign: make(chan struct{}),
		end:    make(chan struct{}),
		triggerElection: make(chan error, 1), 
		stopZKWatch:      make(chan struct{}),
		zkConn:           zkConn,
		ElectionResource: electionResource,
		doneChl:          electionDoneChl,
		clientName:       clientName,
	}, nil
}
func (le *Election) ElectLeader() {

	cleanupOnlyOnce := func() {
		cleanup(le)
	}
	status := runElection(le)

	le.status <- status
	for {
		select {
		case err := <-le.triggerElection:
			if err != nil {
				status.Err = err
				} else {
					determineLeader(le, &status)
				}
				select {
				case le.status <- status:
					if err != nil {
						resignElection(le, cleanupOnlyOnce)
						return
					}
				case <-le.resign:
					resignElection(le, cleanupOnlyOnce)
					return
				case <-le.end:
					endElection(le, cleanupOnlyOnce)
					return
				}
			case <-le.resign:
				resignElection(le, cleanupOnlyOnce)
				return
			case <-le.end:
				endElection(le, cleanupOnlyOnce)
				return
			}
		}
	}

	func (le *Election) EndElection() {
		close(le.end)
	}

	func (le *Election) Resign() {
		close(le.resign)
	}

	func (le *Election) Status() <-chan Status {
		return le.status
	}
	
	func (le *Election) String() string {
		return "Election:" +
			"\tCandidateID:" + le.candidateID +
			"\tResource:" + le.ElectionResource
	}

	func resignElection(le *Election, cleanupOnlyOnce func()) {
		close(le.stopZKWatch)
		err := le.zkConn.Delete(le.candidateID, -1)
		if err != nil {
			// fmt.Println("resignElection:"+"Error deleting ZK node for candidate <", le.candidateID, "> during Resign. Error is ", err)
		}
	le.once.Do(cleanupOnlyOnce)
	return
}

func endElection(le *Election, cleanupOnlyOnce func()) {
	close(le.stopZKWatch)
	deleteAllCandidates(le)
	le.once.Do(cleanupOnlyOnce)
	return
}

func runElection(le *Election) Status {
	status := Status{}
	zkWatchChl := makeOffer(le, &status)
	if status.Err != nil {
		status.Err = fmt.Errorf("%s Unexpected error attempting to nominate candidate. Error <%v>",
			"runElection:", status.Err)
		return status
	}
	le.zkEvents = zkWatchChl
	le.candidateID = status.CandidateID

	determineLeader(le, &status)
	if status.Err != nil {
		status.Err = fmt.Errorf("%s Unexpected error attempting to determine leader. Error (%v)",
			"runElection:", status.Err)
		return status
	}
	
	return status
}


func makeOffer(le *Election, status *Status) <-chan zk.Event {
	flags := int32(zk.FlagSequence | zk.FlagEphemeral)
	acl := zk.WorldACL(zk.PermAll)

	candidates, err := getCandidates(le, status)
	if err != nil {
		status.Err = fmt.Errorf("%s Received error (%v) attempting to retrieve candidates",
			"makeOffer:", err)
		return nil
	}

	if isElectionOver(candidates) {
		status.Err = fmt.Errorf("%s %s", "makeOffer:", ElectionCompletedNotify)
		return nil
	}
	cndtID, err := le.zkConn.Create(strings.Join([]string{le.ElectionResource, "le_"}, "/"), []byte(le.clientName), flags, acl)
	if err != nil {
		status.Err = fmt.Errorf("%s Error <%v> creating candidate node for election <%v>",
			"makeOffer:", err, le.ElectionResource)
		return nil
	}

	exists, _, watchChl, err := le.zkConn.ExistsW(cndtID)
	if err != nil {
		status.Err = fmt.Errorf("%s Received error (%v) attempting to watch newly created node (%v)",
			"makeOffer:", err, cndtID)
		return nil
	}
	if !exists {
		status.Err = fmt.Errorf("%s Newly created node (%v) doesn't exist - %s", "makeOffer:", cndtID,
			ElectionCompletedNotify)
		return nil
	}
	status.CandidateID = cndtID
	return watchChl
}

func determineLeader(le *Election, status *Status) {
	candidates, err := getCandidates(le, status)
	if err != nil {
		status.Err = fmt.Errorf("%s %v", "determineLeader:", err)
		return
	}
	if len(candidates) == 0 {
		// fmt.Printf("determineLeader: "+"No children exist in ZK, not even me:%s\n", le.candidateID)
		status.Err = fmt.Errorf("%s No Leader candidates exist in ZK. Candidate requesting children is (%v)",
			"determineLeader:", status.CandidateID)
		return
	}

	if isElectionOver(candidates) {
		status.Err = fmt.Errorf("%s %s", "determineLeader:", ElectionCompletedNotify)
		return
	}

	iAmLeader, shortCndtID := amILeader(status, candidates, le)
	if iAmLeader {
		return
	}

	followAnotherCandidate(le, shortCndtID, status, candidates)
	return
}

func getCandidates(le *Election, status *Status) ([]string, error) {
	candidates, _, err := le.zkConn.Children(le.ElectionResource)
	if err != nil {
		status.Err = fmt.Errorf("%s Received error getting candidates from ZK. Error is <%v>. "+
			"Candidate requesting children is <%v>.", "getCandidates:", err, status.CandidateID)
		return nil, status.Err
	}
	return candidates, status.Err
}

func amILeader(status *Status, candidates []string, le *Election) (bool, string) {
	pathNodes := strings.Split(status.CandidateID, "/")
	lenPath := len(pathNodes)
	shortCndtID := pathNodes[lenPath-1]
	sort.Strings(candidates)
	if strings.EqualFold(shortCndtID, candidates[0]) {
		status.Role = Leader
		status.WasFollowing = status.NowFollowing
		status.NowFollowing = ""

		// Detect if the leader node is unexpectedly deleted
		go watchForLeaderDeleteEvents(le.zkEvents, le.triggerElection, le.stopZKWatch, le.candidateID,
			le.doneChl)
	}
	return status.Role == Leader, shortCndtID
}

func followAnotherCandidate(le *Election, shortCndtID string, status *Status, candidates []string) {
	watchedNode, err := findWhoToFollow(candidates, shortCndtID, le)
	if err != nil && strings.Contains(err.Error(), leaderWhileDeterminingWhoToFollow) {
		le.triggerElection <- nil //
		return
	}
	if err != nil {
		status.Err = fmt.Errorf("%s %v", "followAnotherCandidate:", err)
		return
	}

	exists, _, zkFollowingWatchChl, err := le.zkConn.ExistsW(watchedNode)
	if err != nil {
		status.Err = fmt.Errorf("%s ZK error checking existence of %q. Error is (%v)",
			"followAnotherCnadidate:", watchedNode, err)
		return
	}
	if !exists {
		le.triggerElection <- nil
	}
	status.WasFollowing = status.NowFollowing
	status.NowFollowing = watchedNode
	status.Role = Follower

	go watchForFollowerEvents(zkFollowingWatchChl, le.zkEvents, le.triggerElection, le.stopZKWatch, le.candidateID,
		le.doneChl)
}

func watchForLeaderDeleteEvents(ldrDltWatchChl <-chan zk.Event, notifyChl chan<- error,
	stopZKWatch <-chan struct{}, candidateID string, electionCompleted <-chan zk.Event) {
	// Continue until a zk.EventNodeDeleted message is received or a message is received on the stopZKWatch channel.
	for {
		select {
		case doneWatchEvent := <-electionCompleted:
			notifyChl <- fmt.Errorf("%s - %s - Election 'DONE' node created or deleted <%v>, the election is over"+
				" for candidate <%s>.", "watchForLeaderDeleteEvents", ElectionCompletedNotify, doneWatchEvent, candidateID)
			return

		case delWatchEvent := <-ldrDltWatchChl:
			if delWatchEvent.Type == zk.EventNodeDeleted {
				err := fmt.Errorf("%s Leader (%v) has been deleted", "watchForLeaderDeleteEvents", candidateID)
				notifyChl <- err
				return
			}
			return
		case <-stopZKWatch:
			return
		}
	}
}

func watchForFollowerEvents(followingWatchChnl <-chan zk.Event, selfDltWatchChl <-chan zk.Event, notifyChl chan<- error,
	stopZKWatch <-chan struct{}, candidateID string, electionCompleted <-chan zk.Event) {
	// Continue until a zk.EventNodeDeleted message is received or a message is received on the stopZKWatch channel.
	for {
		select {
		case doneWatchEvent := <-electionCompleted:
			notifyChl <- fmt.Errorf("%s - %s - Election 'DONE' node created or deleted <%v>, the election is over"+
				" for candidate <%s>.", "watchForFollowerEvents", ElectionCompletedNotify, doneWatchEvent, candidateID)
			return

		case watchEvt := <-followingWatchChnl:
			if watchEvt.Type == zk.EventNodeDeleted {
				notifyChl <- nil
				return
			}

		case delWatchEvent := <-selfDltWatchChl:
			if delWatchEvent.Type == zk.EventNodeDeleted {
				err := fmt.Errorf("%s - %s - Candidate (%s) (i.e., me) has been deleted", "watchForFollowerEvents",
					ElectionSelfDltNotify, candidateID)
				notifyChl <- err
				return
			}
		case <-stopZKWatch:
			return
		}
	}
}

func findWhoToFollow(candidates []string, shortCndtID string, le *Election) (string, error) {
	idx := sort.SearchStrings(candidates, shortCndtID)
	if idx == 0 {
		return "", fmt.Errorf("%s %s Candidate is <%s>", "findWhoToFollow", leaderWhileDeterminingWhoToFollow, shortCndtID)
	}
	if idx == len(candidates) {
		return "", fmt.Errorf("%s Error finding candidate in child list. Candidate attempting match is (%s) - %v",
			"findWhoToFollow", shortCndtID, ElectionCompletedNotify)
	}
	if !strings.EqualFold(candidates[idx], shortCndtID) {
		return "", fmt.Errorf("%s Error finding candidate in child list. Candidate attempting match is (%v). "+
			"Candidate located at position <%d> is <%s> - %v",
			"findWhoToFollow", shortCndtID, idx, candidates[idx], ElectionCompletedNotify)
	}

	nodeToWatchIdx := idx - 1

	watchedNode := strings.Join([]string{le.ElectionResource, candidates[nodeToWatchIdx]}, "/")
	return watchedNode, nil
}

func cleanup(le *Election) {
	close(le.status)
}

func deleteAllCandidates(le *Election) {
	doneNodePath := strings.Join([]string{le.ElectionResource, electionOver}, "/")
	DeleteCandidates(le.zkConn, le.ElectionResource, doneNodePath)
}

// DeleteElection removes the election resource passed to NewElection.
func DeleteElection(zkConn *zk.Conn, electionResource string) error {
	doneNodePath, err := addDoneNode(zkConn, electionResource)
	if err != nil {
		return fmt.Errorf("%s Error adding 'DONE' node: <%v>", "DeleteElection", err)
	}

	err = DeleteCandidates(zkConn, electionResource, doneNodePath)
	if err != nil {
		return fmt.Errorf("%s Error deleting candidates: <%v>", "DeleteElection", err)
	}
	err = deleteDoneNodeAndElection(zkConn, electionResource, doneNodePath)
	if err != nil && err != zk.ErrNoNode {
		exists, _, err2 := zkConn.Exists(electionResource)
		if err2 == nil && !exists {
			} else {
				return fmt.Errorf("%s Unexpected error received from leaderelection.DeleteElection: %v", "DeleteElection", err)
			}
		} 
		
	return nil
}

// DeleteCandidates will remove all the candidates for the provided electionName.
func DeleteCandidates(zkConn *zk.Conn, electionName string, doneNodePath string) error {
	retries := 1
	for i := 0; i < retries; i++ {
		jobWorkers, _, err := zkConn.Children(electionName)
		if err != nil && err != zk.ErrNoNode {
			return fmt.Errorf("%s: Received error getting candidates from ZK. Error is <%v>. \n", "DeleteCandidates", err)
		}

		err = runDeleteCandidates(zkConn, electionName, jobWorkers, doneNodePath)
		if err != nil {
			return fmt.Errorf("%s: Received error deleting candidates. Error is <%v>. \n", "DeleteCandidates", err)
		}
	}

	return nil
}

func runDeleteCandidates(zkConn *zk.Conn, electionName string, candidates []string, doneNodePath string) error {
	numCndts := len(candidates)
	if numCndts <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(candidates)))
	var deleteWorkerRqsts []interface{}
	for _, shortCndt := range candidates {
		longCndt := strings.Join([]string{electionName, shortCndt}, "/")
		if strings.Contains(longCndt, doneNodePath) {
			continue
		}
		deleteWorkerOp := zk.DeleteRequest{Path: longCndt}
		deleteWorkerRqsts = append(deleteWorkerRqsts, &deleteWorkerOp)
	}
	var err error
	for i := 0; i < 10; i++ {
		_, err = zkConn.Multi(deleteWorkerRqsts...)
		if err == nil {
			break
		}
		if err != nil {
			jobWorkers, _, zkErr := zkConn.Children(electionName)
			if zkErr != nil {
				continue
			}
			if len(jobWorkers) == 1 { 
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func addDoneNode(zkConn *zk.Conn, electionName string) (string, error) {
	var flags int32
	acl := zk.WorldACL(zk.PermAll)
	nodePath, err := zkConn.Create(strings.Join([]string{electionName, electionOver}, "/"), []byte("done"),
		flags, acl)
	if err != nil && err != zk.ErrNoNode && err != zk.ErrNodeExists {
			return "", fmt.Errorf("%s Error adding 'ElectionOver` (aka done) node. Error: <%v>", "addDoneNode", err)
	}
	if err != nil && err == zk.ErrNodeExists {
		nodePath = strings.Join([]string{electionName, electionOver}, "/")
	}

	return nodePath, nil
}

func deleteDoneNodeAndElection(zkConn *zk.Conn, electionName, doneNodePath string) error {
	deleteDoneNode := zk.DeleteRequest{Path: doneNodePath, Version: -1}
	deleteElectionNode := zk.DeleteRequest{Path: electionName, Version: -1}
	deleteRequests := []interface{}{&deleteDoneNode, &deleteElectionNode}

	var err error
	for i := 0; i < 10; i++ {
		_, err = zkConn.Multi(deleteRequests...)
		if err == nil {
			break
		}
		exists, _, err2 := zkConn.Exists(electionName)
		if err2 == nil && !exists {
			break
		}
		if err2 == zk.ErrNoNode {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("%s Error <%v> deleting ZK 'done' <%s> and election <%s> nodes",
			"deleteDoneNodeAndElection", err, doneNodePath, electionName)
	}

	return nil
}

func isElectionOver(candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(candidate, electionOver) {
			return true
		}
	}
	return false
}




