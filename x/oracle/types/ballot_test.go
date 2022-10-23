package types

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestToMap(t *testing.T) {
	tests := struct {
		votes   []VoteForTally
		isValid []bool
	}{
		[]VoteForTally{
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.NewDec(1600),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.ZeroDec(),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        UmeeDenom,
				ExchangeRate: sdk.NewDec(1500),
				Power:        100,
			},
		},
		[]bool{true, false, true},
	}

	pb := ExchangeRateBallot(tests.votes)
	mapData := pb.ToMap()

	for i, vote := range tests.votes {
		exchangeRate, ok := mapData[vote.Voter.String()]
		if tests.isValid[i] {
			require.True(t, ok)
			require.Equal(t, exchangeRate, vote.ExchangeRate)
		} else {
			require.False(t, ok)
		}
	}
}

func TestSqrt(t *testing.T) {
	num := sdk.NewDecWithPrec(144, 4)
	floatNum, err := strconv.ParseFloat(num.String(), 64)
	require.NoError(t, err)

	floatNum = math.Sqrt(floatNum)
	num, err = sdk.NewDecFromStr(fmt.Sprintf("%f", floatNum))
	require.NoError(t, err)

	require.Equal(t, sdk.NewDecWithPrec(12, 2), num)
}

func TestPBPower(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	valAccAddrs, sk := GenerateRandomTestCase()
	pb := ExchangeRateBallot{}
	ballotPower := int64(0)

	for i := 0; i < len(sk.Validators()); i++ {
		power := sk.Validator(ctx, valAccAddrs[i]).GetConsensusPower(sdk.DefaultPowerReduction)
		vote := NewVoteForTally(
			sdk.ZeroDec(),
			UmeeDenom,
			valAccAddrs[i],
			power,
		)

		pb = append(pb, vote)
		require.NotEqual(t, int64(0), vote.Power)

		ballotPower += vote.Power
	}

	require.Equal(t, ballotPower, pb.Power())

	// Mix in a fake validator, the total power should not have changed.
	pubKey := secp256k1.GenPrivKey().PubKey()
	faceValAddr := sdk.ValAddress(pubKey.Address())
	fakeVote := NewVoteForTally(
		sdk.OneDec(),
		UmeeDenom,
		faceValAddr,
		0,
	)

	pb = append(pb, fakeVote)
	require.Equal(t, ballotPower, pb.Power())
}

func TestPBWeightedMedian(t *testing.T) {
	tests := []struct {
		inputs      []int64
		weights     []int64
		isValidator []bool
		median      sdk.Dec
		success     bool
	}{
		{
			// Supermajority one number
			[]int64{1, 2, 10, 100000},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(10),
			true,
		},
		{
			// Adding fake validator doesn't change outcome
			[]int64{1, 2, 10, 100000, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdk.NewDec(10),
			true,
		},
		{
			// Tie votes
			[]int64{1, 2, 3, 4},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(2),
			true,
		},
		{
			// No votes
			[]int64{},
			[]int64{},
			[]bool{true, true, true, true},
			sdk.NewDec(0),
			true,
		},
		{
			// Out of order
			[]int64{1, 2, 10, 3},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.NewDec(10),
			false,
		},
	}

	for _, tc := range tests {
		pb := ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := NewVoteForTally(
				sdk.NewDec(int64(input)),
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}

		median, err := pb.WeightedMedian()
		if tc.success {
			require.NoError(t, err)
			require.Equal(t, tc.median, median)
		} else {
			require.Error(t, err)
		}

	}
}

func TestPBStandardDeviation(t *testing.T) {
	tests := []struct {
		inputs            []sdk.Dec
		weights           []int64
		isValidator       []bool
		standardDeviation sdk.Dec
	}{
		{
			// Supermajority one number
			[]sdk.Dec{
				sdk.MustNewDecFromStr("1.0"),
				sdk.MustNewDecFromStr("2.0"),
				sdk.MustNewDecFromStr("10.0"),
				sdk.MustNewDecFromStr("100000.00"),
			},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			sdk.MustNewDecFromStr("49995.000362536252310906"),
		},
		{
			// Adding fake validator doesn't change outcome
			[]sdk.Dec{
				sdk.MustNewDecFromStr("1.0"),
				sdk.MustNewDecFromStr("2.0"),
				sdk.MustNewDecFromStr("10.0"),
				sdk.MustNewDecFromStr("100000.00"),
				sdk.MustNewDecFromStr("10000000000"),
			},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			sdk.MustNewDecFromStr("4472135950.751005519905537611"),
		},
		{
			// Tie votes
			[]sdk.Dec{
				sdk.MustNewDecFromStr("1.0"),
				sdk.MustNewDecFromStr("2.0"),
				sdk.MustNewDecFromStr("3.0"),
				sdk.MustNewDecFromStr("4.00"),
			},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			sdk.MustNewDecFromStr("1.224744871391589049"),
		},
		{
			// No votes
			[]sdk.Dec{},
			[]int64{},
			[]bool{true, true, true, true},
			sdk.NewDecWithPrec(0, 0),
		},
	}

	for _, tc := range tests {
		pb := ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := NewVoteForTally(
				input,
				UmeeDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}
		stdDev, _ := pb.StandardDeviation()

		require.Equal(t, tc.standardDeviation, stdDev)
	}
}

func TestPBStandardDeviation_Overflow(t *testing.T) {
	valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
	overflowRate, err := sdk.NewDecFromStr("100000000000000000000000000000000000000000000000000000000.0")
	require.NoError(t, err)
	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdk.OneDec(),
			UmeeSymbol,
			valAddr,
			2,
		),
		NewVoteForTally(
			sdk.NewDec(1234),
			UmeeSymbol,
			valAddr,
			2,
		),
		NewVoteForTally(
			overflowRate,
			UmeeSymbol,
			valAddr,
			1,
		),
	}

	deviation, err := pb.StandardDeviation()
	require.NoError(t, err)
	expectedDevation := sdk.MustNewDecFromStr("871.862661203013097586")
	require.Equal(t, expectedDevation, deviation)
}

func TestBallotMapToSlice(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)

	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdk.NewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdk.NewDec(12345),
			UmeeSymbol,
			valAddress[0],
			1,
		),
	}

	ballotSlice := BallotMapToSlice(map[string]ExchangeRateBallot{
		UmeeDenom:    pb,
		IbcDenomAtom: pb,
	})
	require.Equal(t, []BallotDenom{{Ballot: pb, Denom: IbcDenomAtom}, {Ballot: pb, Denom: UmeeDenom}}, ballotSlice)
}

func TestExchangeRateBallotSwap(t *testing.T) {
	valAddress := GenerateRandomValAddr(2)

	voteTallies := []VoteForTally{
		NewVoteForTally(
			sdk.NewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdk.NewDec(12345),
			UmeeSymbol,
			valAddress[1],
			1,
		),
	}

	pb := ExchangeRateBallot{voteTallies[0], voteTallies[1]}

	require.Equal(t, pb[0], voteTallies[0])
	require.Equal(t, pb[1], voteTallies[1])
	pb.Swap(1, 0)
	require.Equal(t, pb[1], voteTallies[0])
	require.Equal(t, pb[0], voteTallies[1])
}

func TestStandardDeviationUnsorted(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)
	pb := ExchangeRateBallot{
		NewVoteForTally(
			sdk.NewDec(1234),
			UmeeSymbol,
			valAddress[0],
			2,
		),
		NewVoteForTally(
			sdk.NewDec(12),
			UmeeSymbol,
			valAddress[0],
			1,
		),
	}

	deviation, err := pb.StandardDeviation()
	require.ErrorIs(t, err, ErrBallotNotSorted)
	require.Equal(t, "0.000000000000000000", deviation.String())
}

func TestClaimMapToSlice(t *testing.T) {
	valAddress := GenerateRandomValAddr(1)
	claim := NewClaim(10, 1, 4, valAddress[0])
	claimSlice := ClaimMapToSlice(map[string]Claim{
		"testClaim":    claim,
		"anotherClaim": claim,
	})
	require.Equal(t, []Claim{claim, claim}, claimSlice)
}
