package types

// DefaultParams are the default module parameters for x/feepay
func DefaultParams() Params {
	return Params{
		EnableFeepay: false,
	}
}
