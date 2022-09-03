package server

import (
	"github.com/nfisher/nraft/state"
)

type Raft struct {
	Volatile   state.Volatile
	Persistent state.Persistent
	Leader     state.Leader
}

func (r *Raft) HasVoted() bool {
	return r.Persistent.VotedFor != nil
}

func (r *Raft) HasVotedFor(candidateID [16]byte) bool {
	if len(r.Persistent.VotedFor) != len(candidateID) {
		return false
	}

	for i := range r.Persistent.VotedFor {
		if r.Persistent.VotedFor[i] != candidateID[i] {
			return false
		}
	}

	return true
}

func (r *Raft) VoteRequest(requestVote RequestVoteRequest, voteResponse *RequestVoteResponse) {
	r.Persistent.RLock()
	defer r.Persistent.RUnlock()

	r.Volatile.RLock()
	defer r.Volatile.RUnlock()

	voteResponse.Term = r.Persistent.CurrentTerm
	voteResponse.VoteGranted = true

	if r.Persistent.CurrentTerm > requestVote.Term {
		voteResponse.VoteGranted = false
	}

	if r.HasVoted() && !r.HasVotedFor(requestVote.CandidateID) {
		voteResponse.VoteGranted = false
	}

	if r.Volatile.CommitIndex > requestVote.LastLogIndex {
		voteResponse.VoteGranted = false
	}
}

func (r *Raft) AppendEntries(appendEntries AppendEntriesRequest, appendResponse *AppendEntriesResponse) {
	r.Persistent.RLock()
	defer r.Persistent.RUnlock()

	r.Volatile.RLock()
	defer r.Volatile.RUnlock()

	appendResponse.Success = true
	appendResponse.Term = r.Persistent.CurrentTerm

	if r.Persistent.CurrentTerm > appendEntries.Term {
		appendResponse.Success = false
	}

	if len(r.Persistent.Log) < appendEntries.PrevLogIndex {
		appendResponse.Success = false
	} else if appendEntries.PrevLogIndex > 0 && r.Persistent.Log[appendEntries.PrevLogIndex-1].Term != appendEntries.PrevLogTerm {
		appendResponse.Success = false
		r.Persistent.Log = r.Persistent.Log[:appendEntries.PrevLogIndex-1]
	}
}
