package server

import (
	"github.com/nfisher/nraft/state"
	"net/http"
)

type AppendEntriesRequest struct {
	Term         state.Term      `json:"term"`
	LeaderID     [16]byte        `json:"leader_id"`
	PrevLogIndex int             `json:"prev_log_index"`
	PrevLogTerm  state.Term      `json:"prev_log_term"`
	Entries      []state.Command `json:"entries"`
	LeaderCommit int             `json:"leader_commit"`
}

type AppendEntriesResponse struct {
	Term    state.Term `json:"term"`
	Success bool       `json:"success"`
}

func appendEntries(r *Raft) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var appendEntries AppendEntriesRequest

		defer req.Body.Close()
		err := Decode(req.Body, &appendEntries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var appendResponse = &AppendEntriesResponse{}

		r.AppendEntries(appendEntries, appendResponse)

		err = EncodeTo(w, appendResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
