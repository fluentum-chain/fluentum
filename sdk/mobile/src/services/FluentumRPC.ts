const RPC_ENDPOINT = "https://rpc.fluentum.io";

export class FluentumRPC {
  static async query(publicKey: string, method: string, params: any = {}) {
    const response = await fetch(RPC_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: "2.0",
        id: 1,
        method: `fluentum_${method}`,
        params: { publicKey, ...params }
      })
    });

    const data = await response.json();
    return data.result;
  }

  static getBalance(publicKey: string) {
    return this.query(publicKey, "getBalance");
  }

  static getNFTs(publicKey: string) {
    return this.query(publicKey, "getNFTs");
  }

  static executeSwap(publicKey: string, tokenIn: string, tokenOut: string, amount: number) {
    return this.query(publicKey, "executeSwap", { tokenIn, tokenOut, amount });
  }

  static mintNFT(publicKey: string, metadata: object) {
    return this.query(publicKey, "mintNFT", { metadata });
  }
} 