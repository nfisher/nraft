package server

import (
	"github.com/nfisher/nraft/state"
	"net/http"
)

type RequestVoteRequest struct {
	// CandidateID is candidate requesting vote.
	CandidateID []byte `json:"candidate_id"`
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
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestVote RequestVoteRequest

		defer req.Body.Close()
		err := Decode(req.Body, &requestVote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var voteResponse = &RequestVoteResponse{}

		r.VoteRequest(requestVote, voteResponse)

		err = EncodeTo(w, voteResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
