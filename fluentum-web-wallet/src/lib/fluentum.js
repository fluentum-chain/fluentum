import { FluentumWeb } from '@fluentum-web/sdk'
 
export const fluentum = new FluentumWeb({
  rpcUrl: import.meta.env.VITE_RPC_URL,
  chainId: Number(import.meta.env.VITE_CHAIN_ID),
  walletConnectProjectId: import.meta.env.VITE_WALLETCONNECT_ID
}) 