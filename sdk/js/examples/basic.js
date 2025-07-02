"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const Client_1 = require("../src/Client");
const KeyManager_1 = require("../src/KeyManager");
async function main() {
    const endpoint = 'https://rpc.fluentum.io';
    const mnemonic = await KeyManager_1.KeyManager.createMnemonic();
    const wallet = await KeyManager_1.KeyManager.getWallet(mnemonic);
    const client = await Client_1.FluentumClient.connect(endpoint);
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
