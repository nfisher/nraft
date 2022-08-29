package server

import (
	"encoding/json"
	"github.com/nfisher/nraft/state"
	"net/http"
)

type AppendEntries struct {
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
		r.Persistent.RLock()
		defer r.Persistent.RUnlock()

		r.Volatile.RLock()
		defer r.Volatile.RUnlock()

		var appendEntries AppendEntries

		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(&appendEntries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var appendResponse = AppendEntriesResponse{
			Success: true,
			Term:    r.Persistent.CurrentTerm,
		}

		if r.Persistent.CurrentTerm > appendEntries.Term {
			appendResponse.Success = false
		}

		if len(r.Persistent.Log) < appendEntries.PrevLogIndex {
			appendResponse.Success = false
		} else if appendEntries.PrevLogIndex > 0 && r.Persistent.Log[appendEntries.PrevLogIndex-1].Term != appendEntries.PrevLogTerm {
			appendResponse.Success = false
			r.Persistent.Log = r.Persistent.Log[:appendEntries.PrevLogIndex-1]
		}

		err = json.NewEncoder(w).Encode(&appendResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
