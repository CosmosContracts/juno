package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRevenue returns an instance of Revenue. If the provided withdrawer
// address is empty, it sets the value to an empty string.
func NewRevenue(contract sdk.Address, deployer, withdrawer sdk.AccAddress) Revenue {
	return Revenue{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdrawer.String(),
	}
}

// GetContractAddr returns the contract address
func (fs Revenue) GetContractAddr() sdk.Address {
	contract, err := sdk.AccAddressFromBech32(fs.ContractAddress)
	if err != nil {
		return nil
	}
	return contract
}

// GetDeployerAddr returns the contract deployer address
func (fs Revenue) GetDeployerAddr() sdk.AccAddress {
	contract, err := sdk.AccAddressFromBech32(fs.DeployerAddress)
	if err != nil {
		return nil
	}
	return contract
}

// GetWithdrawerAddr returns the account address to where the funds proceeding
// from the fees will be received. If the withdraw address is not defined, it
// defaults to the deployer address.
func (fs Revenue) GetWithdrawerAddr() sdk.AccAddress {
	contract, err := sdk.AccAddressFromBech32(fs.WithdrawerAddress)
	if err != nil {
		return nil
	}
	return contract
}

// Validate performs a stateless validation of a Revenue
func (fs Revenue) Validate() error {
	if _, err := sdk.AccAddressFromBech32(fs.ContractAddress); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(fs.DeployerAddress); err != nil {
		return err
	}

	if fs.WithdrawerAddress == "" {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, "withdrawer address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(fs.WithdrawerAddress); err != nil {
		return err
	}

	return nil
}
