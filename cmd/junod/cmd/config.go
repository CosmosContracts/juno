package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	viper "github.com/spf13/viper"

	tmcli "github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	scconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

type JunoCustomClient struct {
	scconfig.ClientConfig
	Gas           string `mapstructure:"gas" json:"gas"`
	GasPrices     string `mapstructure:"gas-prices" json:"gas-prices"`
	GasAdjustment string `mapstructure:"gas-adjustment" json:"gas-adjustment"`

	Fees       string `mapstructure:"fees" json:"fees"`
	FeeGranter string `mapstructure:"fee-granter" json:"fee-granter"`
	FeePayer   string `mapstructure:"fee-payer" json:"fee-payer"`

	Note string `mapstructure:"note" json:"note"`
}

// ConfigCmd returns a CLI command to interactively create an application CLI
// config file.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	clientCtx := client.GetClientContextFromCmd(cmd)
	configPath := filepath.Join(clientCtx.HomeDir, "config")

	conf, err := getClientConfig(configPath, clientCtx.Viper)
	if err != nil {
		return fmt.Errorf("couldn't get client config: %v", err)
	}

	jcc := JunoCustomClient{
		*conf,
		os.Getenv("JUNOD_GAS"),
		os.Getenv("JUNOD_GAS_PRICES"),
		os.Getenv("JUNOD_GAS_ADJUSTMENT"),

		os.Getenv("JUNOD_FEES"),
		os.Getenv("JUNOD_FEE_GRANTER"),
		os.Getenv("JUNOD_FEE_PAYER"),

		os.Getenv("JUNOD_NOTE"),
	}

	switch len(args) {
	case 0:
		s, err := json.MarshalIndent(jcc, "", "\t")
		if err != nil {
			return err
		}

		cmd.Println(string(s))

	case 1:
		// it's a get
		key := args[0]

		switch key {
		case flags.FlagChainID:
			cmd.Println(conf.ChainID)
		case flags.FlagKeyringBackend:
			cmd.Println(conf.KeyringBackend)
		case tmcli.OutputFlag:
			cmd.Println(conf.Output)
		case flags.FlagNode:
			cmd.Println(conf.Node)
		case flags.FlagBroadcastMode:
			cmd.Println(conf.BroadcastMode)

		// Custom flags
		case flags.FlagGas:
			cmd.Println(jcc.Gas)
		case flags.FlagGasPrices:
			cmd.Println(jcc.GasPrices)
		case flags.FlagGasAdjustment:
			cmd.Println(jcc.GasAdjustment)
		case flags.FlagFees:
			cmd.Println(jcc.Fees)
		case flags.FlagFeeGranter:
			cmd.Println(jcc.FeeGranter)
		case flags.FlagFeePayer:
			cmd.Println(jcc.FeePayer)
		case flags.FlagNote:
			cmd.Println(jcc.Note)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
		}

	case 2:
		// it's set
		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			jcc.ChainID = value
		case flags.FlagKeyringBackend:
			jcc.KeyringBackend = value
		case tmcli.OutputFlag:
			jcc.Output = value
		case flags.FlagNode:
			jcc.Node = value
		case flags.FlagBroadcastMode:
			jcc.BroadcastMode = value
		case flags.FlagGas:
			jcc.Gas = value
		case flags.FlagGasPrices:
			jcc.GasPrices = value
			jcc.Fees = "" // resets since we can only use 1 at a time
		case flags.FlagGasAdjustment:
			jcc.GasAdjustment = value
		case flags.FlagFees:
			jcc.Fees = value
			jcc.GasPrices = "" // resets since we can only use 1 at a time
		case flags.FlagFeeGranter:
			jcc.FeeGranter = value
		case flags.FlagFeePayer:
			jcc.FeePayer = value
		case flags.FlagNote:
			jcc.Note = value
		default:
			return errUnknownConfigKey(key)
		}

		confFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigToFile(confFile, &jcc); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		panic("cound not execute config command")
	}

	return nil
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                          ###
###############################################################################

# The network chain ID
chain-id = "{{ .ChainID }}"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "{{ .KeyringBackend }}"
# CLI output format (text|json)
output = "{{ .Output }}"
# <host>:<port> to Tendermint RPC interface for this chain
node = "{{ .Node }}"
# Transaction broadcasting mode (sync|async|block)
broadcast-mode = "{{ .BroadcastMode }}"

###############################################################################
###                          Juno Tx Configuration                          ###
###############################################################################

# Amount of gas per transaction
gas = "{{ .Gas }}"
# Price per unit of gas (ex: 0.005ujuno)
gas-prices = "{{ .GasPrices }}"
gas-adjustment = "{{ .GasAdjustment }}"

# Fees to use instead of set gas prices
fees = "{{ .Fees }}"
fee-granter = "{{ .FeeGranter }}"
fee-payer = "{{ .FeePayer }}"

# Memo to include in your Transactions
note = "{{ .Note }}"
`

// writeConfigToFile parses defaultConfigTemplate, renders config using the template and writes it to
// configFilePath.
func writeConfigToFile(configFilePath string, config *JunoCustomClient) error {
	var buffer bytes.Buffer

	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return err
	}

	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*scconfig.ClientConfig, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(scconfig.ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
