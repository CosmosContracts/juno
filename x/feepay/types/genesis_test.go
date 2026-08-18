package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v31/x/feepay/types"
)

func TestGenesisStateValidateWalletUsages(t *testing.T) {
	contractAddr := "cosmos15u3dt79t6sxxa3x3kpkhzsy56edaa5a66wvt3kxmukqjz2sx0hesh45zsv"
	walletAddr := "cosmos168ctmpyppk90d34p3jjy658zf5a5l3w8wk35wht6ccqj4mr0yv8skhnwe8"
	contract := types.FeePayContract{ContractAddress: contractAddr, WalletLimit: 10}
	usage := types.FeePayWalletUsage{ContractAddress: contractAddr, WalletAddress: walletAddr, Uses: 3}

	tests := []struct {
		name  string
		state types.GenesisState
		valid bool
	}{
		{
			name: "valid usage",
			state: types.GenesisState{
				Params:          types.DefaultParams(),
				FeePayContracts: []types.FeePayContract{contract},
				WalletUsages:    []types.FeePayWalletUsage{usage},
			},
			valid: true,
		},
		{
			name: "usage references unregistered contract",
			state: types.GenesisState{
				Params:       types.DefaultParams(),
				WalletUsages: []types.FeePayWalletUsage{usage},
			},
		},
		{
			name: "duplicate usage",
			state: types.GenesisState{
				Params:          types.DefaultParams(),
				FeePayContracts: []types.FeePayContract{contract},
				WalletUsages:    []types.FeePayWalletUsage{usage, usage},
			},
		},
		{
			name: "invalid wallet",
			state: types.GenesisState{
				Params:          types.DefaultParams(),
				FeePayContracts: []types.FeePayContract{contract},
				WalletUsages: []types.FeePayWalletUsage{{
					ContractAddress: contractAddr,
					WalletAddress:   "not-an-address",
					Uses:            1,
				}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGenesisStateValidatePreservesErrorMessages(t *testing.T) {
	contractAddr := "cosmos15u3dt79t6sxxa3x3kpkhzsy56edaa5a66wvt3kxmukqjz2sx0hesh45zsv"
	walletAddr := "cosmos168ctmpyppk90d34p3jjy658zf5a5l3w8wk35wht6ccqj4mr0yv8skhnwe8"
	contract := types.FeePayContract{ContractAddress: contractAddr, WalletLimit: 10}
	usage := types.FeePayWalletUsage{ContractAddress: contractAddr, WalletAddress: walletAddr, Uses: 3}

	tests := []struct {
		name  string
		state types.GenesisState
		want  string
	}{
		{
			name:  "duplicate contract",
			state: types.GenesisState{FeePayContracts: []types.FeePayContract{contract, contract}},
			want:  "duplicate feepay contract " + contractAddr,
		},
		{
			name:  "unregistered contract usage",
			state: types.GenesisState{WalletUsages: []types.FeePayWalletUsage{usage}},
			want:  "wallet usage references unregistered feepay contract " + contractAddr,
		},
		{
			name: "duplicate wallet usage",
			state: types.GenesisState{
				FeePayContracts: []types.FeePayContract{contract},
				WalletUsages:    []types.FeePayWalletUsage{usage, usage},
			},
			want: "duplicate wallet usage for contract " + contractAddr + " and wallet " + walletAddr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.EqualError(t, tc.state.Validate(), tc.want)
		})
	}
}
