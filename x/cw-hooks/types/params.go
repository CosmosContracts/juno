package types

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(250_000, 3)
}

// NewParams creates a new Params object
func NewParams(contractGasLimit uint64, contractFailureRemovalThreshold uint64) Params {
	return Params{
		ContractGasLimit:                contractGasLimit,
		ContractFailureRemovalThreshold: contractFailureRemovalThreshold,
	}
}

// Validate performs basic validation.
func (Params) Validate() error {
	return nil
}
