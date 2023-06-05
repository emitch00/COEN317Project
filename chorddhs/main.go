package main

import (
	"crypto/sha1"
	"fmt"
	"math"
)

var allnodes []*Node

type info struct {
	name string
}

type Node struct {
	ID        int
	Successor *Node
	Finger    []*Node
	storage   []info
}

func createNode(id int, new info) *Node {
	node := &Node{
		ID:        id,
		Successor: nil,
		Finger:    make([]*Node, m),
	}

	node.storage = append(node.storage, new)

	for i := 0; i < m; i++ {
		node.Finger[i] = node
	}

	var check bool = false
	var j int = 0
	for _, value := range allnodes {
		if node.ID < value.ID {
			allnodes = append(allnodes[:j+1], allnodes[j:]...)
			allnodes[j] = node
			check = true
			break
		}
		j = j + 1
	}

	if !check {
		allnodes = append(allnodes, node)
	}

	updateFinger()
	return nil
}

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

func lookup(s string) bool {
	var hashValue int
	hashValue = hash(s)
	if allnodes == nil {
		return false
	}
	var currentNode *Node = allnodes[0]
	for hashValue >= currentNode.ID {
		if hashValue == currentNode.ID {
			for _, infoCheck := range currentNode.storage {
				if s == infoCheck.name {
					return true
				}
			}
			return false
		} else if hashValue < currentNode.ID {
			for _, fingerEntry := range currentNode.Finger {
				if fingerEntry.ID == hashValue {
					for _, infoCheck := range fingerEntry.storage {
						if s == infoCheck.name {
							return true
						}
					}
					return false
				}
			}
			currentNode = currentNode.Successor
		} else {
			currentNode = currentNode.Successor
		}
	}
	return false
}

func userCreation(s string) {
	if lookup(s) {
		fmt.Println("This name has been picked before")
	} else if !lookup(s) {
		var newUser info
		newUser.name = s
		createNode(hash(s), newUser)
	}
}

func hash(key string) int {
	hash := sha1.Sum([]byte(key))
	return int(hash[0])
}

const (
	m    = 8
	bits = 256
)

func printFinger(n *Node) {
	fmt.Printf("Node with id %d ", n.ID)
	fmt.Println(" ")
	for i := 0; i < m; i++ {
		fmt.Println(n.Finger[i].ID)
	}
}

func main() {
	// Create nodes
	userCreation("node1")
	userCreation("node2")
	userCreation("node3")
	userCreation("node4")
	userCreation("node5")
	/*
		node1 := createNode(hash("node1"), "node1")
		node2 := createNode(hash("node2"), "node2")
		node3 := createNode(hash("node3"), "node3")
		node4 := createNode(hash("node4"), "node4")
	*/
	printFinger(allnodes[0])
	printFinger(allnodes[1])
	printFinger(allnodes[2])
	printFinger(allnodes[3])
	printFinger(allnodes[4])
}
