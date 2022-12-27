package types

import (
	"crypto/rand"
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmprotocrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

const (
	IbcDenomLuna = "ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0"
	IbcDenomAtom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
)

var (
	_ StakingKeeper           = MockStakingKeeper{}
	_ stakingtypes.ValidatorI = MockValidator{}

	DenomJuno = Denom{
		BaseDenom:   JunoDenom,
		SymbolDenom: JunoSymbol,
		Exponent:    6,
	}
	DenomLuna = Denom{
		BaseDenom:   IbcDenomLuna,
		SymbolDenom: "LUNA",
		Exponent:    6,
	}
	DenomAtom = Denom{
		BaseDenom:   IbcDenomAtom,
		SymbolDenom: "ATOM",
		Exponent:    6,
	}
)

// StringWithCharset generates a new string with the size of "length" param
// repeating every character of charset, if charset is empty uses "abcd"
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)

	if len(charset) == 0 {
		charset = "abcd"
	}

	for i := 0; i < length; i++ {
		for j := 0; j < len(charset); j++ {
			b[i] = charset[j]
			i++
			if len(b) == length {
				return string(b)
			}
		}
	}

	return string(b)
}

// GenerateRandomValAddr returns N random validator addresses.
func GenerateRandomValAddr(quantity int) (validatorAddrs []sdk.ValAddress) {
	for i := 0; i < quantity; i++ {
		pubKey := secp256k1.GenPrivKey().PubKey()
		valAddr := sdk.ValAddress(pubKey.Address())
		validatorAddrs = append(validatorAddrs, valAddr)
	}

	return validatorAddrs
}

// GenerateRandomTestCase
func GenerateRandomTestCase() (valValAddrs []sdk.ValAddress, stakingKeeper MockStakingKeeper) {
	valValAddrs = []sdk.ValAddress{}
	mockValidators := []MockValidator{}

	randNum, _ := rand.Int(rand.Reader, big.NewInt(10000))
	numInputs := 10 + int((randNum.Int64() % 100))
	for i := 0; i < numInputs; i++ {
		pubKey := secp256k1.GenPrivKey().PubKey()
		valValAddr := sdk.ValAddress(pubKey.Address())
		valValAddrs = append(valValAddrs, valValAddr)

		randomPower, _ := rand.Int(rand.Reader, big.NewInt(10000))
		power := randomPower.Int64()%1000 + 1
		mockValidator := NewMockValidator(valValAddr, power)
		mockValidators = append(mockValidators, mockValidator)
	}

	stakingKeeper = NewMockStakingKeeper(mockValidators)

	return
}

// MockStakingKeeper imlements the StakingKeeper interface.
type MockStakingKeeper struct {
	validators []MockValidator
}

func NewMockStakingKeeper(validators []MockValidator) MockStakingKeeper {
	return MockStakingKeeper{
		validators: validators,
	}
}

func (sk MockStakingKeeper) Validators() []MockValidator {
	return sk.validators
}

func (sk MockStakingKeeper) Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI {
	for _, validator := range sk.validators {
		if validator.GetOperator().Equals(address) {
			return validator
		}
	}

	return nil
}

func (MockStakingKeeper) TotalBondedTokens(_ sdk.Context) sdk.Int {
	return sdk.ZeroInt()
}

func (sk MockStakingKeeper) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	return nil
}

func (MockStakingKeeper) ValidatorsPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStoreReversePrefixIterator(nil, nil)
}

func (sk MockStakingKeeper) GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) (power int64) {
	return sk.Validator(ctx, operator).GetConsensusPower(sdk.DefaultPowerReduction)
}

func (MockStakingKeeper) MaxValidators(sdk.Context) uint32 {
	return 100
}

func (MockStakingKeeper) PowerReduction(ctx sdk.Context) (res sdk.Int) {
	return sdk.DefaultPowerReduction
}

func (MockStakingKeeper) Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec) {
}

func (MockStakingKeeper) Jail(sdk.Context, sdk.ConsAddress) {}

// MockValidator implements the ValidatorI interface.
type MockValidator struct {
	power    int64
	operator sdk.ValAddress
}

func NewMockValidator(valAddr sdk.ValAddress, power int64) MockValidator {
	return MockValidator{
		power:    power,
		operator: valAddr,
	}
}

func (MockValidator) IsJailed() bool {
	return false
}

func (MockValidator) GetMoniker() string {
	return ""
}

func (MockValidator) GetStatus() stakingtypes.BondStatus {
	return stakingtypes.Bonded
}

func (MockValidator) IsBonded() bool {
	return true
}

func (MockValidator) IsUnbonded() bool {
	return false
}

func (MockValidator) IsUnbonding() bool {
	return false
}

func (v MockValidator) GetOperator() sdk.ValAddress {
	return v.operator
}

func (MockValidator) ConsPubKey() (cryptotypes.PubKey, error) {
	return nil, nil
}

func (MockValidator) TmConsPublicKey() (tmprotocrypto.PublicKey, error) {
	return tmprotocrypto.PublicKey{}, nil
}

func (MockValidator) GetConsAddr() (sdk.ConsAddress, error) {
	return nil, nil
}

func (v MockValidator) GetTokens() sdk.Int {
	return sdk.TokensFromConsensusPower(v.power, sdk.DefaultPowerReduction)
}

func (v MockValidator) GetBondedTokens() sdk.Int {
	return sdk.TokensFromConsensusPower(v.power, sdk.DefaultPowerReduction)
}

func (v MockValidator) GetConsensusPower(powerReduction sdk.Int) int64 {
	return v.power
}

func (v *MockValidator) SetConsensusPower(power int64) {
	v.power = power
}

func (v MockValidator) GetCommission() sdk.Dec {
	return sdk.ZeroDec()
}

func (v MockValidator) GetMinSelfDelegation() sdk.Int {
	return sdk.OneInt()
}

func (v MockValidator) GetDelegatorShares() sdk.Dec {
	return sdk.NewDec(v.power)
}

func (v MockValidator) TokensFromShares(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v MockValidator) TokensFromSharesTruncated(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v MockValidator) TokensFromSharesRoundUp(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v MockValidator) SharesFromTokens(amt sdk.Int) (sdk.Dec, error) {
	return sdk.ZeroDec(), nil
}

func (v MockValidator) SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error) {
	return sdk.ZeroDec(), nil
}
