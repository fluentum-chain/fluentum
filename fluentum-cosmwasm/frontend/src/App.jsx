import React, { useState } from "react";
import { useKeplrSigner } from "./hooks/useKeplrSigner";
import { useCounterContract } from "./hooks/useCounterContract";

const CHAIN_ID = "fluentum-testnet-1";
const RPC = "http://34.44.129.207:26657";
const CONTRACT = "<your_contract_address>"; // Replace with actual contract address after deployment

export default function App() {
  const { address, signer, connectKeplr } = useKeplrSigner(CHAIN_ID, RPC);
  const {
    count, owner, txResult, error,
    queryCount, queryOwner, increment, reset, transferOwnership
  } = useCounterContract(RPC, CONTRACT, signer, address);

  const [resetValue, setResetValue] = useState("");
  const [newOwner, setNewOwner] = useState("");

  return (
    <div>
      <button onClick={connectKeplr} disabled={!!address}>
        {address ? "Keplr Connected" : "Connect Keplr"}
      </button>
      {address && <div>Your address: {address}</div>}
      <button onClick={queryCount}>Query Count</button>
      <button onClick={increment}>Increment</button>
      <div>Current count: {count}</div>
      <button onClick={queryOwner}>Query Owner</button>
      <div>Owner: {owner}</div>
      <form onSubmit={e => { e.preventDefault(); reset(Number(resetValue)); }}>
        <input
          type="number"
          placeholder="Reset value"
          value={resetValue}
          onChange={e => setResetValue(e.target.value)}
        />
        <button type="submit">Reset</button>
      </form>
      <form onSubmit={e => { e.preventDefault(); transferOwnership(newOwner); }}>
        <input
          type="text"
          placeholder="New owner address"
          value={newOwner}
          onChange={e => setNewOwner(e.target.value)}
        />
        <button type="submit">Transfer Ownership</button>
      </form>
      {txResult && <div>Tx result: {txResult.transactionHash}</div>}
      {error && <div style={{ color: "red" }}>Error: {error}</div>}
    </div>
  );
} 