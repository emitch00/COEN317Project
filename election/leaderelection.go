package leaderelection

import (
	//"strconv"
	"sync"
)

type LeaderElection struct {
	nodes       []Node
	leaderID    int
	monitorCh   chan int
	candidates  map[int]bool
	lock        sync.Mutex
}

func NewLeaderElection(nodes []Node) *LeaderElection {
	le := &LeaderElection{
		nodes:       nodes,
		leaderID:    0,
		monitorCh:   make(chan int),
		candidates:  make(map[int]bool),
	}

	go le.monitorLeadership()

	return le
}

func (le *LeaderElection) ElectLeader(nodeID int) (int, error) {
	le.lock.Lock()
	defer le.lock.Unlock()

	le.candidates[nodeID] = true

	// Check if the leader has already been elected
	if le.leaderID != -1 {
		return le.leaderID, nil
	}

	// Elect the leader from the candidates
	for candidateID := range le.candidates {
		le.leaderID = candidateID
		break
	}

	return le.leaderID, nil
}



func (le *LeaderElection) monitorLeadership() {
	for {
		leaderID := <-le.monitorCh
		le.lock.Lock()

		// Update the leader ID
		le.leaderID = leaderID

		// Remove all candidates except for the leader
		le.candidates = make(map[int]bool)
		le.candidates[leaderID] = true

		le.lock.Unlock()
	}
}

func (le *LeaderElection) UpdateLeader(leaderID int) {
	le.monitorCh <- leaderID
}

func (le *LeaderElection) GetLeaderID() int {
	le.lock.Lock()
	defer le.lock.Unlock()

	return le.leaderID
}

func (le *LeaderElection) GetCandidates() []int {
	le.lock.Lock()
	defer le.lock.Unlock()

	var candidates []int
	for candidate := range le.candidates {
		candidates = append(candidates, candidate)
	}

	return candidates
}