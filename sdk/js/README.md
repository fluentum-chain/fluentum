# Fluentum TypeScript SDK

This SDK provides tools for interacting with the Fluentum blockchain in Node.js and browser environments.

## Getting Started

```bash
cd sdk/js
npm install
npm run proto:gen
npm run build
```

## Features
- Account/key management
- Transaction creation, signing, and broadcasting
- Querying blockchain state
- Event subscription
- Gas fee estimation

## Directory Structure
- `src/` — SDK source code
- `proto/` — Symlink to shared protobuf definitions
- `test/` — Unit/integration tests
- `examples/` — Usage examples

## Publishing
Only the SDK code is published to npm. Go code and other chain files are excluded.

## Integration Testing

To run integration tests (requires a running Fluentum RPC endpoint):

```bash
export FLUENTUM_RPC=https://rpc.fluentum.io # or your local node
npm test
```

## Event Subscription Example

```typescript
import { EventSubscriber } from './src/events';

(async () => {
  const events = await EventSubscriber.connect('https://rpc.fluentum.io');
  events.subscribeNewBlocks((block) => {
    console.log('New block:', block);
  });
})();
``` 