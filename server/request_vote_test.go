package server_test

import (
	"github.com/nfisher/nraft/server"
	"github.com/nfisher/nraft/state"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func expand(b []byte) []byte {
	var b2 = make([]byte, 16, 16)
	for i, v := range b {
		if i >= len(b2) {
			break
		}
		b2[i] = v
	}
	return b2
}

func Test_should_accept_candidate_if_voted_for_them_previously(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
			VotedFor:    expand([]byte{01}),
		},
	}

	requestVote := server.RequestVoteRequest{
		Term:        2,
		CandidateID: expand([]byte{01}),
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsTrue(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_should_accept_candidate_if_not_voted(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 1,
		},
	}

	requestVote := server.RequestVoteRequest{
		Term:        2,
		CandidateID: expand([]byte{01}),
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsTrue(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(1)
}

func Test_should_reject_candidate_if_id_is_empty(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 1,
		},
	}

	requestVote := server.RequestVoteRequest{
		Term: 2,
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsFalse(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(1)
}

func Test_should_reject_candidate_if_term_less_than_receiver(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
		},
	}

	requestVote := server.RequestVoteRequest{
		Term:        1,
		CandidateID: expand([]byte{01}),
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsFalse(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_should_reject_candidate_if_voted_for_other_candidate(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 1,
			VotedFor:    expand([]byte{02}),
		},
		Volatile: state.Volatile{},
	}

	requestVote := server.RequestVoteRequest{
		Term:        1,
		CandidateID: expand([]byte{01}),
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsFalse(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(1)
}

func Test_should_reject_candidate_if_log_index_is_behind(t *testing.T) {
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

	requestVote := server.RequestVoteRequest{
		Term:         3,
		CandidateID:  expand([]byte{01}),
		LastLogTerm:  1,
		LastLogIndex: 1,
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsFalse(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(2)
}

func Test_get_request_vote_should_fail(t *testing.T) {
	assert := Assert{t}
	follower := &server.Raft{}

	ts := httptest.NewUnstartedServer(server.Mux(follower))
	ts.Start()
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/request_vote")
	assert.NilError(err)
	assert.Int(resp.StatusCode).EqualTo(http.StatusMethodNotAllowed)
}

func callRequestVote(follower *server.Raft, requestVote server.RequestVoteRequest) *server.RequestVoteResponse {
	var followerResp server.RequestVoteResponse

	ts := httptest.NewUnstartedServer(server.Mux(follower))
	ts.Start()
	defer ts.Close()

	buf, err := server.Encode(&requestVote)
	if err != nil {
		panic(err)
	}

	client := ts.Client()

	resp, err2 := client.Post(ts.URL+"/request_vote", "application/msgpack", buf)
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
