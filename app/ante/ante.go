package ante

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	ibcante "github.com/cosmos/ibc-go/v10/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	corestoretypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	decorators "github.com/CosmosContracts/juno/v31/app/ante/decorators"
	feemarketkeeper "github.com/CosmosContracts/juno/v31/x/feemarket/keeper"
	feepaykeeper "github.com/CosmosContracts/juno/v31/x/feepay/keeper"
	feeshareante "github.com/CosmosContracts/juno/v31/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v31/x/feeshare/keeper"
)

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper and a BankKeeper with an added method for fee sharing.
type HandlerOptions struct {
	ante.HandlerOptions

	// cosmos sdk
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper stakingkeeper.Keeper
	BondDenom     string

	// ibc
	IBCKeeper *ibckeeper.Keeper

	// wasm
	TXCounterStoreService corestoretypes.KVStoreService
	NodeConfig            *wasmtypes.NodeConfig
	WasmKeeper            *wasmkeeper.Keeper

	// fees
	FeemarketKeeper      feemarketkeeper.Keeper
	FeepayKeeper         feepaykeeper.Keeper
	FeeshareKeeper       feesharekeeper.Keeper
	BypassMinFeeMsgTypes []string
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.NodeConfig == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm store service is required for ante builder")
	}
	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		// outermost AnteDecorator. SetUpContext must be called first
		ante.NewSetUpContextDecorator(),

		// wasm
		wasmkeeper.NewLimitSimulationGasDecorator(options.NodeConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		wasmkeeper.NewGasRegisterDecorator(options.WasmKeeper.GetGasRegister()),
		wasmkeeper.NewTxContractsDecorator(),

		// custom decorators
		MsgFilterDecorator{},
		decorators.NewChangeRateDecorator(&options.StakingKeeper),

		// cosmos sdk
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),

		// DeductFeeDecorator handles the fee deduction logic.
		// It includes correct fee routing for fee pay transactions
		// every actual fee transfer will be handled by the Feemarket Post Handler
		decorators.NewDeductFeeDecorator(
			options.FeepayKeeper,
			options.FeemarketKeeper,
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.BondDenom,
			options.BypassMinFeeMsgTypes,
			ante.NewDeductFeeDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeegrantKeeper,
				options.TxFeeChecker,
			),
		),
		feeshareante.NewFeeSharePayoutDecorator(options.BankKeeper, options.FeeshareKeeper),

		// signatures
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),

		// ibc
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
