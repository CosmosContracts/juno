package types

import (
	"fmt"

	govTypes "github.com/CosmosContracts/juno/v13/wasmbinding/gov/types"
)

// Query is a container for custom WASM queries (one of).
type Query struct {
	// GovVote returns the vote data for a given proposal and voter.
	GovVote *govTypes.VoteRequest `json:"gov_vote"`
}

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.GovVote != nil {
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one sub-query must be set (fields=%v)", cnt)
	}

	return nil
}
