package main

import "fmt"

func main() {
	// Create nodes
	/*
		userCreation("node1")
		userCreation("node2")
		userCreation("node3")
		userCreation("node4")
		userCreation("node5")
	*/

	loadRing()

	var err error
	nodeslist := []int{7, 2, 3, 4, 5}
	var leader2 = NewLeaderElection(nodeslist)
	leaderID, err = leader2.ElectLeader(hash("node2"))
	if err != nil {
		fmt.Print("Error electing leader", err)
		return
	}
	fmt.Println("The leader is...")
	fmt.Println(leaderID)
//
	nodeIDs := []int{1, 2, 3, 4, 5}
	leaderElection := NewLeaderElection(nodeIDs)

	nodeID := 3
	leaderID, err := leaderElection.ElectLeader(nodeID)
	if err != nil {
		fmt.Println("Error electing leader:", err)
		return
	}

	fmt.Println("Leader elected:", leaderID)


	//printFinger(allnodes[0])
	//printNames(allnodes[0])
}
