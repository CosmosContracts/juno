package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewRevenue returns an instance of Revenue. If the provided withdrawer
// address is empty, it sets the value to an empty string.
func NewRevenue(contract sdk.Address, deployer, withdrawer sdk.AccAddress) Revenue {
	// withdrawerAddr := ""
	// if len(withdrawer) > 0 {
	// 	withdrawerAddr = withdrawer.String()
	// }

	return Revenue{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdrawer.String(),
	}
}

// GetContractAddr returns the contract address
func (fs Revenue) GetContractAddr() sdk.Address {
	return sdk.MustAccAddressFromBech32(fs.ContractAddress)
}

// GetDeployerAddr returns the contract deployer address
func (fs Revenue) GetDeployerAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(fs.DeployerAddress)
}

// GetWithdrawerAddr returns the account address to where the funds proceeding
// from the fees will be received. If the withdraw address is not defined, it
// defaults to the deployer address.
func (fs Revenue) GetWithdrawerAddr() sdk.AccAddress {
	if fs.WithdrawerAddress == "" {
		return nil
	}

	return sdk.MustAccAddressFromBech32(fs.WithdrawerAddress)
}

// Validate performs a stateless validation of a Revenue
func (fs Revenue) Validate() error {
	if _, err := sdk.AccAddressFromBech32(fs.ContractAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(fs.DeployerAddress); err != nil {
		return err
	}

	if fs.WithdrawerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(fs.WithdrawerAddress); err != nil {
			return err
		}
	}

	return nil
}
