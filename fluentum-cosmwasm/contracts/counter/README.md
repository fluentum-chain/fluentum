# Advanced Counter CosmWasm Contract

This is an advanced CosmWasm smart contract for the Fluentum blockchain, featuring:
- Increment and reset counter
- Owner-only reset and ownership transfer
- Query for current count and owner

## Instantiate

No parameters required. The instantiator becomes the initial owner.

## Execute Messages

### Increment
```
{
  "increment": {}
}
```

### Reset (owner only)
```
{
  "reset": { "count": 42 }
}
```

### Transfer Ownership (owner only)
```
{
  "transfer_ownership": { "new_owner": "fluentum1..." }
}
```

## Query Messages

### Get Count
```
{
  "get_count": {}
}
```
Response:
```
{
  "count": 42
}
```

### Get Owner
```
{
  "get_owner": {}
}
```
Response:
```
{
  "owner": "fluentum1..."
}
```

## Build

```
cargo wasm
```

## Deploy to Fluentum

1. **Store contract:**
   ```
   wasmd tx wasm store target/wasm32-unknown-unknown/release/counter.wasm --from <your_wallet> --chain-id fluentum-testnet-1 --gas auto --fees 500uflux
   ```
2. **Instantiate contract:**
   ```
   wasmd tx wasm instantiate <code_id> '{}' --from <your_wallet> --label "Counter" --chain-id fluentum-testnet-1 --gas auto --fees 500uflux --admin <your_wallet>
   ```
3. **Get contract address from the transaction result.**

## Interact

- Use CosmJS, Keplr, or the provided React frontend to interact with the contract.
- See the frontend example in `../../frontend/` for integration. 