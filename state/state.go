package state

import "sync"

type Term uint64

// Persistent is the state that is locally persisted for all servers.
type Persistent struct {
	// CurrentTerm is the latest term this server has seen.
	CurrentTerm Term
	// VotedFor is the candidateID that received a vote in the current term or null if none.
	VotedFor []byte
	// Log contains a command to be applied to the state machine and the term when the entry was received by the leader (first index is 1).
	Log []Command

	sync.RWMutex
}

// Volatile is volatile in memory state for all servers.
type Volatile struct {
	// CommitIndex is the index of the highest log entry known to be committed.
	CommitIndex int
	// LastApplied is the index of the highest log entry applied to state machine.
	LastApplied int
	sync.RWMutex
}

// Leader is volatile in memory state for leader.
type Leader struct {
	// NextIndex is the index of the next log entry to send to each server.
	NextIndex []int
	// MatchIndex index of the highest log entry known to be replicated on each server.
	MatchIndex []int
	sync.RWMutex
}

type Command struct {
	Term Term
}
