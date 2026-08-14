use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, QueryRequest,
    Response, StdResult,
};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(VotingPowerResponse)]
    VotingPowerAt { address: String, height: i64 },
    #[returns(VotingPowerResponse)]
    TotalVotingPowerAt { height: i64 },
    #[returns(VotingPowerOverRangeResponse)]
    VotingPowerOverRange {
        address: String,
        from_height: i64,
        to_height: i64,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum JunoQuery {
    VotingPowerAt { address: String, height: i64 },
    TotalVotingPowerAt { height: i64 },
    VotingPowerOverRange {
        address: String,
        from_height: i64,
        to_height: i64,
    },
}

impl cosmwasm_std::CustomQuery for JunoQuery {}

#[cw_serde]
pub struct VotingPowerResponse {
    pub power: String,
}

#[cw_serde]
pub struct HeightPowerPair {
    pub height: i64,
    pub power: String,
}

#[cw_serde]
pub struct VotingPowerOverRangeResponse {
    pub rows: Vec<HeightPowerPair>,
}

#[entry_point]
pub fn instantiate(
    _deps: DepsMut<JunoQuery>,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    Ok(Response::new().add_attribute("action", "instantiate_voting_power_probe"))
}

#[entry_point]
pub fn query(deps: Deps<JunoQuery>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::VotingPowerAt { address, height } => {
            let response: VotingPowerResponse = deps.querier.query(&QueryRequest::Custom(
                JunoQuery::VotingPowerAt { address, height },
            ))?;
            to_json_binary(&response)
        }
        QueryMsg::TotalVotingPowerAt { height } => {
            let response: VotingPowerResponse = deps.querier.query(&QueryRequest::Custom(
                JunoQuery::TotalVotingPowerAt { height },
            ))?;
            to_json_binary(&response)
        }
        QueryMsg::VotingPowerOverRange {
            address,
            from_height,
            to_height,
        } => {
            let response: VotingPowerOverRangeResponse = deps.querier.query(
                &QueryRequest::Custom(JunoQuery::VotingPowerOverRange {
                    address,
                    from_height,
                    to_height,
                }),
            )?;
            to_json_binary(&response)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_env, MockApi, MockQuerier, MockStorage};
    use cosmwasm_std::{from_json, Addr, ContractResult, OwnedDeps, SystemResult};
    use std::marker::PhantomData;

    fn deps_with_custom_handler() -> OwnedDeps<MockStorage, MockApi, MockQuerier<JunoQuery>, JunoQuery> {
        let querier = MockQuerier::new(&[]).with_custom_handler(|query| {
            let response = match query {
                JunoQuery::VotingPowerAt { .. } | JunoQuery::TotalVotingPowerAt { .. } => {
                    to_json_binary(&VotingPowerResponse { power: "42".to_owned() }).unwrap()
                }
                JunoQuery::VotingPowerOverRange { .. } => to_json_binary(
                    &VotingPowerOverRangeResponse {
                        rows: vec![HeightPowerPair { height: 7, power: "42".to_owned() }],
                    },
                )
                .unwrap(),
            };
            SystemResult::Ok(ContractResult::Ok(response))
        });
        OwnedDeps {
            storage: MockStorage::default(),
            api: MockApi::default(),
            querier,
            custom_query_type: PhantomData,
        }
    }

    #[test]
    fn forwards_all_custom_query_variants() {
        let mut deps = deps_with_custom_handler();
        instantiate(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("creator"), &[]),
            InstantiateMsg {},
        )
        .unwrap();

        let power: VotingPowerResponse = from_json(
            query(
                deps.as_ref(),
                mock_env(),
                QueryMsg::VotingPowerAt { address: "juno1voter".to_owned(), height: 12 },
            )
            .unwrap(),
        )
        .unwrap();
        assert_eq!(power.power, "42");

        let total: VotingPowerResponse = from_json(
            query(deps.as_ref(), mock_env(), QueryMsg::TotalVotingPowerAt { height: 12 }).unwrap(),
        )
        .unwrap();
        assert_eq!(total.power, "42");

        let range: VotingPowerOverRangeResponse = from_json(
            query(
                deps.as_ref(),
                mock_env(),
                QueryMsg::VotingPowerOverRange {
                    address: "juno1voter".to_owned(),
                    from_height: 7,
                    to_height: 12,
                },
            )
            .unwrap(),
        )
        .unwrap();
        assert_eq!(range.rows, vec![HeightPowerPair { height: 7, power: "42".to_owned() }]);
    }
}
