import { useState, useCallback } from "react";

export function useKeplrSigner(chainId, rpcEndpoint) {
  const [address, setAddress] = useState("");
  const [signer, setSigner] = useState(null);

  const connectKeplr = useCallback(async () => {
    if (!window.keplr) {
      alert("Please install Keplr extension!");
      return;
    }
    await window.keplr.enable(chainId);
    const offlineSigner = window.getOfflineSigner(chainId);
    const accounts = await offlineSigner.getAccounts();
    setAddress(accounts[0].address);
    setSigner(offlineSigner);
  }, [chainId]);

  return { address, signer, connectKeplr };
} 