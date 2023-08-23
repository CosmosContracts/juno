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
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	decorators "github.com/CosmosContracts/juno/v17/app/decorators"
	feeshareante "github.com/CosmosContracts/juno/v17/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v17/x/feeshare/keeper"
	globalfeeante "github.com/CosmosContracts/juno/v17/x/globalfee/ante"
	globalfeekeeper "github.com/CosmosContracts/juno/v17/x/globalfee/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Lower back to 1 mil after https://github.com/cosmos/relayer/issues/1255
const maxBypassMinFeeMsgGasUsage = 2_000_000

// HandlerOptions extends the SDK's AnteHandler options by requiring the IBC
// channel keeper and a BankKeeper with an added method for fee sharing.
type HandlerOptions struct {
	ante.HandlerOptions

	GovKeeper         govkeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper
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

	// inactive account (brand new, no tokens)
	// mint tokens (with x/Bank) to the account and set the fee to be valid (for FREE for the user, cost to the chain in the POC) - modify Tx for the fees ONLY

	// tx.GetGas() * minGasprice = the amount of fee to mint them from x/bank
	// update the Tx Fee Amount of be the amount of fee we minted them from x/bank OR just give them the funds and use that directly or something

	// get account
	// p := k.GetParams(ctx)
	// var minGasprice sdk.DecCoins
	// for _, c := range p.MinimumGasPrices {
	// 	if c.Denom == "ujuno" {
	// 		// get that amount
	// 		amt := c.Amount
	// 	}
	// }
	//
	// junod tx tokenfactory create-denom joel --gas=2200000 --from carbonator

	anteDecorators := []sdk.AnteDecorator{
		// check if account exists, if not, mint it tokens and use that as the fee. Then create the account (may be a future ante handler since it has fees now)
		// GLobalFee query params for minimum fee
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		// TODO: joel
		decorators.NewMsgFeePrepayDecorator(),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TxCounterStoreKey),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		decorators.MsgFilterDecorator{},
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		globalfeeante.NewFeeDecorator(options.BypassMinFeeMsgTypes, options.GlobalFeeKeeper, options.StakingKeeper, maxBypassMinFeeMsgGasUsage),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
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
