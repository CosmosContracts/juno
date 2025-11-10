package types

import (
	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	ModuleName = "cw-hooks"
	StoreKey   = ModuleName
)

var (
	ParamsKey             = collections.NewPrefix("params")
	ContractsKey          = collections.NewPrefix("contracts")
	ContractsByAddressKey = collections.NewPrefix("contract_addr_index")

	// supported modules
	StakingPrefixKey = collections.NewPrefix("staking")
	GovPrefixKey     = collections.NewPrefix("gov")
)

func ModulePrefixFromModule(module string) (collections.Prefix, error) {
	switch module {
	case "staking":
		return StakingPrefixKey, nil
	case "gov":
		return GovPrefixKey, nil
	default:
		return collections.Prefix{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "module not supported/found: %s", module)
	}
}

func BuildContractPrimaryKey(prefix collections.Prefix, addr sdk.AccAddress) collections.Pair[[]byte, sdk.AccAddress] {
	return collections.Join(prefix.Bytes(), addr)
}
