package simulation

import (
	"bytes"
	"fmt"

	"github.com/CosmosContracts/juno/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that umarshals the KVPair's
// Value to the corresponding mint type.
<<<<<<< HEAD
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
=======
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
>>>>>>> disperze/mint-module
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.MinterKey):
			var minterA, minterB types.Minter
<<<<<<< HEAD
			cdc.MustUnmarshal(kvA.Value, &minterA)
			cdc.MustUnmarshal(kvB.Value, &minterB)
=======
			cdc.MustUnmarshalBinaryBare(kvA.Value, &minterA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &minterB)
>>>>>>> disperze/mint-module
			return fmt.Sprintf("%v\n%v", minterA, minterB)
		default:
			panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
		}
	}
}
