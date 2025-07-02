import { SigningStargateClient } from "@cosmjs/stargate";

export class TransactionHandler {
  static async sendTokens(
    sender: string,
    recipient: string,
    amount: string,
    signer: any, // Wallet adapter
    rpcEndpoint: string
  ) {
    const signingClient = await SigningStargateClient.connectWithSigner(
      rpcEndpoint,
      signer
    );
    return signingClient.sendTokens(
      sender,
      recipient,
      [{ denom: "uflt", amount }],
      "auto" // Gas
    );
  }
} 