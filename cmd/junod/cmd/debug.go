package cmd

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/version"
)

var flagBech32Prefix = "prefix"

// Cmd creates a main CLI command
func DebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Tool for helping with debugging your application",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(
		CodecCmd(),
		PrefixesCmd(),
		PubkeyCmd(),
		AddrCmd(),
		RawBytesCmd(),
		ConvertBech32Cmd(),
		ExportDeriveBalancesCmd(),
		StakedToCSVCmd(),
	)

	return cmd
}

func PrefixesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "prefixes",
		Short:   "List prefixes used for Human-Readable Part (HRP) in Bech32",
		Long:    "List prefixes used in Bech32 addresses.",
		Example: fmt.Sprintf("$ %s debug prefixes", version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Printf("Bech32 Acc: %s\n", sdk.GetConfig().GetBech32AccountAddrPrefix())
			cmd.Printf("Bech32 Val: %s\n", sdk.GetConfig().GetBech32ValidatorAddrPrefix())
			cmd.Printf("Bech32 Con: %s\n", sdk.GetConfig().GetBech32ConsensusAddrPrefix())
			return nil
		},
	}
}

// CodecCmd creates and returns a new codec debug cmd.
func CodecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codec",
		Short: "Tool for helping with debugging your application codec",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(getCodecInterfaces())
	cmd.AddCommand(getCodecInterfaceImpls())

	return cmd
}

// getCodecInterfaces creates and returns a new cmd used for listing all registered interfaces on the application codec.
func getCodecInterfaces() *cobra.Command {
	return &cobra.Command{
		Use:   "list-interfaces",
		Short: "List all registered interface type URLs",
		Long:  "List all registered interface type URLs using the application codec",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			iFaces := clientCtx.Codec.InterfaceRegistry().ListAllInterfaces()
			for _, iFace := range iFaces {
				cmd.Println(iFace)
			}
			return nil
		},
	}
}

// getCodecInterfaceImpls creates and returns a new cmd used for listing all registered implementations of a given interface on the application codec.
func getCodecInterfaceImpls() *cobra.Command {
	return &cobra.Command{
		Use:   "list-implementations [interface]",
		Short: "List the registered type URLs for the provided interface",
		Long:  "List the registered type URLs that can be used for the provided interface name using the application codec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			impls := clientCtx.Codec.InterfaceRegistry().ListImplementations(args[0])
			for _, imp := range impls {
				cmd.Println(imp)
			}
			return nil
		},
	}
}

// get cmd to convert any bech32 address to a juno prefix.
func ConvertBech32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32-convert [bech32 string]",
		Short: "Convert any bech32 string to the juno prefix",
		Long: `Convert any bech32 string to the juno prefix
Especially useful for converting cosmos addresses to juno addresses
Example:
	junod bech32-convert juno1ey69r37gfxvxg62sh4r0ktpuc46pzjrm5cxnjg -p osmo
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32prefix, err := cmd.Flags().GetString(flagBech32Prefix)
			if err != nil {
				return err
			}

			_, bz, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return err
			}

			bech32Addr, err := bech32.ConvertAndEncode(bech32prefix, bz)
			if err != nil {
				panic(err)
			}

			cmd.Println(bech32Addr)

			return nil
		},
	}

	cmd.Flags().StringP(flagBech32Prefix, "p", "juno", "Bech32 Prefix to encode to")

	return cmd
}

// getPubKeyFromString decodes SDK PubKey using JSON marshaler.
func getPubKeyFromString(ctx client.Context, pkstr string) (cryptotypes.PubKey, error) {
	var pk cryptotypes.PubKey
	err := ctx.Codec.UnmarshalInterfaceJSON([]byte(pkstr), &pk)
	return pk, err
}

func PubkeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pubkey [pubkey]",
		Short: "Decode a pubkey from proto JSON",
		Long: fmt.Sprintf(`Decode a pubkey from proto JSON and display it's address.

Example:
$ %s debug pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"}'
			`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pk, err := getPubKeyFromString(clientCtx, args[0])
			if err != nil {
				return err
			}
			cmd.Println("Address:", pk.Address())
			cmd.Println("PubKey Hex:", hex.EncodeToString(pk.Bytes()))
			return nil
		},
	}
}

func AddrCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "addr [address]",
		Short: "Convert an address between hex and bech32",
		Long: fmt.Sprintf(`Convert an address between hex encoding and bech32.

Example:
$ %s debug addr cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg
			`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrString := args[0]
			var addr []byte

			// try hex, then bech32
			var err error
			addr, err = hex.DecodeString(addrString)
			if err != nil {
				var err2 error
				addr, err2 = sdk.AccAddressFromBech32(addrString)
				if err2 != nil {
					var err3 error
					addr, err3 = sdk.ValAddressFromBech32(addrString)

					if err3 != nil {
						return fmt.Errorf("expected hex or bech32. Got errors: hex: %v, bech32 acc: %v, bech32 val: %v", err, err2, err3)
					}
				}
			}

			cmd.Println("Address:", addr)
			cmd.Printf("Address (hex): %X\n", addr)
			cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr))
			cmd.Printf("Bech32 Val: %s\n", sdk.ValAddress(addr))
			return nil
		},
	}
}

func RawBytesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "raw-bytes [raw-bytes]",
		Short: "Convert raw bytes output (eg. [10 21 13 255]) to hex",
		Long: fmt.Sprintf(`Convert raw-bytes to hex.

Example:
$ %s debug raw-bytes [72 101 108 108 111 44 32 112 108 97 121 103 114 111 117 110 100]
			`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			stringBytes := args[0]
			stringBytes = strings.Trim(stringBytes, "[")
			stringBytes = strings.Trim(stringBytes, "]")
			spl := strings.Split(stringBytes, " ")

			byteArray := []byte{}
			for _, s := range spl {
				b, err := strconv.ParseInt(s, 10, 8)
				if err != nil {
					return err
				}
				byteArray = append(byteArray, byte(b))
			}
			_, err := fmt.Printf("%X\n", byteArray)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
