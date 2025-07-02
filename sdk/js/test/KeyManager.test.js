"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
// @jest-environment node
const KeyManager_1 = require("../src/KeyManager");
describe('KeyManager', () => {
    it('generates a valid mnemonic', async () => {
        const mnemonic = await KeyManager_1.KeyManager.createMnemonic();
        expect(mnemonic.split(' ').length).toBe(24);
    });
});
