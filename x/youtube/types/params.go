package types

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{}
}

// NewParams creates a new Params object
func NewParams(
	contracts []string,
	contractGasLimit uint64,
) Params {
	return Params{}
}

// Validate performs basic validation.
func (p Params) Validate() error {
	return nil
}
