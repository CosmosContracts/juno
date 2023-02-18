package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoteRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       VoteRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: Valid request",
			query: VoteRequest{
				ProposalID: 1,
				Voter:      "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
		},
		{
			name: "Fail: Missing proposal id",
			query: VoteRequest{
				Voter: "cosmos14450hpujwlct9x0la3wv46sgk79czrl9phh0dm",
			},
			errExpected: true,
		},
		{
			name: "Fail: Missing voter",
			query: VoteRequest{
				ProposalID: 1,
			},
			errExpected: true,
		},
		{
			name:        "Fail: Empty request",
			query:       VoteRequest{},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
