package bhs

import "fmt"

// Agent is an abstraction of agents that move around the ring
type Agent struct {
	Direction      Direction
	Position       *Node
	Ring           Ring
	Active         bool
	Moves          uint64
	cautiousWalk   bool
	UnexploredSet  [2]NodeID
	ActAsSmall     bool
	HomebaseNodeID NodeID
}

// NewAgent helps construct an agent
func NewAgent(direction Direction, ring Ring, cautiousWalk bool) *Agent {
	unexploredSet := [2]NodeID{1, NodeID(len(ring) - 1)}
	homebaseNodeID := NodeID(0)
	return &Agent{direction, ring[homebaseNodeID], ring, true, 0, cautiousWalk, unexploredSet, true, homebaseNodeID}
}

// Move combines logic for moving left and right
func (agent *Agent) Move(direction Direction) (updateFound bool, err error) {
	if !agent.Active {
		return false, fmt.Errorf("non-active agent can't move")
	}

	oppositeDirection := GetOppositeDirection(direction)
	var outgoingEdgeLabel ExploredType
	var sourceNodeWhiteboard *Whiteboard

	// cautious walk: mark edge as active before leaving, immediately come back to mark as explored if safe
	if agent.cautiousWalk {
		sourceNodeWhiteboard = agent.Position.whiteboard
		sourceNodeWhiteboard.Lock()
		if agent.checkForUpdate() { // always check for an update before moving
			sourceNodeWhiteboard.Unlock()
			return true, nil
		}
		outgoingEdgeLabel = sourceNodeWhiteboard.label[direction]
		switch outgoingEdgeLabel {
		case unexplored:
			sourceNodeWhiteboard.label[direction] = active
		case active:
			sourceNodeWhiteboard.Unlock()
			return false, fmt.Errorf("cannot cross an active link")
		}
		sourceNodeWhiteboard.Unlock()
	}

	newIndex := agent.getNewIndex(direction)
	agent.Position = agent.Ring[newIndex]

	if agent.Position.BlackHole {
		agent.Active = false
		return false, fmt.Errorf("reached a black hole")
	}

	agent.Moves++
	if !agent.cautiousWalk {
		return
	}

	// Arrived at destination, mark incoming edge label as explored
	destinationSourceWhiteboard := agent.Position.whiteboard
	destinationSourceWhiteboard.Lock()
	destinationSourceWhiteboard.label[oppositeDirection] = explored
	destinationSourceWhiteboard.Unlock()

	// Update agent's unexplored set with node just visited
	switch agent.Position.ID {
	case agent.UnexploredSet[1]:
		agent.UnexploredSet[1]-- // if the agent is located at the rightmost unexplored node, decrement the index of the rightmost unexplored node
	case agent.UnexploredSet[0]:
		agent.UnexploredSet[0]++ // if the agent is located at the leftmost unexplored node, increment the index of the leftmost unexplored node
	}

	if outgoingEdgeLabel != unexplored { // Stop here unless agent needs to go back to mark outgoing label as explored
		return false, nil
	}

	// go back to source to mark its outgoing edge label as explored and check for new instructions
	if updateFound, err := agent.Move(oppositeDirection); err != nil || updateFound {
		return updateFound, err
	}

	// otherwise, keep doing your thing
	if updateFound, err := agent.Move(direction); err != nil || updateFound {
		return updateFound, err
	}

	return false, nil // successful, nothing to declare
}

// MoveUntil moves agent to the direction specified until it reaches a given index
// Returns true if made it alive to the destination, otherwise false
func (agent *Agent) MoveUntil(direction Direction, id NodeID) (successul bool, updateFound bool) {
	for agent.Position.ID != id {
		updateFound, err := agent.Move(direction)
		if err != nil || updateFound {
			return err == nil, updateFound
		}
	}
	return true, false
}

// MoveToLastExplored is used for cautious walk
func (agent *Agent) MoveToLastExplored(direction Direction) bool {
	agent.Position.whiteboard.Lock()
	for agent.Position.whiteboard.label[direction] == explored {
		agent.Position.whiteboard.Unlock()
		if _, err := agent.Move(direction); err != nil {
			return false
		}
		agent.Position.whiteboard.Lock()
	}
	return true
}

// Small is used for optTeamSize
func (agent *Agent) Small(isLeftAgent bool, remainingIterationsAsSmall uint8, blackHole chan<- NodeID) {
	var nodeToExplore, err = exploreUpTo(isLeftAgent, agent.UnexploredSet)
	if err != nil {

	}
	if isLeftAgent {
		if ok, _ := agent.MoveUntil(Left, nodeToExplore); !ok { // visit one node
			return
		}
		agent.MoveUntil(Right, agent.HomebaseNodeID)
		if agent.UnexploredSet[0] == agent.UnexploredSet[1] {
			blackHole <- agent.UnexploredSet[1]
			return
		}
	} else {
		if ok, _ := agent.MoveUntil(Right, agent.UnexploredSet[1]); !ok {
			return
		}
		agent.MoveUntil(Left, agent.HomebaseNodeID)
		if agent.UnexploredSet[0] == agent.UnexploredSet[1] {
			blackHole <- agent.UnexploredSet[1]
			return
		}
	}

	agent.LeaveUpdate(isLeftAgent)

	remainingIterationsAsSmall--
	switch remainingIterationsAsSmall {
	case 1:
		agent.Small(isLeftAgent, remainingIterationsAsSmall, blackHole)
	case 0:
		agent.Big(isLeftAgent, blackHole)
	}
	return
}

// Big is used for optTeamSize
func (agent *Agent) Big(isLeftAgent bool, blackHole chan<- NodeID) {
	const cautiousWalk = true
	if isLeftAgent {
		ok, updateFound := agent.MoveUntil(Left, agent.UnexploredSet[1]-1) // visit all but one unexplored nodes
		if !ok {
			return
		}
		if !updateFound {
			agent.MoveUntil(Right, agent.HomebaseNodeID) // return home
		}
		agent.Small(isLeftAgent, 2, blackHole)
	} else {
		ok, updateFound := agent.MoveUntil(Right, agent.UnexploredSet[0]+1) // visit all but one unexplored nodes
		if !ok {
			return
		}
		if !updateFound {
			agent.MoveUntil(Left, agent.HomebaseNodeID) // return home
		}
		agent.Small(isLeftAgent, 2, blackHole)
	}
	blackHole <- agent.UnexploredSet[1]
	return
}

// LeaveUpdate is used for optTeamSize
func (agent *Agent) LeaveUpdate(isLeftAgent bool) {
	var direction Direction
	if isLeftAgent {
		direction = Right
	} else {
		direction = Left
	}
	agent.MoveToLastExplored(direction)

	whiteboard := agent.Position.whiteboard
	// whiteboard.Lock() ALREADY LOCKED FROM PREVIOUS METHOD CALL

	whiteboard.updateForAgent = direction
	whiteboard.actAsSmall = !agent.ActAsSmall
	// getting the halfway point of the unexplored set, then finding the node halfway around the ring from it should be the center of the explored set
	// cannot do negative modulo, because NodeID is an unsigned integer
	ringSize := NodeID(len(agent.Ring) / 2)
	middleOfUnexploredSetNodeID := (agent.UnexploredSet[1] - agent.UnexploredSet[0]) / 2
	whiteboard.homebaseNodeID = (ringSize + middleOfUnexploredSetNodeID) % ringSize
	whiteboard.unexploredSet = agent.UnexploredSet

	whiteboard.Unlock()
}

// LeaveUpdateDivide is used for updating the other agent during the divide algorithm
func (agent *Agent) LeaveUpdateDivide() {
	oppositeDirection := GetOppositeDirection(agent.Direction)
	agent.MoveToLastExplored(oppositeDirection)

	whiteboard := agent.Position.whiteboard
	// whiteboard.Lock() ALREADY LOCKED FROM PREVIOUS METHOD CALL

	whiteboard.updateForAgent = oppositeDirection
	whiteboard.unexploredSet = agent.UnexploredSet

	whiteboard.Unlock()
}

func exploreUpTo(isLeftAgent bool, unexploredSet [2]NodeID) (nodeIndex NodeID, err error) {
	if unexploredSet[0] == unexploredSet[1] {
		return unexploredSet[1], fmt.Errorf("only one node left to explore")
	}

	if isLeftAgent {
		return unexploredSet[0], nil
	}
	return unexploredSet[1], nil
}

func (agent *Agent) getNewIndex(direction Direction) NodeID {
	newID := agent.Position.ID
	if direction == Right {
		newID += NodeID(len(agent.Ring)) - 1 // go around the ring to previous node
	} else if direction == Left {
		newID++ // go to next node
	}
	newID = newID % NodeID(len(agent.Ring))
	return newID
}

func (agent *Agent) checkForUpdate() bool {
	if agent.ActAsSmall { // only big agents check for updates
		return false
	}

	whiteboard := agent.Position.whiteboard

	if whiteboard.unexploredSet == [2]NodeID{} || agent.Direction != whiteboard.updateForAgent {
		return false
	}

	// store updates
	agent.ActAsSmall = whiteboard.actAsSmall
	agent.UnexploredSet = whiteboard.unexploredSet
	agent.HomebaseNodeID = whiteboard.homebaseNodeID

	// erase unexplored set as an indicator that update was read
	whiteboard.unexploredSet = [2]NodeID{}
	whiteboard.updateForAgent = None

	return true
}

// GetOppositeDirection is self-explanatory
func GetOppositeDirection(direction Direction) (oppositeDirection Direction) {
	return (direction + 1) % 2
}
