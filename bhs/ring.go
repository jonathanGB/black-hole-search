package bhs

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// whiteboard is an abstraction of the sync.Map structure
type whiteboard struct {
	sync.Map
}

// describes how to stringify the whiteboard
func (wb *whiteboard) String() string {
	var buffer bytes.Buffer
	wb.Range(func(key, val interface{}) bool {
		buffer.WriteString(fmt.Sprintf("(%s => %s) ", key, val))
		return true
	})

	return buffer.String()
}

// Node contains the information of a node, as well helper functions to navigate through Nodes
type Node struct {
	Bh bool
	ID int
	wb *whiteboard
}

// Ring defines the structure of a Ring network
type Ring []*Node

// HOMEBASE is a helper index to index 0
var HOMEBASE int

// describes how to stringify the ring
func (r Ring) String() string {
	str := make([]string, 0, len(r)+1)
	for _, n := range r {
		str = append(str, fmt.Sprintf("{ ID: %d | Black hole: %t | Whiteboard: %s}\n", n.ID, n.Bh, n.wb))
	}
	str = append(str, "\n")

	return strings.Join(str, "")
}

// BuildRing creates a Ring network made of Nodes
// Requires the position of the black hole, the number of nodes, and whether Nodes should include whiteboards
// The ring is returned with an error if the black hole position is out of obunds
func BuildRing(bhPos, len int, hasWhiteBoards bool) (Ring, error) {
	if bhPos < 0 || bhPos >= len {
		return nil, fmt.Errorf("position of black hole is out of bounds")
	}

	r := make([]*Node, 0, len)

	for i := 0; i < len; i++ {
		var isBlackHole bool
		if i == bhPos {
			isBlackHole = true
		}

		var wb *whiteboard
		if hasWhiteBoards {
			wb = &whiteboard{}
		}

		r = append(r, &Node{isBlackHole, i, wb})
	}

	return r, nil
}
