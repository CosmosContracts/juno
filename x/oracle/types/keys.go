package types

import (
	"fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// QuerierRoute is the query router key for the oracle module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the oracle module
	RouterKey = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixExchangeRate                 = []byte{0x01} // prefix for each key to a rate
	KeyPrefixFeederDelegation             = []byte{0x02} // prefix for each key to a feeder delegation
	KeyPrefixMissCounter                  = []byte{0x03} // prefix for each key to a miss counter
	KeyPrefixAggregateExchangeRatePrevote = []byte{0x04} // prefix for each key to a aggregate prevote
	KeyPrefixAggregateExchangeRateVote    = []byte{0x05} // prefix for each key to a aggregate vote
	KeyPrefixPriceHistory                 = []byte{0x06} // prefix for price history
)

// Store price history <Prefix> <baseDenom> <votePeriodCount> <PriceHistoryEntry>
func GetPriceHistoryKey(symbolDenom string) (key []byte) {
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	key = append(key, KeyPrefixPriceHistory...)
	key = append(key, []byte(upperSymbolDenom)...)
	return key
}

// GetExchangeRateKey - stored by *denom*
func GetExchangeRateKey(denom string) (key []byte) {
	key = append(key, KeyPrefixExchangeRate...)
	key = append(key, []byte(denom)...)
	return append(key, 0) // append 0 for null-termination
}

// GetFeederDelegationKey - stored by *Validator* address
func GetFeederDelegationKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixFeederDelegation...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetMissCounterKey - stored by *Validator* address
func GetMissCounterKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixMissCounter...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetAggregateExchangeRatePrevoteKey - stored by *Validator* address
func GetAggregateExchangeRatePrevoteKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixAggregateExchangeRatePrevote...)
	return append(key, address.MustLengthPrefix(v)...)
}

// GetAggregateExchangeRateVoteKey - stored by *Validator* address
func GetAggregateExchangeRateVoteKey(v sdk.ValAddress) (key []byte) {
	key = append(key, KeyPrefixAggregateExchangeRateVote...)
	return append(key, address.MustLengthPrefix(v)...)
}

var (
	// Contract: Coin denoms cannot contain this character
	KeySeparator = "|"

	historicalTimeIndexNoSeparator = "historical_time_index"

	// format is pool id | denom1 | denom2 | time
	// made for efficiently getting records given (pool id, denom1, denom2) and time bounds
	HistoricalTWAPTimeIndexPrefix = historicalTimeIndexNoSeparator + KeySeparator
)

func FormatTimeString(t time.Time) string {
	return t.UTC().Round(0).Format(sdk.SortableTimeFormat)
}

func FormatHistoricalDenomIndexKey(accumulatorWriteTime time.Time, denom string) []byte {
	timeS := FormatTimeString(accumulatorWriteTime)
	return []byte(fmt.Sprintf("%s%s%s%s", HistoricalTWAPTimeIndexPrefix, denom, KeySeparator, timeS))
}

func FormatHistoricalDenomIndexPrefix(denom string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", HistoricalTWAPTimeIndexPrefix, denom, KeySeparator))
}
