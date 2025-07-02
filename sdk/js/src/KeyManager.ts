import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";

export class KeyManager {
  static async createMnemonic(): Promise<string> {
    const wallet = await DirectSecp256k1HdWallet.generate(24);
    return wallet.mnemonic;
  }

  static async getWallet(mnemonic: string): Promise<DirectSecp256k1HdWallet> {
    return DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
      prefix: "fluentum", // Your chain's address prefix
    });
  }
} 