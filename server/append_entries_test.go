package server_test

import (
	"github.com/nfisher/nraft/server"
	"github.com/nfisher/nraft/state"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_heartbeat_should_respond_success_if_follower_is_synchronised(t *testing.T) {
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

	appendEntries := server.AppendEntriesRequest{
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

func Test_heartbeat_should_respond_failure_if_log_term_differs(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
			Log: []state.Command{
				{Term: 2},
			},
		},
	}

	appendEntries := server.AppendEntriesRequest{
		Term:         3,
		LeaderID:     [16]byte{01},
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		LeaderCommit: 1,
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsFalse(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(2)
	assert.Int(len(follower.Persistent.Log)).EqualTo(0)
}

func Test_heartbeat_should_respond_failure_if_log_shorter_than_request(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 1,
		},
	}

	appendEntries := server.AppendEntriesRequest{
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

func Test_heartbeat_should_respond_failure_if_term_less_than_receiver(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
		},
	}

	appendEntries := server.AppendEntriesRequest{
		Term:     1,
		LeaderID: [16]byte{01},
	}

	followerResp := callAppendEntries(follower, appendEntries)

	assert.IsFalse(followerResp.Success)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_get_append_entries_should_fail(t *testing.T) {
	assert := Assert{t}
	follower := &server.Raft{}

	ts := httptest.NewUnstartedServer(server.Mux(follower))
	ts.Start()
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/append_entries")
	assert.NilError(err)
	assert.Int(resp.StatusCode).EqualTo(http.StatusMethodNotAllowed)
}

func callAppendEntries(follower *server.Raft, requestVote server.AppendEntriesRequest) *server.AppendEntriesResponse {
	var followerResp server.AppendEntriesResponse

	ts := httptest.NewUnstartedServer(server.Mux(follower))
	ts.Start()
	defer ts.Close()

	buf, err := server.Encode(&requestVote)
	if err != nil {
		panic(err)
	}

	client := ts.Client()

	resp, err2 := client.Post(ts.URL+"/append_entries", "application/msgpack", buf)
	if err2 != nil {
		panic(err2)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}

	err3 := server.Decode(resp.Body, &followerResp)
	if err3 != nil && err3 != io.EOF {
		panic(err3)
	}
	err3 = resp.Body.Close()
	if err3 != nil {
		panic(err3)
	}
	return &followerResp
}
