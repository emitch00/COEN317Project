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
	//nodeIDs := []int{1, 2, 3, 4, 5}

	fmt.Println("node id is", allnodes[17].ID)
	fmt.Println("node own public key is", allnodes[17].OwnPublicKey)
	fmt.Println("leader is", allnodes[17].leader)
	fmt.Println("leader public key is", allnodes[17].LeadersPublicKey)

	fmt.Println("node id is", allnodes[20].ID)
	fmt.Println("node own public key is", allnodes[20].OwnPublicKey)
	fmt.Println("leader is", allnodes[20].leader)
	fmt.Println("leader public key is", allnodes[20].LeadersPublicKey)

	fmt.Println("before leader")
	leaderElection := NewLeaderElection(allnodes)
	leaderID, err := leaderElection.ElectLeader()
	if err != nil {
		fmt.Println("Error electing leader:", err)
		return
	}

	fmt.Println("Leader elected:", leaderID)

	fmt.Println("node id is", allnodes[17].ID)
	fmt.Println("node own public key is", allnodes[17].OwnPublicKey)
	fmt.Println("leader is", allnodes[17].leader)
	fmt.Println("leader public key is", allnodes[17].LeadersPublicKey)

	fmt.Println("node id is", allnodes[20].ID)
	fmt.Println("node own public key is", allnodes[20].OwnPublicKey)
	fmt.Println("leader is", allnodes[20].leader)
	fmt.Println("leader public key is", allnodes[20].LeadersPublicKey)

	//printFinger(allnodes[0])
	//printNames(allnodes[0])
}
