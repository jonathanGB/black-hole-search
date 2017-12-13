package helpers

import "../bhs"

// Min returns the minimum of two NodeIDs
func Min(a bhs.NodeID, b bhs.NodeID) bhs.NodeID {
	if a < b {
		return a
	}
	return b
}

// Max returns the minimum of two NodeIDs
func Max(a, b bhs.NodeID) bhs.NodeID {
	if a < b {
		return b
	}
	return a
}

// MinUint64 returns the minimum of two uint64s
func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// MaxUint64 returns the minimum of two NodeIDs
func MaxUint64(a, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}
