package server_test

import (
	"github.com/nfisher/nraft/server"
	"github.com/nfisher/nraft/state"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_should_accept_candidate_if_voted_for_them_previously(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
			VotedFor: []byte{
				01, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
		},
	}

	requestVote := server.RequestVote{
		Term:        2,
		CandidateID: [16]byte{01},
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

	requestVote := server.RequestVote{
		Term:        2,
		CandidateID: [16]byte{01},
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsTrue(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(1)
}

func Test_should_reject_candidate_if_term_less_than_receiver(t *testing.T) {
	assert := Assert{t}

	follower := &server.Raft{
		Persistent: state.Persistent{
			CurrentTerm: 2,
		},
	}

	requestVote := server.RequestVote{
		Term:        1,
		CandidateID: [16]byte{01},
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
			VotedFor: []byte{
				02, 00, 00, 00,
				00, 00, 00, 00,
				00, 00, 00, 00,
				00, 00, 00, 00,
			},
		},
		Volatile: state.Volatile{},
	}

	requestVote := server.RequestVote{
		Term:        1,
		CandidateID: [16]byte{01},
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

	requestVote := server.RequestVote{
		Term:         3,
		CandidateID:  [16]byte{01},
		LastLogTerm:  1,
		LastLogIndex: 1,
	}

	followerResp := callRequestVote(follower, requestVote)

	assert.IsFalse(followerResp.VoteGranted)
	assert.Term(followerResp.Term).EqualTo(2)
}

func callRequestVote(follower *server.Raft, requestVote server.RequestVote) *server.RequestVoteResponse {
	var followerResp server.RequestVoteResponse

	ts := httptest.NewUnstartedServer(follower.Mux())
	ts.Start()
	defer ts.Close()

	buf, err := encode(&requestVote)
	if err != nil {
		panic(err)
	}

	client := ts.Client()

	resp, err2 := client.Post(ts.URL+"/request_vote", "application/json", buf)
	if err2 != nil {
		panic(err2)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}

	err3 := decode(resp.Body, &followerResp)
	if err3 != nil && err3 != io.EOF {
		panic(err3)
	}
	err3 = resp.Body.Close()
	if err3 != nil {
		panic(err3)
	}
	return &followerResp
}
