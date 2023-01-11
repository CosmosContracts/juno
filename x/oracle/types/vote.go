package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v3"
)

func NewAggregateExchangeRatePrevote(
	hash AggregateVoteHash,
	voter sdk.ValAddress,
	submitBlock uint64,
) AggregateExchangeRatePrevote {
	return AggregateExchangeRatePrevote{
		Hash:        hash.String(),
		Voter:       voter.String(),
		SubmitBlock: submitBlock,
	}
}

// String implement stringify
func (v AggregateExchangeRatePrevote) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

func NewAggregateExchangeRateVote(
	exchangeRateTuples ExchangeRateTuples,
	voter sdk.ValAddress,
) AggregateExchangeRateVote {
	return AggregateExchangeRateVote{
		ExchangeRateTuples: exchangeRateTuples,
		Voter:              voter.String(),
	}
}

// String implement stringify
func (v AggregateExchangeRateVote) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// NewExchangeRateTuple creates a ExchangeRateTuple instance
func NewExchangeRateTuple(denom string, exchangeRate sdk.Dec) ExchangeRateTuple {
	return ExchangeRateTuple{
		denom,
		exchangeRate,
	}
}

// String implement stringify
func (v ExchangeRateTuple) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// ExchangeRateTuples - array of ExchangeRateTuple
type ExchangeRateTuples []ExchangeRateTuple

// String implements fmt.Stringer interface
func (tuples ExchangeRateTuples) String() string {
	out, _ := yaml.Marshal(tuples)
	return string(out)
}

// ParseExchangeRateTuples ExchangeRateTuple parser
func ParseExchangeRateTuples(tuplesStr string) (ExchangeRateTuples, error) {
	if len(tuplesStr) == 0 {
		return nil, nil
	}

	tupleStrs := strings.Split(tuplesStr, ",")
	tuples := make(ExchangeRateTuples, len(tupleStrs))

	duplicateCheckMap := make(map[string]bool)
	for i, tupleStr := range tupleStrs {
		denomAmountStr := strings.Split(tupleStr, ":")
		if len(denomAmountStr) != 2 {
			return nil, fmt.Errorf("invalid exchange rate %s", tupleStr)
		}

		decCoin, err := sdk.NewDecFromStr(denomAmountStr[1])
		if err != nil {
			return nil, err
		}
		if !decCoin.IsPositive() {
			return nil, ErrInvalidOraclePrice
		}

		denom := strings.ToUpper(denomAmountStr[0])

		tuples[i] = ExchangeRateTuple{
			Denom:        denom,
			ExchangeRate: decCoin,
		}

		if _, ok := duplicateCheckMap[denom]; ok {
			return nil, fmt.Errorf("duplicated denom %s", denom)
		}

		duplicateCheckMap[denom] = true
	}

	return tuples, nil
}
