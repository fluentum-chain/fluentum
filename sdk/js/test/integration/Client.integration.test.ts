// @jest-environment node
import { FluentumClient } from '../../src/Client';
import { EventSubscriber } from '../../src/events';

describe('FluentumClient Integration', () => {
  const endpoint = process.env.FLUENTUM_RPC || '';
  const isWebSocket = endpoint.startsWith('ws://') || endpoint.startsWith('wss://');

  (endpoint && isWebSocket ? it : it.skip)('connects to the chain and subscribes to new blocks', (done: jest.DoneCallback) => {
    (async () => {
      const client = await FluentumClient.connect(endpoint);
      expect(client).toBeDefined();

      const events = await EventSubscriber.connect(endpoint);
      events.subscribeNewBlocks((block) => {
        expect(block).toBeDefined();
        done();
      });
    })();
  });
}); 