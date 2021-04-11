package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
)

const (
	flagJunoSupply = "juno-supply"
)

// GenesisStateV036 is minimum structure to import airdrop accounts
type GenesisStateV036 struct {
	AppState AppStateV036 `json:"app_state"`
}

// AppStateV036 is app state structure for app state
type AppStateV036 struct {
	Accounts []v036genaccounts.GenesisAccount `json:"accounts"`
	Staking  v036staking.GenesisState         `json:"staking"`
}

type Vote struct {
	Address string `json:"address"`
	Vote    string `json:"vote"`
}

type Poll []Vote

// SnapshotFields provide fields of snapshot per account
type SnapshotFields struct {
	AtomAddress string `json:"atom_address"`
	// Atom Balance = AtomStakedBalance + AtomUnstakedBalance
	AtomBalance         sdk.Int `json:"atom_balance"`
	AtomStakedBalance   sdk.Int `json:"atom_staked_balance"`
	AtomUnstakedBalance sdk.Int `json:"atom_unstaked_balance"`
	// AtomStakedPercent = AtomStakedBalance / AtomBalance
	AtomStakedPercent     sdk.Dec `json:"atom_staked_percent"`
	AtomOwnershipPercent  sdk.Dec `json:"atom_ownership_percent"`
	JunoNormalizedBalance sdk.Int `json:"juno_balance_normalized"`
	// JunoBalance = sqrt( AtomBalance ) * (1.0 * atom staked percent )
	JunoBalance sdk.Int `json:"osmo_balance"`
	// Juno = JunoBalanceBase * (1.0 * atom staked percent) limited 50_000 Juno
	Juno sdk.Int `json:"juno_balance_bonus"`
	// JunoBalanceBase = sqrt(atom balance)
	JunoBalanceBase sdk.Int `json:"juno_balance_base"`
	// JunoPercent = JunoNormalizedBalance / TotalJunoSupply
	JunoPercent sdk.Dec `json:"juno_ownership_percent"`
}

// setCosmosBech32Prefixes set config for cosmos address system
func setCosmosBech32Prefixes() {
	defaultConfig := sdk.NewConfig()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(defaultConfig.GetBech32AccountAddrPrefix(), defaultConfig.GetBech32AccountPubPrefix())
	config.SetBech32PrefixForValidator(defaultConfig.GetBech32ValidatorAddrPrefix(), defaultConfig.GetBech32ValidatorPubPrefix())
	config.SetBech32PrefixForConsensusNode(defaultConfig.GetBech32ConsensusAddrPrefix(), defaultConfig.GetBech32ConsensusPubPrefix())
}

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided cosmos-sdk v0.36 genesis export.
func ExportAirdropSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-airdrop-snapshot [airdrop-to-denom] [input-genesis-file] [output-snapshot-json] --juno-supply=[juno-genesis-supply]",
		Short: "Export Juno snapshot from a provided cosmos-sdk v0.36 genesis export",
		Long: `Export a Juno snapshot snapshot from a provided cosmos-sdk v0.36 genesis export
Sample genesis file:
	https://raw.githubusercontent.com/cephalopodequipment/cosmoshub-3/master/genesis.json
Example:
	osmosisd export-airdrop-genesis uatom ~/.gaiad/config/genesis.json ../snapshot.json --osmo-supply=100000000000000
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
	- Snapshot
		file is at "../snapshot.json"
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			denom := args[0]
			genesisFile := args[1]
			pollFile := args[2]
			snapshotOutput := args[3]

			// Parse CLI input for osmo supply
			junoSupplyStr, err := cmd.Flags().GetString(flagJunoSupply)
			if err != nil {
				return fmt.Errorf("failed to get osmo total supply: %w", err)
			}
			junoSupply, ok := sdk.NewIntFromString(junoSupplyStr)
			if !ok {
				return fmt.Errorf("failed to parse osmo supply: %s", junoSupplyStr)
			}

			// Read genesis file
			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var genStateV036 GenesisStateV036

			setCosmosBech32Prefixes()
			err = aminoCodec.UnmarshalJSON(byteValue, &genStateV036)
			if err != nil {
				return err
			}

			// Read Poll file
			pollJson, err := os.Open(pollFile)
			if err != nil {
				return err
			}
			defer pollJson.Close()

			pollBytes, _ := ioutil.ReadAll(genesisJson)
			var pollData Poll
			err = json.Unmarshal(pollBytes, &pollData)
			if err != nil {
				return err
			}

			pollMap := make(map[string]bool)

			for _, p := range pollData {
				pollMap[p.Address] = true
			}

			// Produce the map of address to total atom balance, both staked and unstaked
			snapshot := make(map[string]SnapshotFields)

			totalAtomBalance := sdk.NewInt(0)
			for _, account := range genStateV036.AppState.Accounts {

				balance := account.Coins.AmountOf(denom)
				totalAtomBalance = totalAtomBalance.Add(balance)

				if account.ModuleName != "" {
					continue
				}

				snapshot[account.Address.String()] = SnapshotFields{
					AtomAddress:         account.Address.String(),
					AtomBalance:         balance,
					AtomUnstakedBalance: balance,
					AtomStakedBalance:   sdk.ZeroInt(),
				}
			}

			for _, unbonding := range genStateV036.AppState.Staking.UnbondingDelegations {
				address := unbonding.DelegatorAddress.String()
				acc, ok := snapshot[address]
				if !ok {
					panic("no account found for unbonding")
				}

				unbondingAtoms := sdk.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingAtoms = unbondingAtoms.Add(entry.Balance)
				}

				acc.AtomBalance = acc.AtomBalance.Add(unbondingAtoms)
				acc.AtomUnstakedBalance = acc.AtomUnstakedBalance.Add(unbondingAtoms)

				snapshot[address] = acc
			}

			// Make a map from validator operator address to the v036 validator type
			validators := make(map[string]v036staking.Validator)
			for _, validator := range genStateV036.AppState.Staking.Validators {
				validators[validator.OperatorAddress.String()] = validator
			}

			for _, delegation := range genStateV036.AppState.Staking.Delegations {
				address := delegation.DelegatorAddress.String()

				acc, ok := snapshot[address]
				if !ok {
					panic("no account found for delegation")
				}

				val := validators[delegation.ValidatorAddress.String()]

				// If an account did not vote and their validator did not vote skip.
				if !pollMap[sdk.AccAddress(delegation.ValidatorAddress.Bytes()).String()] && !pollMap[address] {
					continue
				}
				stakedAtoms := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.AtomBalance = acc.AtomBalance.Add(stakedAtoms)
				acc.AtomStakedBalance = acc.AtomStakedBalance.Add(stakedAtoms)

				snapshot[address] = acc
			}

			totalJunoBalance := sdk.NewInt(0)

			onePointFive := sdk.MustNewDecFromStr("1.5")

			for address, acc := range snapshot {
				allAtoms := acc.AtomBalance.ToDec()

				//Remove dust accounts

				if allAtoms.LTE(sdk.NewDec(1000000)) {
					acc.JunoBalanceBase = sdk.ZeroInt()
					acc.Juno = sdk.ZeroInt()
					acc.JunoBalance = sdk.ZeroInt()
					continue
				}

				stakedAtoms := acc.AtomStakedBalance.ToDec()
				stakedPercent := stakedAtoms.Quo(allAtoms)
				acc.AtomStakedPercent = stakedPercent

				baseJuno, err := allAtoms.ApproxSqrt()
				if err != nil {
					panic(fmt.Sprintf("failed to root atom balance: %s", err))
				}
				acc.JunoBalanceBase = baseJuno.RoundInt()

				bonusJuno := baseJuno.Mul(onePointFive).Mul(stakedPercent)
				acc.Juno = bonusJuno.RoundInt()

				allJuno := baseJuno.Add(bonusJuno)
				// OsmoBalance = sqrt( all atoms) * (1 + 1.5) * (staked atom percent) =
				acc.JunoBalance = allJuno.RoundInt()

				totalJunoBalance = totalJunoBalance.Add(allJuno.RoundInt())

				snapshot[address] = acc
			}

			// normalize to desired genesis osmo supply
			noarmalizationFactor := junoSupply.ToDec().Quo(totalJunoBalance.ToDec())

			for address, acc := range snapshot {
				acc.JunoPercent = acc.JunoBalance.ToDec().Quo(totalJunoBalance.ToDec())

				acc.JunoNormalizedBalance = acc.JunoBalance.ToDec().Mul(noarmalizationFactor).RoundInt()

				snapshot[address] = acc
			}

			fmt.Printf("cosmos accounts: %d\n", len(snapshot))
			fmt.Printf("atomTotalSupply: %s\n", totalAtomBalance.String())
			fmt.Printf("osmoTotalSupply (pre-normalization): %s\n", totalJunoBalance.String())

			// export snapshot json
			snapshotJSON, err := aminoCodec.MarshalJSON(snapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}
			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	cmd.Flags().String(flagJunoSupply, "", "OSMO total genesis supply")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
