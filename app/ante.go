package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	decorators "github.com/CosmosContracts/juno/v14/app/decorators"

	feeshareante "github.com/CosmosContracts/juno/v14/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v14/x/feeshare/keeper"

	gaiafeeante "github.com/cosmos/gaia/v9/x/globalfee/ante"
)

const maxBypassMinFeeMsgGasUsage = 1_000_000

func updateAppSimulationFlag(flag bool) {
	decorators.DefaultIsAppSimulation = flag
}

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper and a BankKeeper with an added method for fee sharing.
type HandlerOptions struct {
	ante.HandlerOptions

	GovKeeper         govkeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper
	FeeShareKeeper    feesharekeeper.Keeper
	BankKeeperFork    feeshareante.BankKeeper
	TxCounterStoreKey sdk.StoreKey
	WasmConfig        wasmTypes.WasmConfig
	Cdc               codec.BinaryCodec
	StakingSubspace   paramtypes.Subspace

	BypassMinFeeMsgTypes []string
	GlobalFeeSubspace    paramtypes.Subspace
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		decorators.NewMinCommissionDecorator(options.Cdc),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TxCounterStoreKey),
		ante.NewRejectExtensionOptionsDecorator(),
		decorators.MsgFilterDecorator{},
		decorators.NewGovPreventSpamDecorator(options.Cdc, options.GovKeeper),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		gaiafeeante.NewFeeDecorator(options.BypassMinFeeMsgTypes, options.GlobalFeeSubspace, options.StakingSubspace, maxBypassMinFeeMsgGasUsage),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		feeshareante.NewFeeSharePayoutDecorator(options.BankKeeperFork, options.FeeShareKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
