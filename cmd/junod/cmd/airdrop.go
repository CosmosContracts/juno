package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
)

const (
	flagJunoWhaleCap = "juno-whalecap"
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
	// JunoBalance = 1:1 with atom, capped to 50k
	JunoBalance sdk.Int `json:"juno_balance"`
	Juno sdk.Int `json:"juno_balance_bonus"`
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
		Use:   "export-airdrop-snapshot [airdrop-to-denom] [input-genesis-file] [input-exchange-addresses] [output-snapshot-json] --juno-whalecap=[juno-whalecap]",
		Short: "Export Juno snapshot from a provided cosmos-sdk v0.36 genesis export",
		Long: `Export a Juno snapshot snapshot from a provided cosmos-sdk v0.36 genesis export
Sample genesis file:
	https://raw.githubusercontent.com/cephalopodequipment/cosmoshub-3/master/genesis.json
Example:
    junod export-airdrop-genesis uatom ~/.gaiad/config/genesis.json ./exchanges.json ../snapshot.json --juno-whalecap=50000000000
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
	- Snapshot
		file is at "../snapshot.json"
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			denom := args[0]
			genesisFile := args[1]
			exchangeFile := args[2]
			snapshotOutput := args[3]

			// Parse CLI input for juno supply
			junoWhaleCapStr, err := cmd.Flags().GetString(flagJunoWhaleCap)
			if err != nil {
				return fmt.Errorf("failed to get juno total supply: %w", err)
			}
			junoWhaleCap, ok := sdk.NewIntFromString(junoWhaleCapStr)
			if !ok {
				return fmt.Errorf("failed to parse juno supply: %s", junoWhaleCap)
			}

			// Read genesis file
			genesisJSON, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJSON.Close()

			byteValue, _ := ioutil.ReadAll(genesisJSON)

			var genStateV036 GenesisStateV036

			setCosmosBech32Prefixes()
			err = aminoCodec.UnmarshalJSON(byteValue, &genStateV036)
			if err != nil {
				return err
			}

			// Read Poll file
			exchangeJSON, err := os.Open(exchangeFile)
			if err != nil {
				return err
			}
			defer exchangeJSON.Close()

			exchangeBytes, _ := ioutil.ReadAll(exchangeJSON)
			var exchangeAddresses []string
			err = json.Unmarshal(exchangeBytes, &exchangeAddresses)
			if err != nil {
				return err
			}

			exchangeMap := make(map[string]bool)
			for _, p := range exchangeAddresses {
				exchangeMap[p] = true
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

				// If an account was delegated to an exchange skip
				if exchangeMap[sdk.AccAddress(delegation.ValidatorAddress.Bytes()).String()] {
					continue
				}

				stakedAtoms := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.AtomBalance = acc.AtomBalance.Add(stakedAtoms)
				acc.AtomStakedBalance = acc.AtomStakedBalance.Add(stakedAtoms)

				snapshot[address] = acc
			}

			totalJunoBalance := sdk.NewInt(0)

			for address, acc := range snapshot {
				allAtoms := acc.AtomBalance.ToDec()

				//Remove dust accounts
				if allAtoms.LTE(sdk.NewDec(1000000)) {
					delete(snapshot, address);
					continue
				}

				stakedAtoms := acc.AtomStakedBalance.ToDec()
				stakedPercent := stakedAtoms.Quo(allAtoms)
				acc.AtomStakedPercent = stakedPercent

				// We could use math.Min but too many type conversions
				if acc.AtomStakedBalance.GTE(junoWhaleCap) {
					acc.JunoBalance = junoWhaleCap
				} else {
					acc.JunoBalance = stakedAtoms.RoundInt()
				}

				totalJunoBalance = totalJunoBalance.Add(acc.JunoBalance)

				snapshot[address] = acc
			}

			fmt.Printf("cosmos accounts: %d\n", len(snapshot))
			fmt.Printf("atomTotalSupply: %s\n", totalAtomBalance.String())

			// export snapshot json
			snapshotJSON, err := aminoCodec.MarshalJSON(snapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}
			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	cmd.Flags().String(flagJunoWhaleCap, "", "JUNO whale cap")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

const (
	flagGenesisTime     = "genesis-time"
)

// AddAirdropAccounts Add balances of accounts to genesis, based on cosmos hub snapshot file
func AddAirdropAccounts()  *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-airdrop-accounts [airdrop-snapshot-file] [denom]",
		Short: "Add balances of accounts to genesis, based on cosmos hub snapshot file",
		Args:  cobra.ExactArgs(2),
		Long: fmt.Sprintf(`Add balances of accounts to genesis, based on cosmos hub snapshot file
Example:
$ %s add-airdrop-accounts /path/to/snapshot.json
`, version.AppName),

		RunE: func(cmd *cobra.Command, args []string) error {

			var ctx = client.GetClientContextFromCmd(cmd)
			aminoCodec := ctx.LegacyAmino.Amino
			depCdc := ctx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(ctx.HomeDir)

			blob, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}

			snapshot := make(map[string]SnapshotFields)
			err = aminoCodec.UnmarshalJSON(blob, &snapshot)
			if err != nil {
				return err
			}

			denom := args[1]

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)

			fmt.Printf("Accounts %d\n", len(snapshot))

			count := 0
			for address, acc := range snapshot {

				addr, err := ConvertCosmosAddressToJuno(address)
				if err != nil {
					return err
				}

				// Skip if account already exists
				if accs.Contains(addr) {
					continue;
				}

				coin := sdk.NewCoin(denom, acc.JunoBalance)
				coins := sdk.NewCoins(coin)
			
				// create concrete account type based on input parameters
				balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
				genAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)

				accs = append(accs, genAccount)

				bankGenState.Balances = append(bankGenState.Balances, balances)

				count++;

				if (count % 1000 == 0) {
					fmt.Printf("Progress (%d of %d)\n", count, len(snapshot))
				}
			}

			fmt.Println("Done! Sorting...")
			accs = authtypes.SanitizeGenesisAccounts(accs)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}
			appState[authtypes.ModuleName] = authGenStateBz


			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			
			fmt.Printf("Saving genesis...")
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().String(flags.FlagChainID, "", "override chain_id with this flag")
	

	return cmd
}

// ConvertCosmosAddressToJuno convert cosmos1 address to juno1 address
func ConvertCosmosAddressToJuno(address string) (sdk.AccAddress, error) {

	config := sdk.GetConfig()

	junoPrefix := config.GetBech32AccountAddrPrefix()

	_, bytes, err := bech32.DecodeAndConvert(address)
	if (err != nil) {
		return nil, err
	}

	newAddr, err := bech32.ConvertAndEncode(junoPrefix, bytes)
	if (err != nil) {
		return nil, err
	}

	sdkAddr, err := sdk.AccAddressFromBech32(newAddr)
	if (err != nil) {
		return nil, err
	}

	return sdkAddr, nil
}