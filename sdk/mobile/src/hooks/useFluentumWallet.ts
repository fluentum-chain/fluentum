import { useState, useEffect, useCallback } from 'react';
import { WalletAdapter } from '../adapters/WalletAdapter';
import { FluentumRPC } from '../services/FluentumRPC';

export function useFluentumWallet() {
  const [publicKey, setPublicKey] = useState<string | null>(null);
  const [balance, setBalance] = useState<number>(0);
  const [nfts, setNFTs] = useState<Array<any>>([]);

  const connect = useCallback(async (walletName: string) => {
    const pk = await WalletAdapter.connect(walletName);
    setPublicKey(pk);
    return pk;
  }, []);

  const refreshData = useCallback(async () => {
    if (!publicKey) return;
    const [bal, nftList] = await Promise.all([
      FluentumRPC.getBalance(publicKey),
      FluentumRPC.getNFTs(publicKey)
    ]);
    setBalance(bal);
    setNFTs(nftList);
  }, [publicKey]);

  useEffect(() => {
    if (publicKey) refreshData();
  }, [publicKey, refreshData]);

  return {
    publicKey,
    balance,
    nfts,
    connect,
    refreshData,
    executeSwap: FluentumRPC.executeSwap,
    mintNFT: FluentumRPC.mintNFT
  };
} 