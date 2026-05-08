package types

// TokenFactoryQuery represents possible queries to the x/tokenfactory module using wasmbindings
// DO NOT USE
type TokenFactoryQuery struct {
	// Given a subdenom minted by a contract via `OsmosisMsg::MintTokens`,
	// returns the full denom as used by `BankMsg::Send`.
	FullDenom       *FullDenom       `json:"full_denom,omitempty"`
	Admin           *DenomAdmin      `json:"admin,omitempty"`
	Metadata        *GetMetadata     `json:"metadata,omitempty"`
	DenomsByCreator *DenomsByCreator `json:"denoms_by_creator,omitempty"`
	Params          *GetParams       `json:"params,omitempty"`

	// x/voting-snapshot — historical staking-power queries for DAO
	// proposal-vote tallying. See planning/05-staking-snapshot.md.
	VotingPowerAt        *VotingPowerAt        `json:"voting_power_at,omitempty"`
	TotalVotingPowerAt   *TotalVotingPowerAt   `json:"total_voting_power_at,omitempty"`
	VotingPowerOverRange *VotingPowerOverRange `json:"voting_power_over_range,omitempty"`
}

// query types

type FullDenom struct {
	CreatorAddr string `json:"creator_addr"`
	Subdenom    string `json:"subdenom"`
}

type GetMetadata struct {
	Denom string `json:"denom"`
}

type DenomAdmin struct {
	Denom string `json:"denom"`
}

type DenomsByCreator struct {
	Creator string `json:"creator"`
}

type GetParams struct{}

// responses

type FullDenomResponse struct {
	Denom string `json:"denom"`
}

type AdminResponse struct {
	Admin string `json:"admin"`
}

type MetadataResponse struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

type DenomsByCreatorResponse struct {
	Denoms []string `json:"denoms"`
}

type ParamsResponse struct {
	Params Params `json:"params"`
}

// x/voting-snapshot

type VotingPowerAt struct {
	Address string `json:"address"`
	Height  int64  `json:"height"`
}

type TotalVotingPowerAt struct {
	Height int64 `json:"height"`
}

type VotingPowerResponse struct {
	// Power is the bonded stake amount as a base-10 string (uint).
	// Caller compares as Uint128 in CosmWasm contracts.
	Power string `json:"power"`
}

type VotingPowerOverRange struct {
	Address    string `json:"address"`
	FromHeight int64  `json:"from_height"`
	ToHeight   int64  `json:"to_height"`
}

type VotingPowerOverRangeResponse struct {
	Rows []HeightPowerPair `json:"rows"`
}

type HeightPowerPair struct {
	Height int64  `json:"height"`
	Power  string `json:"power"`
}
