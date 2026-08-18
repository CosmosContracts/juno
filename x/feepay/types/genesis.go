package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	errDuplicateFeePayContract   = errors.New("duplicate feepay contract")
	errUnregisteredUsageContract = errors.New("wallet usage references unregistered feepay contract")
	errDuplicateWalletUsage      = errors.New("duplicate wallet usage for contract")
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, feePayContracts []FeePayContract) GenesisState {
	return GenesisState{
		Params:          params,
		FeePayContracts: feePayContracts,
		WalletUsages:    []FeePayWalletUsage{},
	}
}

// DefaultGenesisState sets default genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableFeepay: true,
		},
		FeePayContracts: []FeePayContract{},
		WalletUsages:    []FeePayWalletUsage{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	contracts := make(map[string]struct{}, len(gs.FeePayContracts))
	// Loop through all fee pay contracts and validate they
	// have a valid bech32 address
	for _, contract := range gs.FeePayContracts {
		if _, err := sdk.AccAddressFromBech32(contract.ContractAddress); err != nil {
			return err
		}
		if _, exists := contracts[contract.ContractAddress]; exists {
			return fmt.Errorf("%w %s", errDuplicateFeePayContract, contract.ContractAddress)
		}
		contracts[contract.ContractAddress] = struct{}{}
	}

	seenUsages := make(map[string]struct{}, len(gs.WalletUsages))
	for _, usage := range gs.WalletUsages {
		if _, err := sdk.AccAddressFromBech32(usage.ContractAddress); err != nil {
			return err
		}
		if _, err := sdk.AccAddressFromBech32(usage.WalletAddress); err != nil {
			return err
		}
		if _, exists := contracts[usage.ContractAddress]; !exists {
			return fmt.Errorf("%w %s", errUnregisteredUsageContract, usage.ContractAddress)
		}
		key := usage.ContractAddress + "\x00" + usage.WalletAddress
		if _, exists := seenUsages[key]; exists {
			return fmt.Errorf("%w %s and wallet %s", errDuplicateWalletUsage, usage.ContractAddress, usage.WalletAddress)
		}
		seenUsages[key] = struct{}{}
	}

	return nil
}
