package main

import (
	"sync"
	"fmt"
)

type LeaderElection struct {
	nodes      []*Node
	leaderID   int
	monitorCh  chan int
	candidates map[int]bool
	lock       sync.Mutex
}

func NewLeaderElection(nodes []*Node) *LeaderElection {
	le := &LeaderElection{
		nodes:      nodes,
		leaderID:   -1,
		monitorCh:  make(chan int),
		candidates: make(map[int]bool),
	}

	go le.monitorLeadership()

	return le
}

func (le *LeaderElection) ElectLeader() (int, *Node, error) {
	le.lock.Lock()
	defer le.lock.Unlock()

	// Check if the current leader is still a candidate
	if _, ok := le.candidates[le.leaderID]; ok {
		// Current leader is still available, return the current leader's ID
		//var leaderIndex int = 0

		//for _, value := range le.candidates {
			//if value.ID == leaderID {
		//		break
		//	}
		//leaderIndex = leaderIndex + 1
		//}
		var leaderIndex int = 0
		for _, value := range le.nodes {
			if value.ID == leaderID {
				break
			}
		leaderIndex = leaderIndex + 1
		}

		return le.leaderID, le.nodes[leaderIndex], nil
	}

	// Find the candidate with the highest ID
	var maxID int
	for candidateID := range le.candidates {
		if candidateID > maxID {
			maxID = candidateID
		}
	}

	// Set the candidate with the highest ID as the leader
	le.leaderID = maxID

	var leaderIndex int = -1
	for _, value := range le.nodes {
		if value.ID == leaderID {
			break
		}
		//fmt.Println("incrementing leaderIndex")
		leaderIndex = leaderIndex + 1
	}

	// Update leader information on all other nodes
	fmt.Println("LeaderIndex:", leaderIndex)
	fmt.Println("LeaderIndex Public Key:", le.nodes[leaderIndex].OwnPublicKey)
	BroadcastUpdateLeaderInfo(le.nodes, le.leaderID, le.nodes[leaderIndex].OwnPublicKey)

	return le.leaderID, le.nodes[leaderIndex], nil
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

		var leaderIndex int = 0
		for _, value := range le.nodes {
			if value.ID == leaderID {
				break
			}
			leaderIndex = leaderIndex + 1
		}

		// Update leader information on all other nodes
		BroadcastUpdateLeaderInfo(le.nodes, le.leaderID, le.nodes[leaderIndex].OwnPublicKey)

		le.lock.Unlock()
	}
}

func (le *LeaderElection) UpdateLeader(leaderID int) {
	le.monitorCh <- leaderID
	// Update leader information on all other nodes

	var leaderIndex int = 0
	for _, value := range le.nodes {
		if value.ID == leaderID {
			break
		}
		leaderIndex = leaderIndex + 1
	}

	BroadcastUpdateLeaderInfo(le.nodes, le.leaderID, le.nodes[leaderIndex].OwnPublicKey)
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

	return candidates
}

func (n *Node) UpdateLeader(leaderID, leadersPublicKey int) {
	n.leader = leaderID
	n.LeadersPublicKey = leadersPublicKey
}

func BroadcastUpdateLeaderInfo(nodes []*Node, leaderID, leadersPublicKey int) {
	for _, node := range nodes {
		if node.ID != leaderID {
			node.UpdateLeaderInfo(leaderID, leadersPublicKey)
		}
	}
}

func (n *Node) UpdateLeaderInfo(leaderID, leadersPublicKey int) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.leader = leaderID
	n.LeadersPublicKey = leadersPublicKey
}
