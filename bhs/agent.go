package bhs

import "fmt"

// Agent is an abstraction of agents that move around the ring
type agent struct {
	Pos    *Node
	R      Ring
	Active bool
}

// NewAgent helps construct an agent
func NewAgent(r Ring) *agent {
	return &agent{r[HOMEBASE], r, true}
}

// Right moves the agent to the right, unless it is not active
func (a *agent) Right() error {
	if !a.Active {
		return fmt.Errorf("non-active agent can't move")
	}

	currIndex := a.Pos.ID
	newIndex := (((currIndex - 1) % len(a.R)) + len(a.R)) % len(a.R) // compute a negative modulo
	a.Pos = a.R[newIndex]

	if a.Pos.Bh {
		a.Active = false
		return fmt.Errorf("reached a black hole")
	}

	return nil
}

// RightUntil moves agent to the right until it reaches a given index
// Returns true if made it alive to the destination, otherwise false
func (a *agent) RightUntil(i int) bool {
	for a.Pos.ID != i {
		if err := a.Right(); err != nil {
			return false
		}
	}

	return true
}

// Left moves the agent to the left, unless it is not active
func (a *agent) Left() error {
	if !a.Active {
		return fmt.Errorf("non-active agent can't move")
	}

	currIndex := a.Pos.ID
	newIndex := (currIndex + 1) % len(a.R)
	a.Pos = a.R[newIndex]

	if a.Pos.Bh {
		a.Active = false
		return fmt.Errorf("reached a black hole")
	}

	return nil
}

// LeftUntil moves agent to the left until it reaches a given index
// Returns true if made it alive to the destination, otherwise false
func (a *agent) LeftUntil(i int) bool {
	for a.Pos.ID != i {
		if err := a.Left(); err != nil {
			return false
		}
	}

	return true
}

// ReadWb reads the data found in the whiteboard of the current node at a specified key
func (a *agent) ReadWb(key interface{}) (interface{}, error) {
	if a.Pos.wb == nil {
		return nil, fmt.Errorf("Node has no whiteboard to read from")
	}

	val, _ := a.Pos.wb.Load(key)
	return val, nil
}

// WriteWb writes the data given to the given key in the current node's whiteboard
func (a *agent) WriteWb(key, val interface{}) error {
	if a.Pos.wb == nil {
		return fmt.Errorf("Node has no whiteboard to write on")
	}

	a.Pos.wb.Store(key, val)
	return nil
}
