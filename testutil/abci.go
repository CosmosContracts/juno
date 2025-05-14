package testutil

import (
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	coreheader "cosmossdk.io/core/header"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestHelper) Commit() {
	_, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.Ctx.BlockHeight(), Time: s.Ctx.BlockTime()})
	if err != nil {
		panic(err)
	}
	_, err = s.App.Commit()
	if err != nil {
		panic(err)
	}

	newBlockTime := s.Ctx.BlockTime().Add(time.Second)

	header := s.Ctx.BlockHeader()
	header.Time = newBlockTime
	header.Height++

	s.Ctx = s.App.BaseApp.NewUncachedContext(false, header).WithHeaderInfo(coreheader.Info{
		Height: header.Height,
		Time:   header.Time,
	})

	s.hasUsedAbci = true
}

// BeginNewBlock starts a new block.
func (s *KeeperTestHelper) BeginNewBlock() {
	var valAddr []byte

	validators, err := s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	if len(validators) >= 1 {
		valAddrFancy, err := validators[0].GetConsAddr()
		s.Require().NoError(err)
		valAddr = valAddrFancy
	} else {
		valAddrFancy := s.SetupValidator(stakingtypes.Bonded)
		validator, _ := s.App.AppKeepers.StakingKeeper.GetValidator(s.Ctx, valAddrFancy)
		valAddr2, _ := validator.GetConsAddr()
		valAddr = valAddr2
	}

	s.BeginNewBlockWithProposer(valAddr)
}

// BeginNewBlockWithProposer begins a new block with a proposer.
func (s *KeeperTestHelper) BeginNewBlockWithProposer(proposer sdk.ValAddress) {
	validator, err := s.App.AppKeepers.StakingKeeper.GetValidator(s.Ctx, proposer)
	s.Assert().NoError(err)

	valConsAddr, err := validator.GetConsAddr()
	s.Require().NoError(err)

	valAddr := valConsAddr
	newBlockTime := s.Ctx.BlockTime().Add(3 * time.Second)

	header := tmtypes.Header{Height: s.Ctx.BlockHeight() + 1, Time: newBlockTime}
	s.Ctx = s.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(s.Ctx.BlockHeight() + 1)
	voteInfos := []abci.VoteInfo{{
		Validator:   abci.Validator{Address: valAddr, Power: 1000},
		BlockIdFlag: tmtypes.BlockIDFlagCommit,
	}}
	s.Ctx = s.Ctx.WithVoteInfos(voteInfos)

	_, err = fmt.Println("beginning block ", s.Ctx.BlockHeight())
	s.Require().NoError(err)

	_, err = s.App.BeginBlocker(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.App.NewContextLegacy(false, header)
	s.hasUsedAbci = true
}

// EndBlock ends the block, and runs commit
func (s *KeeperTestHelper) EndBlock() {
	_, err := s.App.EndBlocker(s.Ctx)
	s.Require().NoError(err)
	s.hasUsedAbci = true
}
