use cosmwasm_std::{entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Addr, StdError};
use cw_storage_plus::Item;
use serde::{Deserialize, Serialize};

const COUNTER: Item<i32> = Item::new("counter");
const OWNER: Item<Addr> = Item::new("owner");

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Increment {},
    Reset { count: i32 },
    TransferOwnership { new_owner: String },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetCount {},
    GetOwner {},
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct CountResponse {
    pub count: i32,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct OwnerResponse {
    pub owner: String,
}

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: (),
) -> StdResult<Response> {
    COUNTER.save(deps.storage, &0)?;
    OWNER.save(deps.storage, &info.sender)?;
    Ok(Response::default())
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    match msg {
        ExecuteMsg::Increment {} => try_increment(deps),
        ExecuteMsg::Reset { count } => try_reset(deps, info, count),
        ExecuteMsg::TransferOwnership { new_owner } => try_transfer_ownership(deps, info, new_owner),
    }
}

fn try_increment(deps: DepsMut) -> StdResult<Response> {
    COUNTER.update(deps.storage, |count| -> StdResult<_> { Ok(count + 1) })?;
    Ok(Response::default())
}

fn try_reset(deps: DepsMut, info: MessageInfo, count: i32) -> StdResult<Response> {
    let owner = OWNER.load(deps.storage)?;
    if info.sender != owner {
        return Err(StdError::generic_err("Only the owner can reset the counter"));
    }
    COUNTER.save(deps.storage, &count)?;
    Ok(Response::default())
}

fn try_transfer_ownership(deps: DepsMut, info: MessageInfo, new_owner: String) -> StdResult<Response> {
    let owner = OWNER.load(deps.storage)?;
    if info.sender != owner {
        return Err(StdError::generic_err("Only the owner can transfer ownership"));
    }
    let new_owner_addr = deps.api.addr_validate(&new_owner)?;
    OWNER.save(deps.storage, &new_owner_addr)?;
    Ok(Response::default())
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount {} => {
            let count = COUNTER.load(deps.storage)?;
            to_binary(&CountResponse { count })
        }
        QueryMsg::GetOwner {} => {
            let owner = OWNER.load(deps.storage)?;
            to_binary(&OwnerResponse { owner: owner.to_string() })
        }
    }
} 