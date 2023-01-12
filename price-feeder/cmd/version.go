package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	flagFormat = "format"
)

var (
	// Version defines the application version (defined at compile time)
	Version = ""

	// Commit defines the application commit hash (defined at compile time)
	Commit = ""

	// SDKVersion defines the sdk version (defined at compile time)
	SDKVersion = ""

	versionFormat string
)

type versionInfo struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	SDK     string `json:"sdk" yaml:"sdk"`
	Go      string `json:"go" yaml:"go"`
}

func getVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print binary version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			verInfo := versionInfo{
				Version: Version,
				Commit:  Commit,
				SDK:     SDKVersion,
				Go:      fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			}

			var bz []byte

			var err error
			switch versionFormat {
			case "json":
				bz, err = json.Marshal(verInfo)

			default:
				bz, err = yaml.Marshal(&verInfo)
			}
			if err != nil {
				return err
			}

			_, err = fmt.Println(string(bz))
			return err
		},
	}

	versionCmd.Flags().StringVar(&versionFormat, flagFormat, "text", "Print the version in the given format (text|json)")

	return versionCmd
}
