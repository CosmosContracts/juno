package app

import (
	"errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	corestoretypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	decorators "github.com/CosmosContracts/juno/v30/app/decorators"
	feepayante "github.com/CosmosContracts/juno/v30/x/feepay/ante"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
	feeshareante "github.com/CosmosContracts/juno/v30/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v30/x/feeshare/keeper"
	globalfeeante "github.com/CosmosContracts/juno/v30/x/globalfee/ante"
	globalfeekeeper "github.com/CosmosContracts/juno/v30/x/globalfee/keeper"
)

// Lower back to 1 mil after https://github.com/cosmos/relayer/issues/1255
const maxBypassMinFeeMsgGasUsage = 2_000_000

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper and a BankKeeper with an added method for fee sharing.
type HandlerOptions struct {
	ante.HandlerOptions

	// cosmos sdk
	StakingKeeper stakingkeeper.Keeper
	BondDenom     string
	BankKeeper    bankkeeper.Keeper

	// ibc
	IBCKeeper *ibckeeper.Keeper

	// wasm
	TXCounterStoreService corestoretypes.KVStoreService
	NodeConfig            *wasmtypes.NodeConfig
	WasmKeeper            *wasmkeeper.Keeper

	// fee modules
	FeePayKeeper         feepaykeeper.Keeper
	FeeShareKeeper       feesharekeeper.Keeper
	GlobalFeeKeeper      globalfeekeeper.Keeper
	BypassMinFeeMsgTypes []string
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
	if options.NodeConfig == nil {
		return nil, errors.New("wasm config is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errors.New("wasm store service is required for ante builder")
	}
	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	// Flag for determining if the tx is a FeePay transaction. This flag
	// is used to communicate between the FeePay decorator and the GlobalFee decorator.
	isFeePayTx := false

	// Define FeePay and Global Fee decorators. These decorators are called in different orders based on the type of
	// transaction. The FeePay decorator is called first for FeePay transactions, and the GlobalFee decorator is called
	// first for all other transactions. See the FeeRouteDecorator for more details.
	fpd := feepayante.NewDeductFeeDecorator(options.FeePayKeeper, options.GlobalFeeKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.BondDenom, &isFeePayTx)
	gfd := globalfeeante.NewFeeDecorator(options.BypassMinFeeMsgTypes, options.GlobalFeeKeeper, options.StakingKeeper, maxBypassMinFeeMsgGasUsage, &isFeePayTx)

	anteDecorators := []sdk.AnteDecorator{
		// outermost AnteDecorator. SetUpContext must be called first
		ante.NewSetUpContextDecorator(),

		// wasm
		wasmkeeper.NewLimitSimulationGasDecorator(options.NodeConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		wasmkeeper.NewGasRegisterDecorator(options.WasmKeeper.GetGasRegister()),
		wasmkeeper.NewTxContractsDecorator(),

		// custom decorators
		decorators.MsgFilterDecorator{},
		decorators.NewChangeRateDecorator(&options.StakingKeeper),

		// cosmos sdk
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),

		// juno custom modules
		// Fee route decorator calls FeePay and Global Fee decorators in different orders
		// depending on the type of incoming tx.
		feepayante.NewFeeRouteDecorator(options.FeePayKeeper, &fpd, &gfd, &isFeePayTx),
		feeshareante.NewFeeSharePayoutDecorator(options.BankKeeper, options.FeeShareKeeper),

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
