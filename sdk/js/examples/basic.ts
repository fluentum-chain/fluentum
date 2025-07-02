import { FluentumClient } from '../src/Client';
import { KeyManager } from '../src/KeyManager';
import { TransactionHandler } from '../src/Transaction';

async function main() {
  const endpoint = 'https://rpc.fluentum.io';
  const mnemonic = await KeyManager.createMnemonic();
  const wallet = await KeyManager.getWallet(mnemonic);
  const client = await FluentumClient.connect(endpoint);

  // Example send tokens (replace with real addresses)
  // await TransactionHandler.sendTokens(
  //   'fluentum1sender...',
  //   'fluentum1recipient...',
  //   '1000000',
  //   wallet,
  //   endpoint
  // );

  console.log('Wallet mnemonic:', mnemonic);
}

main(); 