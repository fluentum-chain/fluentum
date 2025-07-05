import { useState } from "react";
import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";

export function useCounterContract(rpcEndpoint, contractAddress, signer, sender) {
  const [count, setCount] = useState(null);
  const [owner, setOwner] = useState(null);
  const [txResult, setTxResult] = useState(null);
  const [error, setError] = useState(null);

  const queryCount = async () => {
    try {
      const client = await SigningCosmWasmClient.connectWithSigner(rpcEndpoint, signer);
      const result = await client.queryContractSmart(contractAddress, { get_count: {} });
      setCount(result.count);
    } catch (err) {
      setError(err.message);
    }
  };

  const queryOwner = async () => {
    try {
      const client = await SigningCosmWasmClient.connectWithSigner(rpcEndpoint, signer);
      const result = await client.queryContractSmart(contractAddress, { get_owner: {} });
      setOwner(result.owner);
    } catch (err) {
      setError(err.message);
    }
  };

  const increment = async () => {
    try {
      const client = await SigningCosmWasmClient.connectWithSigner(rpcEndpoint, signer);
      const result = await client.execute(sender, contractAddress, { increment: {} }, "auto");
      setTxResult(result);
      await queryCount();
    } catch (err) {
      setError(err.message);
    }
  };

  const reset = async (newCount) => {
    try {
      const client = await SigningCosmWasmClient.connectWithSigner(rpcEndpoint, signer);
      const result = await client.execute(sender, contractAddress, { reset: { count: newCount } }, "auto");
      setTxResult(result);
      await queryCount();
    } catch (err) {
      setError(err.message);
    }
  };

  const transferOwnership = async (newOwner) => {
    try {
      const client = await SigningCosmWasmClient.connectWithSigner(rpcEndpoint, signer);
      const result = await client.execute(sender, contractAddress, { transfer_ownership: { new_owner: newOwner } }, "auto");
      setTxResult(result);
      await queryOwner();
    } catch (err) {
      setError(err.message);
    }
  };

  return { count, owner, txResult, error, queryCount, queryOwner, increment, reset, transferOwnership };
} 