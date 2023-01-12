package simulation

import (
	"bytes"
	"fmt"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixExchangeRate):
			var exchangeRateA, exchangeRateB sdk.DecProto
			cdc.MustUnmarshal(kvA.Value, &exchangeRateA)
			cdc.MustUnmarshal(kvB.Value, &exchangeRateB)
			return fmt.Sprintf("%v\n%v", exchangeRateA, exchangeRateB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixFeederDelegation):
			return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixMissCounter):
			var counterA, counterB gogotypes.UInt64Value
			cdc.MustUnmarshal(kvA.Value, &counterA)
			cdc.MustUnmarshal(kvB.Value, &counterB)
			return fmt.Sprintf("%v\n%v", counterA.Value, counterB.Value)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixAggregateExchangeRatePrevote):
			var prevoteA, prevoteB types.AggregateExchangeRatePrevote
			cdc.MustUnmarshal(kvA.Value, &prevoteA)
			cdc.MustUnmarshal(kvB.Value, &prevoteB)
			return fmt.Sprintf("%v\n%v", prevoteA, prevoteB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixAggregateExchangeRateVote):
			var voteA, voteB types.AggregateExchangeRateVote
			cdc.MustUnmarshal(kvA.Value, &voteA)
			cdc.MustUnmarshal(kvB.Value, &voteB)
			return fmt.Sprintf("%v\n%v", voteA, voteB)

		default:
			panic(fmt.Sprintf("invalid oracle key prefix %X", kvA.Key[:1]))
		}
	}
}
