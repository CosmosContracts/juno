// Package daodao_test exercises released DAO DAO contracts against the
// candidate Juno image. Contract bytes are checked in and verified before use;
// the test never downloads artifacts at runtime.
package daodao_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	votingsnapshottypes "github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

const (
	contractsDir = "../../contracts"
	proposalID   = uint64(1)
)

var cw4Artifacts = map[string]string{
	"cw4_group.wasm":           "dd2216f1114fc68bc4c043701b02e55ce3e5598cdeb616985388215a400db277",
	"dao_voting_cw4.wasm":      "d0e6bac4d7c1861f36328e7c0367f863999f999e2ae21df612e301eea5fe90d8",
	"dao_proposal_single.wasm": "e38fc5bb1b5e74ef154340567c673492515498b2120e5f15b0c990cd9fa5fe6a",
	"dao_dao_core.wasm":        "5d078fc9aec04df18c335446eb8df03d24c73ee745f76fd39624d4c5fa768b4c",
	"voting_power_probe.wasm":  "a074bc275b8b38eebd79db1603645ae71759a679385da2ccd82622f39f1f57af",
}

type DaoDaoTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestDaoDaoTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		if s.Ic != nil {
			_ = s.Ic.Close()
		}
	})

	testSuite := &DaoDaoTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestCw4GroupDao stores authentic release artifacts, instantiates a DAO whose
// voting module creates its cw4 group, discovers all child contracts, and
// asserts the proposal's Open -> Passed -> Executed lifecycle.
func (s *DaoDaoTestSuite) TestCw4GroupDao() {
	t := s.T()
	require := s.Require()

	require.NoError(verifyCw4Artifacts(contractsDir), "checked-in release artifacts must match documented SHA-256 sums")

	user := s.GetAndFundTestUser(t.Name(), 10_000_000_000, s.Chain)
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))

	cw4GroupCodeID := s.StoreContract(s.Chain, user.KeyName(), filepath.Join(contractsDir, "cw4_group.wasm"), fees)
	votingCodeID := s.StoreContract(s.Chain, user.KeyName(), filepath.Join(contractsDir, "dao_voting_cw4.wasm"), fees)
	proposalCodeID := s.StoreContract(s.Chain, user.KeyName(), filepath.Join(contractsDir, "dao_proposal_single.wasm"), fees)
	coreCodeID := s.StoreContract(s.Chain, user.KeyName(), filepath.Join(contractsDir, "dao_dao_core.wasm"), fees)
	require.NotEmpty(cw4GroupCodeID)
	require.NotEmpty(votingCodeID)
	require.NotEmpty(proposalCodeID)
	require.NotEmpty(coreCodeID)

	daoMsg, err := buildDaoInstantiate(user.FormattedAddress(), votingCodeID, proposalCodeID, cw4GroupCodeID)
	require.NoError(err)
	dao, err := s.InstantiateContract(s.Chain, user.KeyName(), coreCodeID, daoMsg, fees, false, false)
	require.NoError(err)
	require.NotEmpty(dao, "core must be instantiated")

	voting, err := queryVotingModule(s.Ctx, s.Chain, dao)
	require.NoError(err)
	require.NotEmpty(voting, "core must report its voting child")
	proposal, err := queryProposalModule(s.Ctx, s.Chain, dao)
	require.NoError(err)
	require.NotEmpty(proposal, "core must report its enabled proposal child")
	group, err := queryGroupContract(s.Ctx, s.Chain, voting)
	require.NoError(err)
	require.NotEmpty(group, "voting module must report its cw4 group child")

	power, err := queryVotingPower(s.Ctx, s.Chain, dao, user.FormattedAddress())
	require.NoError(err)
	require.Equal("1", power, "the sole cw4 member must have voting power")

	require.NoError(openProposal(s.Ctx, s.Chain, user.KeyName(), proposal, "cw4 lifecycle", "no-op proposal", fees))
	status, err := queryProposalStatus(s.Ctx, s.Chain, proposal, proposalID)
	require.NoError(err)
	require.Equal("open", status, "proposal must be open before voting")

	require.NoError(voteOnProposal(s.Ctx, s.Chain, user.KeyName(), proposal, proposalID, "yes", fees))
	status, err = queryProposalStatus(s.Ctx, s.Chain, proposal, proposalID)
	require.NoError(err)
	require.Equal("passed", status, "the sole member's yes vote must pass the proposal")

	require.NoError(executeProposal(s.Ctx, s.Chain, user.KeyName(), proposal, proposalID, fees))
	status, err = queryProposalStatus(s.Ctx, s.Chain, proposal, proposalID)
	require.NoError(err)
	require.Equal("executed", status, "executing the passed proposal must be persisted")
}

// The cw20-staked and custom-binding legs remain explicit follow-up scope; issue
// #14's release gate is the cw4 lifecycle above.
func (s *DaoDaoTestSuite) TestCw20StakedDao() {
	s.T().Skip("follow-up: cw20-staked is outside issue #14")
}

// TestWasmbindingsVotingPowerAt deploys a source-controlled probe contract,
// delegates real stake, and compares current, historical, total, and range
// custom-query responses against the module's gRPC API.
func (s *DaoDaoTestSuite) TestWasmbindingsVotingPowerAt() {
	t := s.T()
	require := s.Require()
	const stakeAmount = int64(1_000_000)

	require.NoError(verifyCw4Artifacts(contractsDir), "checked-in probe must match its source-controlled SHA-256")
	user := s.GetAndFundTestUser(t.Name(), 10_000_000_000, s.Chain)
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))

	probeCodeID := s.StoreContract(s.Chain, user.KeyName(), filepath.Join(contractsDir, "voting_power_probe.wasm"), fees)
	require.NotEmpty(probeCodeID)
	probe, err := s.InstantiateContract(s.Chain, user.KeyName(), probeCodeID, `{}`, fees, false, false)
	require.NoError(err)
	require.NotEmpty(probe)

	beforeHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(err)
	before, err := s.VotingSnapshotClient.VotingPowerAt(s.Ctx, &votingsnapshottypes.QueryVotingPowerAtRequest{
		Address: user.FormattedAddress(), AtHeight: beforeHeight,
	})
	require.NoError(err)
	require.Equal("0", before.Power)

	validators, err := s.StakingClient.Validators(s.Ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
	require.NoError(err)
	require.NotEmpty(validators.Validators)
	s.StakeTokens(
		s.Chain,
		user,
		validators.Validators[0].OperatorAddress,
		sdk.NewInt64Coin(s.Denom, stakeAmount).String(),
		fees,
		false,
	)

	afterHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(err)
	require.Greater(afterHeight, beforeHeight)

	directPower, err := s.VotingSnapshotClient.VotingPowerAt(s.Ctx, &votingsnapshottypes.QueryVotingPowerAtRequest{
		Address: user.FormattedAddress(), AtHeight: afterHeight,
	})
	require.NoError(err)
	require.Equal(fmt.Sprint(stakeAmount), directPower.Power)
	directTotal, err := s.VotingSnapshotClient.TotalVotingPowerAt(s.Ctx, &votingsnapshottypes.QueryTotalVotingPowerAtRequest{
		AtHeight: afterHeight,
	})
	require.NoError(err)
	directRange, err := s.VotingSnapshotClient.VotingPowerOverRange(s.Ctx, &votingsnapshottypes.QueryVotingPowerOverRangeRequest{
		Address: user.FormattedAddress(), FromHeight: beforeHeight, ToHeight: afterHeight,
	})
	require.NoError(err)
	require.NotEmpty(directRange.Rows)

	var historical votingPowerResponse
	require.NoError(s.Chain.QueryContract(s.Ctx, probe, map[string]any{
		"voting_power_at": map[string]any{"address": user.FormattedAddress(), "height": beforeHeight},
	}, &historical))
	require.Equal(before.Power, historical.Power)

	var current votingPowerResponse
	require.NoError(s.Chain.QueryContract(s.Ctx, probe, map[string]any{
		"voting_power_at": map[string]any{"address": user.FormattedAddress(), "height": afterHeight},
	}, &current))
	require.Equal(directPower.Power, current.Power)

	var total votingPowerResponse
	require.NoError(s.Chain.QueryContract(s.Ctx, probe, map[string]any{
		"total_voting_power_at": map[string]any{"height": afterHeight},
	}, &total))
	require.Equal(directTotal.Power, total.Power)

	var powerRange votingPowerRangeResponse
	require.NoError(s.Chain.QueryContract(s.Ctx, probe, map[string]any{
		"voting_power_over_range": map[string]any{
			"address": user.FormattedAddress(), "from_height": beforeHeight, "to_height": afterHeight,
		},
	}, &powerRange))
	require.Len(powerRange.Rows, len(directRange.Rows))
	for i := range directRange.Rows {
		require.Equal(directRange.Rows[i].Height, powerRange.Rows[i].Height)
		require.Equal(directRange.Rows[i].Power, powerRange.Rows[i].Power)
	}
}

type votingPowerResponse struct {
	Power string `json:"power"`
}

type votingPowerRangeResponse struct {
	Rows []struct {
		Height int64  `json:"height"`
		Power  string `json:"power"`
	} `json:"rows"`
}

type moduleInstantiateInfo struct {
	CodeID uint64          `json:"code_id"`
	Msg    json.RawMessage `json:"-"`
	Admin  map[string]any  `json:"admin"`
	Funds  any             `json:"funds"`
	Label  string          `json:"label"`
	Salt   any             `json:"salt"`
}

func (m moduleInstantiateInfo) MarshalJSON() ([]byte, error) {
	type wire struct {
		CodeID uint64         `json:"code_id"`
		Msg    []byte         `json:"msg"`
		Admin  map[string]any `json:"admin"`
		Funds  any            `json:"funds"`
		Label  string         `json:"label"`
		Salt   any            `json:"salt"`
	}
	return json.Marshal(wire{m.CodeID, []byte(m.Msg), m.Admin, m.Funds, m.Label, m.Salt})
}

func buildDaoInstantiate(member, votingCodeID, proposalCodeID, cw4CodeID string) (string, error) {
	votingID, err := strconv.ParseUint(votingCodeID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse voting code ID: %w", err)
	}
	proposalID, err := strconv.ParseUint(proposalCodeID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse proposal code ID: %w", err)
	}
	groupID, err := strconv.ParseUint(cw4CodeID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse cw4 code ID: %w", err)
	}

	votingMsg, err := json.Marshal(map[string]any{
		"group_contract": map[string]any{
			"new": map[string]any{
				"cw4_group_code_id": groupID,
				"cw4_group_salt":    nil,
				"initial_members": []map[string]any{{
					"addr": member, "weight": 1,
				}},
			},
		},
	})
	if err != nil {
		return "", err
	}
	proposalMsg, err := json.Marshal(map[string]any{
		"threshold":                           map[string]any{"absolute_percentage": map[string]any{"percentage": map[string]any{"majority": map[string]any{}}}},
		"max_voting_period":                   map[string]any{"height": 100},
		"min_voting_period":                   nil,
		"only_members_execute":                false,
		"allow_revoting":                      false,
		"pre_propose_info":                    map[string]any{"anyone_may_propose": map[string]any{}},
		"close_proposal_on_execution_failure": false,
		"veto":                                nil,
		"delegation_module":                   nil,
	})
	if err != nil {
		return "", err
	}

	coreMsg := map[string]any{
		"admin":                    nil,
		"name":                     "DAO DAO cw4 lifecycle",
		"description":              "Juno candidate image compatibility test",
		"image_url":                nil,
		"automatically_add_cw20s":  false,
		"automatically_add_cw721s": false,
		"voting_module_instantiate_info": moduleInstantiateInfo{
			CodeID: votingID, Msg: votingMsg, Admin: map[string]any{"core_module": map[string]any{}},
			Funds: nil, Label: "cw4 voting module", Salt: nil,
		},
		"proposal_modules_instantiate_info": []moduleInstantiateInfo{{
			CodeID: proposalID, Msg: proposalMsg, Admin: map[string]any{"core_module": map[string]any{}},
			Funds: nil, Label: "single proposal module", Salt: nil,
		}},
		"initial_items":   nil,
		"initial_actions": nil,
		"dao_uri":         nil,
	}
	encoded, err := json.Marshal(coreMsg)
	if err != nil {
		return "", fmt.Errorf("marshal core instantiate message: %w", err)
	}
	return string(encoded), nil
}

func queryVotingModule(ctx context.Context, chain *cosmos.CosmosChain, dao string) (string, error) {
	var address string
	err := chain.QueryContract(ctx, dao, map[string]any{"voting_module": map[string]any{}}, &address)
	return address, err
}

func queryProposalModule(ctx context.Context, chain *cosmos.CosmosChain, dao string) (string, error) {
	var modules []struct {
		Address string `json:"address"`
		Status  string `json:"status"`
	}
	err := chain.QueryContract(ctx, dao, map[string]any{
		"proposal_modules": map[string]any{"start_after": nil, "limit": nil},
	}, &modules)
	if err != nil {
		return "", err
	}
	if len(modules) != 1 {
		return "", fmt.Errorf("expected one proposal module, got %d", len(modules))
	}
	if modules[0].Status != "enabled" {
		return "", fmt.Errorf("proposal module %s is %q, want enabled", modules[0].Address, modules[0].Status)
	}
	return modules[0].Address, nil
}

func queryGroupContract(ctx context.Context, chain *cosmos.CosmosChain, voting string) (string, error) {
	var address string
	err := chain.QueryContract(ctx, voting, map[string]any{"group_contract": map[string]any{}}, &address)
	return address, err
}

func queryVotingPower(ctx context.Context, chain *cosmos.CosmosChain, dao, member string) (string, error) {
	var response struct {
		Power string `json:"power"`
	}
	err := chain.QueryContract(ctx, dao, map[string]any{
		"voting_power_at_height": map[string]any{"address": member, "height": nil},
	}, &response)
	return response.Power, err
}

func openProposal(ctx context.Context, chain *cosmos.CosmosChain, key, proposal, title, description string, fees sdk.Coins) error {
	msg, err := json.Marshal(map[string]any{"propose": map[string]any{
		"title": title, "description": description, "msgs": []any{}, "proposer": nil, "vote": nil,
	}})
	if err != nil {
		return err
	}
	_, err = chain.ExecuteContract(ctx, key, proposal, string(msg), "--gas", "auto", "--fees", fees.String())
	return err
}

func voteOnProposal(ctx context.Context, chain *cosmos.CosmosChain, key, proposal string, id uint64, vote string, fees sdk.Coins) error {
	msg, err := json.Marshal(map[string]any{"vote": map[string]any{
		"proposal_id": id, "vote": vote, "rationale": nil,
	}})
	if err != nil {
		return err
	}
	_, err = chain.ExecuteContract(ctx, key, proposal, string(msg), "--gas", "auto", "--fees", fees.String())
	return err
}

func executeProposal(ctx context.Context, chain *cosmos.CosmosChain, key, proposal string, id uint64, fees sdk.Coins) error {
	msg, err := json.Marshal(map[string]any{"execute": map[string]any{"proposal_id": id}})
	if err != nil {
		return err
	}
	_, err = chain.ExecuteContract(ctx, key, proposal, string(msg), "--gas", "auto", "--fees", fees.String())
	return err
}

func queryProposalStatus(ctx context.Context, chain *cosmos.CosmosChain, proposal string, id uint64) (string, error) {
	var response struct {
		Proposal struct {
			Status string `json:"status"`
		} `json:"proposal"`
	}
	err := chain.QueryContract(ctx, proposal, map[string]any{
		"proposal": map[string]any{"proposal_id": id},
	}, &response)
	return response.Proposal.Status, err
}

func verifyCw4Artifacts(dir string) error {
	for name, expected := range cw4Artifacts {
		file, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("open %s: %w", name, err)
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("hash %s: %w", name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", name, closeErr)
		}
		actual := hex.EncodeToString(hash.Sum(nil))
		if actual != expected {
			return fmt.Errorf("%s SHA-256 mismatch: got %s, want %s", name, actual, expected)
		}
	}
	return nil
}
