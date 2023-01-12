package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VoteForTally is a convenience wrapper to reduce redundant lookup cost.
type VoteForTally struct {
	Denom        string
	ExchangeRate sdk.Dec
	Voter        sdk.ValAddress
	Power        int64
}

// NewVoteForTally returns a new VoteForTally instance.
func NewVoteForTally(rate sdk.Dec, denom string, voter sdk.ValAddress, power int64) VoteForTally {
	return VoteForTally{
		ExchangeRate: rate,
		Denom:        denom,
		Voter:        voter,
		Power:        power,
	}
}

// ExchangeRateBallot is a convenience wrapper around a ExchangeRateVote slice.
type ExchangeRateBallot []VoteForTally

// ToMap return organized exchange rate map by validator.
func (pb ExchangeRateBallot) ToMap() map[string]sdk.Dec {
	exchangeRateMap := make(map[string]sdk.Dec)
	for _, vote := range pb {
		if vote.ExchangeRate.IsPositive() {
			exchangeRateMap[vote.Voter.String()] = vote.ExchangeRate
		}
	}

	return exchangeRateMap
}

// Power returns the total amount of voting power in the ballot.
func (pb ExchangeRateBallot) Power() int64 {
	var totalPower int64
	for _, vote := range pb {
		totalPower += vote.Power
	}

	return totalPower
}

// WeightedMedian returns the median weighted by the power of the ExchangeRateVote.
// CONTRACT: The ballot must be sorted.
func (pb ExchangeRateBallot) WeightedMedian() (sdk.Dec, error) {
	if !sort.IsSorted(pb) {
		return sdk.ZeroDec(), ErrBallotNotSorted
	}
	totalPower := pb.Power()

	if pb.Len() > 0 {
		var pivot int64
		for _, v := range pb {
			votePower := v.Power

			pivot += votePower
			if pivot >= (totalPower / 2) {
				return v.ExchangeRate, nil
			}
		}
	}

	return sdk.ZeroDec(), nil
}

// StandardDeviation returns the standard deviation by the power of the ExchangeRateVote.
func (pb ExchangeRateBallot) StandardDeviation() (sdk.Dec, error) {
	if len(pb) == 0 {
		return sdk.ZeroDec(), nil
	}

	median, err := pb.WeightedMedian()
	if err != nil {
		return sdk.ZeroDec(), err
	}

	sum := sdk.ZeroDec()
	ballotLength := int64(len(pb))
	for _, v := range pb {
		func() {
			defer func() {
				if e := recover(); e != nil {
					ballotLength--
				}
			}()
			deviation := v.ExchangeRate.Sub(median)
			sum = sum.Add(deviation.Mul(deviation))
		}()
	}

	variance := sum.QuoInt64(ballotLength)

	standardDeviation, err := variance.ApproxSqrt()
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return standardDeviation, nil
}

// Len implements sort.Interface
func (pb ExchangeRateBallot) Len() int {
	return len(pb)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (pb ExchangeRateBallot) Less(i, j int) bool {
	return pb[i].ExchangeRate.LT(pb[j].ExchangeRate)
}

// Swap implements sort.Interface.
func (pb ExchangeRateBallot) Swap(i, j int) {
	pb[i], pb[j] = pb[j], pb[i]
}

// BallotDenom is a convenience wrapper for setting rates deterministically.
type BallotDenom struct {
	Ballot ExchangeRateBallot
	Denom  string
}

// BallotMapToSlice returns an array of sorted exchange rate ballots.
func BallotMapToSlice(votes map[string]ExchangeRateBallot) []BallotDenom {
	b := make([]BallotDenom, len(votes))
	i := 0
	for denom, ballot := range votes {
		b[i] = BallotDenom{
			Denom:  denom,
			Ballot: ballot,
		}
		i++
	}
	sort.Slice(b, func(i, j int) bool {
		return b[i].Denom < b[j].Denom
	})
	return b
}

// Claim is an interface that directs its rewards to an attached bank account.
type Claim struct {
	Power     int64
	Weight    int64
	WinCount  int64
	Recipient sdk.ValAddress
}

// NewClaim generates a Claim instance.
func NewClaim(power, weight, winCount int64, recipient sdk.ValAddress) Claim {
	return Claim{
		Power:     power,
		Weight:    weight,
		WinCount:  winCount,
		Recipient: recipient,
	}
}

// ClaimMapToSlice returns an array of sorted exchange rate ballots.
func ClaimMapToSlice(claims map[string]Claim) []Claim {
	c := make([]Claim, len(claims))
	i := 0
	for _, claim := range claims {
		c[i] = Claim{
			Power:     claim.Power,
			Weight:    claim.Weight,
			WinCount:  claim.WinCount,
			Recipient: claim.Recipient,
		}
		i++
	}
	sort.Slice(c, func(i, j int) bool {
		return c[i].Recipient.String() < c[j].Recipient.String()
	})
	return c
}
