package main

import (
	//"strconv"
	"sync"
)

type LeaderElection struct {
	nodes      []int
	leaderID   int
	monitorCh  chan int
	candidates map[int]bool
	lock       sync.Mutex
}

func NewLeaderElection(nodes []int) *LeaderElection {
	le := &LeaderElection{
		nodes:      nodes,
		leaderID:   0,
		monitorCh:  make(chan int),
		candidates: make(map[int]bool),
	}

	go le.monitorLeadership()

	return le
}

func (le *LeaderElection) ElectLeader() (int, error) {
	le.lock.Lock()
	defer le.lock.Unlock()

	// Check if the current leader is still a candidate
	if _, ok := le.candidates[le.leaderID]; ok {
		// Current leader is still available, return the current leader's ID
		return le.leaderID, nil
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
