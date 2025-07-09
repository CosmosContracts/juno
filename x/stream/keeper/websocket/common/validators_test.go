package common_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

func init() {
	// Set the bech32 prefixes for testing
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("juno", "junopub")
	config.SetBech32PrefixForValidator("junovaloper", "junovaloperpub")
	config.SetBech32PrefixForConsensusNode("junovalcons", "junovalconspub")
}

func TestValidateAccAddress(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		expectValid   bool
		errorContains string
	}{
		{
			name:        "valid account address",
			address:     "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve",
			expectValid: true,
		},
		{
			name:          "empty address",
			address:       "",
			expectValid:   false,
			errorContains: "address cannot be empty",
		},
		{
			name:          "invalid bech32 format",
			address:       "invalid!address",
			expectValid:   false,
			errorContains: "invalid address",
		},
		{
			name:          "wrong prefix",
			address:       "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq92u8wj",
			expectValid:   false,
			errorContains: "invalid address",
		},
		{
			name:          "invalid checksum",
			address:       "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq12345",
			expectValid:   false,
			errorContains: "invalid address",
		},
		{
			name:          "address too short",
			address:       "juno1qq",
			expectValid:   false,
			errorContains: "invalid address",
		},
		{
			name:          "address with invalid characters",
			address:       "juno1qqqqqqq#qqqqqqqqqqqqqqqqqqqqqqqq89mgve",
			expectValid:   false,
			errorContains: "invalid address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validator := common.ValidateAccAddress(tc.address)

			if tc.expectValid {
				require.NoError(t, validator.Error())
				require.True(t, validator.IsValid())
				require.NotNil(t, validator.AccAddress())
				require.Nil(t, validator.ValAddress())

				// Verify the address was parsed correctly
				addr := validator.AccAddress()
				require.Equal(t, tc.address, addr.String())
			} else {
				require.Error(t, validator.Error())
				require.False(t, validator.IsValid())
				require.Nil(t, validator.AccAddress())
				require.Contains(t, validator.Error().Error(), tc.errorContains)
			}
		})
	}
}

func TestValidateValAddress(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		expectValid   bool
		errorContains string
	}{
		{
			name:        "valid validator address",
			address:     "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf",
			expectValid: true,
		},
		{
			name:          "empty address",
			address:       "",
			expectValid:   false,
			errorContains: "validator address cannot be empty",
		},
		{
			name:          "invalid bech32 format",
			address:       "invalid!validator",
			expectValid:   false,
			errorContains: "invalid validator address",
		},
		{
			name:          "wrong prefix",
			address:       "cosmosvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdj0uaw",
			expectValid:   false,
			errorContains: "invalid validator address",
		},
		{
			name:          "account address instead of validator",
			address:       "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve",
			expectValid:   false,
			errorContains: "invalid validator address",
		},
		{
			name:          "invalid checksum",
			address:       "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq12345",
			expectValid:   false,
			errorContains: "invalid validator address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validator := common.ValidateValAddress(tc.address)

			if tc.expectValid {
				require.NoError(t, validator.Error())
				require.True(t, validator.IsValid())
				require.NotNil(t, validator.ValAddress())
				require.Nil(t, validator.AccAddress())

				// Verify the address was parsed correctly
				addr := validator.ValAddress()
				require.Equal(t, tc.address, addr.String())
			} else {
				require.Error(t, validator.Error())
				require.False(t, validator.IsValid())
				require.Nil(t, validator.ValAddress())
				require.Contains(t, validator.Error().Error(), tc.errorContains)
			}
		})
	}
}

func TestAddressValidatorValidationFunc(t *testing.T) {
	t.Run("valid address validation func", func(t *testing.T) {
		validator := common.ValidateAccAddress("juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve")
		validationFunc := validator.ValidationFunc()

		require.NotNil(t, validationFunc)
		require.NoError(t, validationFunc())
	})

	t.Run("invalid address validation func", func(t *testing.T) {
		validator := common.ValidateAccAddress("invalid")
		validationFunc := validator.ValidationFunc()

		require.NotNil(t, validationFunc)
		err := validationFunc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid address")
	})
}

func TestCompositeValidator(t *testing.T) {
	t.Run("empty composite validator", func(t *testing.T) {
		cv := common.NewCompositeValidator()
		validationFunc := cv.ValidationFunc()

		require.NotNil(t, validationFunc)
		require.NoError(t, validationFunc())
	})

	t.Run("single validator success", func(t *testing.T) {
		validator1 := func() error { return nil }

		cv := common.NewCompositeValidator(validator1)
		validationFunc := cv.ValidationFunc()

		require.NoError(t, validationFunc())
	})

	t.Run("single validator failure", func(t *testing.T) {
		expectedErr := errors.New("validation failed")
		validator1 := func() error { return expectedErr }

		cv := common.NewCompositeValidator(validator1)
		validationFunc := cv.ValidationFunc()

		err := validationFunc()
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})

	t.Run("multiple validators all succeed", func(t *testing.T) {
		validator1 := func() error { return nil }
		validator2 := func() error { return nil }
		validator3 := func() error { return nil }

		cv := common.NewCompositeValidator(validator1, validator2, validator3)
		validationFunc := cv.ValidationFunc()

		require.NoError(t, validationFunc())
	})

	t.Run("multiple validators first fails", func(t *testing.T) {
		expectedErr := errors.New("first validation failed")
		validator1 := func() error { return expectedErr }
		validator2 := func() error { return nil }
		validator3 := func() error { return nil }

		cv := common.NewCompositeValidator(validator1, validator2, validator3)
		validationFunc := cv.ValidationFunc()

		err := validationFunc()
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})

	t.Run("multiple validators middle fails", func(t *testing.T) {
		expectedErr := errors.New("middle validation failed")
		validator1 := func() error { return nil }
		validator2 := func() error { return expectedErr }
		validator3 := func() error { return nil }

		cv := common.NewCompositeValidator(validator1, validator2, validator3)
		validationFunc := cv.ValidationFunc()

		err := validationFunc()
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
}

func TestCompositeValidatorAdd(t *testing.T) {
	t.Run("add validator function", func(t *testing.T) {
		cv := common.NewCompositeValidator()

		validator1 := func() error { return nil }
		validator2 := func() error { return errors.New("fail") }

		cv.Add(validator1).Add(validator2)

		validationFunc := cv.ValidationFunc()
		err := validationFunc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "fail")
	})
}

func TestCompositeValidatorAddAddressValidator(t *testing.T) {
	t.Run("add valid address validator", func(t *testing.T) {
		cv := common.NewCompositeValidator()

		addrValidator := common.ValidateAccAddress("juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve")
		cv.AddAddressValidator(addrValidator)

		validationFunc := cv.ValidationFunc()
		require.NoError(t, validationFunc())
	})

	t.Run("add invalid address validator", func(t *testing.T) {
		cv := common.NewCompositeValidator()

		addrValidator := common.ValidateAccAddress("invalid")
		cv.AddAddressValidator(addrValidator)

		validationFunc := cv.ValidationFunc()
		err := validationFunc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid address")
	})

	t.Run("add nil address validator", func(t *testing.T) {
		cv := common.NewCompositeValidator()
		cv.AddAddressValidator(nil)

		validationFunc := cv.ValidationFunc()
		require.NoError(t, validationFunc())
	})

	t.Run("chain multiple address validators", func(t *testing.T) {
		cv := common.NewCompositeValidator()

		// Add valid account address validator
		accValidator := common.ValidateAccAddress("juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve")
		cv.AddAddressValidator(accValidator)

		// Add valid validator address validator
		valValidator := common.ValidateValAddress("junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf")
		cv.AddAddressValidator(valValidator)

		validationFunc := cv.ValidationFunc()
		require.NoError(t, validationFunc())
	})

	t.Run("chain with one invalid address validator", func(t *testing.T) {
		cv := common.NewCompositeValidator()

		// Add valid address validator
		accValidator := common.ValidateAccAddress("juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve")
		cv.AddAddressValidator(accValidator)

		// Add invalid address validator
		invalidValidator := common.ValidateAccAddress("")
		cv.AddAddressValidator(invalidValidator)

		validationFunc := cv.ValidationFunc()
		err := validationFunc()
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})
}

func TestAddressValidatorConcurrency(t *testing.T) {
	// Test concurrent access to validators
	addresses := []string{
		"juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve",
		"",
		"invalid!address",
		"juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve",
	}

	type result struct {
		valid bool
		err   error
	}

	results := make(chan result, len(addresses))

	for _, addr := range addresses {
		go func(address string) {
			validator := common.ValidateAccAddress(address)
			results <- result{
				valid: validator.IsValid(),
				err:   validator.Error(),
			}
		}(addr)
	}

	// Collect results
	validCount := 0
	invalidCount := 0
	for i := 0; i < len(addresses); i++ {
		res := <-results
		if res.valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	require.Equal(t, 2, validCount)
	require.Equal(t, 2, invalidCount)
}

func TestValidatorEdgeCases(t *testing.T) {
	t.Run("address with spaces", func(t *testing.T) {
		validator := common.ValidateAccAddress(" juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve ")
		require.Error(t, validator.Error())
		require.Contains(t, validator.Error().Error(), "invalid address")
	})

	t.Run("address with unicode characters", func(t *testing.T) {
		validator := common.ValidateAccAddress("juno1qqq😀qqqqqqqqqqqqqqqqqqqqqqqqqq89mgve")
		require.Error(t, validator.Error())
		require.Contains(t, validator.Error().Error(), "invalid address")
	})

	t.Run("very long invalid string", func(t *testing.T) {
		longString := "juno1" + string(make([]byte, 1000))
		validator := common.ValidateAccAddress(longString)
		require.Error(t, validator.Error())
		require.Contains(t, validator.Error().Error(), "invalid address")
	})
}
