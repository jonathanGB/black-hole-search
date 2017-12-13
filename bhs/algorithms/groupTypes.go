package algorithms

import "../../bhs"

type groupChannelResult struct {
	direction    bhs.Direction
	visitedRange [2]bhs.NodeID
}

type groupChannelResponse struct {
	success    bool
	result     groupChannelResult
	moves      uint64
	group      AgentGroup
	groupIndex uint64
}

// AgentGroup ...
type AgentGroup uint8

// Groups
const (
	LeftGroup AgentGroup = iota
	RightGroup
	MiddleGroup
	TieBreakerGroup
)
