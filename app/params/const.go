package params

const (
	// Name defines the application name of the Juno network.
	Name = "ujuno"

	// BondDenom defines the native staking token denomination.
	BondDenom = "ujuno"

	// DisplayDenom defines the name, symbol, and display value of the Juno token.
	DisplayDenom = "JUNO"

	// DefaultGasLimit - set to the same value as cosmos-sdk flags.DefaultGasLimit
	// this value is currently only used in tests.
	DefaultGasLimit = 200000
)
