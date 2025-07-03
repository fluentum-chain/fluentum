import { Linking } from 'react-native';

type WalletConfig = {
  name: string;
  deeplink: string;
  installLink: string;
};

const WALLETS: Record<string, WalletConfig> = {
  phantom: {
    name: "Phantom",
    deeplink: "phantom://connect?app=fluentum&redirect=flt://wallet-auth",
    installLink: "https://phantom.app/download"
  },
  backpack: {
    name: "Backpack",
    deeplink: "backpack://connect?network=fluentum&callback=flt://",
    installLink: "https://backpack.app/download"
  }
};

export class WalletAdapter {
  static async connect(walletName: string): Promise<string> {
    const wallet = WALLETS[walletName];
    try {
      await Linking.openURL(wallet.deeplink);
      return new Promise((resolve) => {
        Linking.addEventListener('url', (event) => {
          const url = new URL(event.url);
          resolve(url.searchParams.get('publicKey') || '');
        });
      });
    } catch (e) {
      Linking.openURL(wallet.installLink);
      throw new Error(`${wallet.name} not installed`);
    }
  }

  static async signTransaction(tx: string, walletName: string): Promise<string> {
    const wallet = WALLETS[walletName];
    const signUrl = `${wallet.deeplink}/sign?tx=${encodeURIComponent(tx)}`;
    await Linking.openURL(signUrl);
    return new Promise((resolve) => {
      Linking.addEventListener('url', (event) => {
        const url = new URL(event.url);
        resolve(url.searchParams.get('signedTx') || '');
      });
    });
  }
} 