// @jest-environment node
import { KeyManager } from '../src/KeyManager';

describe('KeyManager', () => {
  it('generates a valid mnemonic', async () => {
    const mnemonic = await KeyManager.createMnemonic();
    expect(mnemonic.split(' ').length).toBe(24);
  });
}); 