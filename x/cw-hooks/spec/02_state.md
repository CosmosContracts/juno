# State

## State Objects

The `x/cw-hooks` module keeps the following objects in the state:

| State Object          | Description                           | Key                                                               | Value              | Store |
| :-------------------- | :------------------------------------ | :---------------------------------------------------------------- | :----------------- | :---- |
| `Staking Contract`    | contract registered for staking events| `[]byte{"staking"} + []byte(contract_address)`                    | `[]byte{}`         | KV    |
| `Governance Contract` | contract registered for gov events    | `[]byte{"gov"} + []byte(contract_address)`                        | `[]byte{}`         | KV    |

### ContractAddress

`ContractAddress` defines the contract address that has been registered for fee distribution.

## Genesis State

The `x/cw-hooks` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains the module parameters and all registered contracts:

```go
type Params struct {
    // contract_gas_limit is the contract call gas limit
    ContractGasLimit uint64 `protobuf:"varint,1,opt,name=contract_gas_limit,json=contractGasLimit,proto3" json:"contract_gas_limit,omitempty" yaml:"contract_gas_limit"`
}

// GenesisState defines the module's genesis state.
type GenesisState struct {  
  Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params,omitempty"`
  
  StakingContractAddresses []string `protobuf:"bytes,2,rep,name=staking_contract_addresses,json=stakingContractAddresses,proto3" json:"staking_contract_addresses,omitempty" yaml:"staking_contract_addresses"`
  
  GovContractAddresses []string `protobuf:"bytes,3,rep,name=gov_contract_addresses,json=govContractAddresses,proto3" json:"gov_contract_addresses,omitempty" yaml:"gov_contract_addresses"`
}
```
