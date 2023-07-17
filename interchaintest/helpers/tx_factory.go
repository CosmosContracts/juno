package helpers

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func NewTxFactory(clientCtx client.Context) tx.Factory {
	return new(tx.Factory).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithSimulateAndExecute(true).
		WithGasAdjustment(1.5).
		WithChainID(clientCtx.ChainID).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
}
