package utils

import "fmt"

// Agent is an abstraction of agents that move around the ring
type Agent struct {
	Pos    *Node
	R      Ring
	Active bool
}

// Right moves the agent to the right, unless it is not active
func (a *Agent) Right() error {
	if !a.Active {
		return fmt.Errorf("non-active agent can't move")
	}

	currIndex := a.Pos.ID
	newIndex := (currIndex - 1) % len(a.R)
	a.Pos = a.R[newIndex]

	if a.Pos.Bh {
		a.Active = false
	}

	return nil
}

// Left moves the agent to the left, unless it is not active
func (a *Agent) Left() error {
	if !a.Active {
		return fmt.Errorf("non-active agent can't move")
	}

	currIndex := a.Pos.ID
	newIndex := (currIndex + 1) % len(a.R)
	a.Pos = a.R[newIndex]

	if a.Pos.Bh {
		a.Active = false
	}

	return nil
}

// ReadWb reads the data found in the whiteboard of the current node at a specified key
func (a *Agent) ReadWb(key interface{}) (interface{}, error) {
	if a.Pos.wb == nil {
		return nil, fmt.Errorf("Node has no whiteboard to read from")
	}

	val, _ := a.Pos.wb.Load(key)
	return val, nil
}

// WriteWb writes the data given to the given key in the current node's whiteboard
func (a *Agent) WriteWb(key, val interface{}) error {
	if a.Pos.wb == nil {
		return fmt.Errorf("Node has no whiteboard to write on")
	}

	a.Pos.wb.Store(key, val)
	return nil
}
