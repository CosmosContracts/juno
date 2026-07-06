package app

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarketkeeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarketpost "github.com/CosmosContracts/juno/v30/x/feemarket/post"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
)

// PostHandlerOptions are the options required for constructing a FeeMarket PostHandler.
type PostHandlerOptions struct {
	AccountKeeper   authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper
	FeePayKeeper    feepaykeeper.Keeper
	StakingKeeper   feemarkettypes.StakingKeeper
}

// NewPostHandler returns a PostHandler chain with the fee deduct decorator.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for post builder")
	}
	if options.StakingKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "staking keeper is required for post builder")
	}

	postDecorators := []sdk.PostDecorator{
		feemarketpost.NewFeeMarketDeductDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeeMarketKeeper,
			options.FeePayKeeper,
			options.StakingKeeper,
		),
	}

	return sdk.ChainPostDecorators(postDecorators...), nil
}
