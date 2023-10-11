package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	builderante "github.com/skip-mev/pob/x/builder/ante"
	builderkeeper "github.com/skip-mev/pob/x/builder/keeper"

	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	decorators "github.com/CosmosContracts/juno/v18/app/decorators"
	feepayante "github.com/CosmosContracts/juno/v18/x/feepay/ante"
	feepaykeeper "github.com/CosmosContracts/juno/v18/x/feepay/keeper"
	feeshareante "github.com/CosmosContracts/juno/v18/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v18/x/feeshare/keeper"
	globalfeekeeper "github.com/CosmosContracts/juno/v18/x/globalfee/keeper"
)

// Lower back to 1 mil after https://github.com/cosmos/relayer/issues/1255
const maxBypassMinFeeMsgGasUsage = 2_000_000

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper and a BankKeeper with an added method for fee sharing.
type HandlerOptions struct {
	ante.HandlerOptions

	GovKeeper         govkeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper
	FeePayKeeper      feepaykeeper.Keeper
	FeeShareKeeper    feesharekeeper.Keeper
	BankKeeper        bankkeeper.Keeper
	TxCounterStoreKey storetypes.StoreKey
	WasmConfig        wasmtypes.WasmConfig
	Cdc               codec.BinaryCodec

	BypassMinFeeMsgTypes []string

	GlobalFeeKeeper globalfeekeeper.Keeper
	StakingKeeper   stakingkeeper.Keeper

	BuilderKeeper builderkeeper.Keeper
	TxEncoder     sdk.TxEncoder
	Mempool       builderante.Mempool

	BondDenom string
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		// GLobalFee query params for minimum fee
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TxCounterStoreKey),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		decorators.MsgFilterDecorator{},
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// globalfeeante.NewFeeDecorator(options.BypassMinFeeMsgTypes, options.GlobalFeeKeeper, options.StakingKeeper, maxBypassMinFeeMsgGasUsage), // TODO: Has a minimum fee requirement check. So if 0 fees are passed, but 500 fee is required, it fails.
		feepayante.NewDeductFeeDecorator(options.FeePayKeeper, options.GlobalFeeKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker, options.BondDenom),
		feeshareante.NewFeeSharePayoutDecorator(options.BankKeeper, options.FeeShareKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		builderante.NewBuilderDecorator(options.BuilderKeeper, options.TxEncoder, options.Mempool),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
