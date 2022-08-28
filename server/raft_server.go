package server

import (
	"github.com/nfisher/nraft/state"
	"net/http"
)

type Raft struct {
	Volatile   state.Volatile
	Persistent state.Persistent
	Leader     state.Leader
	Server     http.Server
	Addr       string
}

func (r *Raft) LastTerm() state.Term {
	i := r.Volatile.CommitIndex
	if r.Volatile.CommitIndex < 1 {
		return 0
	}

	return r.Persistent.Log[i-1].Term
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

func (r *Raft) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/request_vote", requestVote(r))
	return mux
}
