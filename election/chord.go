package main

import (
	"crypto/sha1"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

var allnodes []*Node //global array to account all available nodes in ascending order
var leaderID int
var leader LeaderElection
var allnodesID []int

// represents the client's information
type info struct {
	name     string
	Username string
	password string
}

// each node has a finger table to oversee other nodes it can go to
type Node struct {
	ID               int
	Successor        *Node
	Finger           []*Node
	storage          []info
	leader           int
	OwnPublicKey     int //all nodes know leader's public as well
	LeadersPublicKey int
	privateKey       int
	lock             sync.Mutex
}

// either a node is created to add the user or the uder is added to a preexisting node
func createUser(id int, new info) *Node {
	node := &Node{
		ID:               id,
		Successor:        nil,
		Finger:           make([]*Node, m),
		LeadersPublicKey: 420,
		OwnPublicKey:     699,
	}

	node.storage = append(node.storage, new)

	var check bool = false
	var j int = 0
	for _, value := range allnodes {
		if node.ID == value.ID {
			value.storage = append(value.storage, new)
			check = true
			break
		} else if node.ID < value.ID {
			allnodes = append(allnodes[:j+1], allnodes[j:]...)
			//leader.nodes = append(leader.nodes, id)
			allnodesID = append(allnodesID, id)
			allnodes[j] = node
			check = true
			break
		}
		j = j + 1
	}

	if !check {
		allnodes = append(allnodes, node)
		//leader.nodes = append(leader.nodes, id)
		allnodesID = append(allnodesID, id)
	}

	updateFinger()

	//electLeader(int)
	return nil
}

// used after a new node enters
func updateFinger() {
	for _, value := range allnodes {
		for i := 0; i < m; i++ {
			var done bool = false
			var total = value.ID + int(math.Pow(2, float64(i)))
			total = total % bits
			if total > value.ID {
				for k := total; k < bits; k++ {
					for _, nodeCheck := range allnodes {
						if nodeCheck.ID == k {
							value.Finger[i] = nodeCheck
							done = true
							break
						}
					}
					if done {
						break
					}
				}
				//total has looped to begining of chord
			} else {
				for k := total; k < value.ID; k++ {
					for _, nodeCheck := range allnodes {
						if nodeCheck.ID == k {
							value.Finger[i] = nodeCheck
							done = true
							break
						}
					}
					if done {
						break
					}
				}
			}
			if !done {
				value.Finger[i] = allnodes[0]
				done = true
			}
		}
		value.Successor = value.Finger[0]
	}
}

// returns true or false if name exists
func lookup(s string) bool {
	var hashValue int
	hashValue = hash(s)
	if allnodes == nil {
		return false
	}
	var currentNode *Node = allnodes[0]

	for hashValue > currentNode.ID {
		if hashValue == currentNode.ID {
			for _, infoCheck := range currentNode.storage {
				if s == infoCheck.name {
					return true
				}
			}
			return false
		} else if hashValue < currentNode.ID {
			var reference int = 0
			for _, fingerEntry := range currentNode.Finger {
				if hashValue == fingerEntry.ID {
					for _, infoCheck := range fingerEntry.storage {
						if s == infoCheck.name {
							return true
						}
					}
					return false
				} else if hashValue < fingerEntry.ID {
					reference = reference + 1
				}
			}
			currentNode = currentNode.Finger[reference]
		} else {
			return false
		}
	}
	return false
}

// Users to prevent duplicate names
func userCreation(s string, p string) {
	if lookup(s) {
		fmt.Println("This name has been picked before. You need to pick another name")
	} else if !lookup(s) {
		var newUser info
		newUser.name = s
		newUser.password = p
		createUser(hash(s), newUser)
	}
}

// hash function to convert to 8 bits
func hash(key string) int {
	hash := sha1.Sum([]byte(key))
	return int(hash[0])
}

func printFinger(n *Node) {
	fmt.Printf("Node with id %d ", n.ID)
	fmt.Println(" ")
	for i := 0; i < m; i++ {
		fmt.Println(n.Finger[i].ID)
	}
}

func printNames(n *Node) {
	fmt.Printf("Node with id %d ", n.ID)
	fmt.Println(" ")
	for i := 0; i < len(n.storage); i++ {
		fmt.Println(n.storage[i].name)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// random string of length
func randomString(length int) string {
	return StringWithCharset(length, charset)
}

// load the ring by a number of randoms users (startingUsers)
func loadRing() {
	for i := 0; i < startingUsers; i++ {
		userCreation(randomString(12), "password")
	}
}

func testSpeed() {
	start := time.Now()

	var entireLookup bool = false

	for _, value := range allnodes {
		for _, value2 := range value.storage {
			if value2.name == "testuser" {
				entireLookup = true
			}
		}
	}

	fmt.Println(entireLookup)

	elapsed := time.Since(start)

	fmt.Println("Entire lookup lookup took ", elapsed)

	start = time.Now()

	dhsLookup := lookup("testuser")

	fmt.Println(dhsLookup)

	elapsed = time.Since(start)

	fmt.Println("Chord DHS lookup took ", elapsed)
}

const (
	m             = 8
	bits          = 256
	startingUsers = 1000
)
