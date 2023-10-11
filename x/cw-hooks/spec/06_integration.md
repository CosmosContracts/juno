# CosmWasm Integration

This module does __not__ require any custom bindings. Rather, you must just add the following Sudo messages to your contract.

If you have not permissionlessly registered your contract, you can do so [here](./04_register.md)

You can find a basic [cw-hooks contract here](https://github.com/Reecepbcups/cw-juno-staking-hooks-example)

## Implementation

## Staking

Add the following to your Rust Contract:

```rust
// msg.rs
use cosmwasm_schema::cw_serde;

#[cw_serde]
pub enum SudoMsg {    
    // Validators
    AfterValidatorCreated {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorRemoved {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    BeforeValidatorModified {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorModified {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorBonded {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    AfterValidatorBeginUnbonding {
        moniker: String,
        validator_address: String,
        commission: String,
        validator_tokens: String,
        bonded_tokens: String,
        bond_status: String,
    },
    BeforeValidatorSlashed {
        moniker: String,
        validator_address: String,
        slashed_amount: String,
    },

    // Delegations
    BeforeDelegationCreated {
        validator_address: String,
        delegator_address: String,
        shares: String,
    },
    BeforeDelegationSharesModified {
        validator_address: String,
        delegator_address: String,
        shares: String,
    },
    AfterDelegationModified {
        validator_address: String,
        delegator_address: String,
        shares: String,
    },
    BeforeDelegationRemoved {
        validator_address: String,
        delegator_address: String,
        shares: String,
    },
}

// state.rs
#[cw_serde]
pub struct ValidatorSlashed {
    pub moniker: String,
    pub validator_address: String,
    pub slashed_amount: String,
}
pub const VALIDATOR_SLASHED: Item<ValidatorSlashed> = Item::new("vs");

// contract.rs
use crate::state::{ValidatorSlashed, VALIDATOR_SLASHED};

pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {

    ...

    VALIDATOR_SLASHED.save(
        deps.storage,
        &ValidatorSlashed {
            moniker: "".to_string(),
            validator_address: "".to_string(),
            slashed_amount: "".to_string(),
        },
    )?;

    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {       

        ...

        SudoMsg::BeforeValidatorSlashed {
            moniker,
            validator_address,
            slashed_amount,
        } => {
            // perform ANY logic here when a validator is slashed on your contract.
            VALIDATOR_SLASHED.save(
                deps.storage,
                &ValidatorSlashed {
                    moniker,
                    validator_address,
                    slashed_amount,
                },
            )?;
            Ok(Response::new())
        }
    }
}
```

## Governance

```rust
use cosmwasm_schema::cw_serde;

pub struct VoteOption {
    pub option: String,
    pub weight: String,
}

#[cw_serde]
pub enum SudoMsg {    
    // Validators
    AfterProposalSubmission {
        proposal_id: String,
        proposer: String,
        status: String,
        submit_time: String,
        metadata: String,
        title: String,
        summary: String,
    },    
    AfterProposalDeposit {
        proposal_id: String,
        proposer: String,
        status: String,
        submit_time: String,
        metadata: String,
        title: String,
        summary: String,
    },
    AfterProposalVote {
        proposal_id: String,
        voter_address: String,
        vote_option: Vec<VoteOption>,
    },
    AfterProposalVotingPeriodEnded {
        proposal_id: String,
    },
}
```
