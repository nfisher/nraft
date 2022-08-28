package server_test

import (
	"bytes"
	"encoding/json"
	"github.com/nfisher/nraft/server"
	"github.com/nfisher/nraft/state"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_should_respond_success_if_follower_is_synchronised(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
			Log: []state.Command{
				{Term: 1},
				{Term: 2},
			},
		},
		Volatile: state.Volatile{
			CommitIndex: 2,
		},
	}

	appendEntries := server.AppendEntries{
		Term:         2,
		LeaderID:     [16]byte{01},
		PrevLogIndex: 2,
		PrevLogTerm:  2,
		LeaderCommit: 2,
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsTrue(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_should_respond_failure_if_log_term_differs(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
			Log: []state.Command{
				{Term: 2},
			},
		},
	}

	appendEntries := server.AppendEntries{
		Term:         3,
		LeaderID:     [16]byte{01},
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		LeaderCommit: 1,
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsFalse(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_should_respond_failure_if_log_shorter_than_request(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 1,
		},
	}

	appendEntries := server.AppendEntries{
		Term:         2,
		LeaderID:     [16]byte{01},
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		LeaderCommit: 1,
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsFalse(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(1)
}

func Test_should_respond_failure_if_term_less_than_receiver(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
		},
	}

	appendEntries := server.AppendEntries{
		Term:     1,
		LeaderID: [16]byte{01},
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsFalse(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(2)
}

func callAppendEntries(follower *server.Raft, requestVote server.AppendEntries) *server.AppendEntriesResponse {
	var followerResp server.AppendEntriesResponse

	ts := httptest.NewUnstartedServer(follower.Mux())
	ts.Start()
	defer ts.Close()

	buf, err := json.Marshal(&requestVote)
	if err != nil {
		panic(err)
	}

	client := ts.Client()

	resp, err2 := client.Post(ts.URL+"/append_entries", "application/json", bytes.NewBuffer(buf))
	if err2 != nil {
		panic(err2)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}

	err3 := json.NewDecoder(resp.Body).Decode(&followerResp)
	if err3 != nil && err3 != io.EOF {
		panic(err3)
	}
	err3 = resp.Body.Close()
	if err3 != nil {
		panic(err3)
	}
	return &followerResp
}
