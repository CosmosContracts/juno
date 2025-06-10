package types

const (
	// ModuleName defines the module name
	ModuleName = "stream"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// Key prefixes for different state listeners
var (
	// BankBalancesPrefix is the key prefix for bank module balance changes
	BankBalancesPrefix = []byte{0x02}

	// StakingDelegationPrefix is the key prefix for staking delegation changes
	StakingDelegationPrefix = []byte{0x31}

	// StakingUnbondingDelegationPrefix is the key prefix for staking unbonding delegation changes
	StakingUnbondingDelegationPrefix = []byte{0x32}
)

// Subscription types
const (
	SubscriptionTypeBalance              = "balance"
	SubscriptionTypeAllBalances          = "all_balances"
	SubscriptionTypeDelegations          = "delegations"
	SubscriptionTypeDelegation           = "delegation"
	SubscriptionTypeUnbondingDelegations = "unbonding_delegations"
	SubscriptionTypeUnbondingDelegation  = "unbonding_delegation"
)

// Event types
const (
	EventTypeBalanceChange             = "balance_change"
	EventTypeDelegationChange          = "delegation_change"
	EventTypeUnbondingDelegationChange = "unbonding_delegation_change"
)

// Module names
const (
	ModuleNameBank    = "bank"
	ModuleNameStaking = "staking"
)
