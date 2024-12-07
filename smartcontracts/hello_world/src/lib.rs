use cosmwasm_std::{entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError, StdResult};

#[entry_point]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: Binary,
) -> StdResult<Response> {
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[entry_point]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: Binary,
) -> StdResult<Response> {
    let message = String::from_utf8(msg.to_vec()).map_err(|_| StdError::generic_err("Invalid UTF-8"))?;
    Ok(Response::new()
        .add_attribute("method", "execute")
        .add_attribute("message", message))
}

#[entry_point]
pub fn query(_deps: Deps, _env: Env, _msg: Binary) -> StdResult<Binary> {
    let response = Binary::from(b"Hello, World!".to_vec());
    Ok(response)
}
