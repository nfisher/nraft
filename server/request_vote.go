package server

import (
	"encoding/json"
	"github.com/nfisher/nraft/state"
	"net/http"
)

type RequestVote struct {
	// CandidateID is candidate requesting vote.
	CandidateID [16]byte `json:"candidate_id"`
	// LastLogIndex is index of candidate’s last log entry.
	LastLogIndex int `json:"last_log_index"`
	// LastLogTerm is term of candidate’s last log entry.
	LastLogTerm state.Term `json:"last_log_term"`
	// Term is candidate’s term.
	Term state.Term `json:"term"`
}

type RequestVoteResponse struct {
	// Term is currentTerm, for candidate to update itself.
	Term state.Term `json:"term"`
	// VoteGranted if true means candidate received vote.
	VoteGranted bool `json:"vote_granted"`
}

func requestVote(r *Raft) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.Persistent.RLock()
		defer r.Persistent.RUnlock()

		r.Volatile.RLock()
		defer r.Volatile.RUnlock()

		var requestVote RequestVote

		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(&requestVote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var voteResponse = RequestVoteResponse{
			Term:        r.Persistent.CurrentTerm,
			VoteGranted: true,
		}

		if r.Persistent.CurrentTerm > requestVote.Term {
			voteResponse.VoteGranted = false
		}

		if r.HasVoted() && !r.HasVotedFor(requestVote.CandidateID) {
			voteResponse.VoteGranted = false
		}

		if r.Volatile.CommitIndex > requestVote.LastLogIndex {
			voteResponse.VoteGranted = false
		}

		err = json.NewEncoder(w).Encode(&voteResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
