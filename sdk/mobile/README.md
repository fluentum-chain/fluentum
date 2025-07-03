# Fluentum Mobile Toolkit

A React Native toolkit for building Fluentum dApps with wallet connection, transaction signing, and blockchain queries.

## Features

- Deep linking wallet adapters (Phantom, Backpack)
- Universal wallet connection & transaction signing
- Pre-built swap and NFT templates
- Hooks-based API for easy integration
- Mobile-first, touch-optimized components
- TypeScript support

## Installation

### Quick Install (Recommended)

Run the provided setup script to install all dependencies:

```bash
cd sdk/mobile
./scripts/setup.sh
```

### Manual Install

```bash
npm install fluentum-mobile-toolkit
```

## Quick Start

```tsx
import { useFluentumWallet } from 'fluentum-mobile-toolkit';

const { publicKey, balance, nfts, connect } = useFluentumWallet();
```

## Example Usage

```tsx
import React from 'react';
import { SafeAreaView, ScrollView, Text } from 'react-native';
import { useFluentumWallet, WalletConnectButton, NFTGallery, SwapTemplate } from 'fluentum-mobile-toolkit';

export default function App() {
  const { publicKey, balance, nfts, connect } = useFluentumWallet();

  return (
    <SafeAreaView style={{ flex: 1 }}>
      <ScrollView contentContainerStyle={{ padding: 16 }}>
        {publicKey ? (
          <>
            <Text>Balance: {balance} FLT</Text>
            <SwapTemplate publicKey={publicKey} />
            <Text>Your NFTs:</Text>
            <NFTGallery nfts={nfts} />
          </>
        ) : (
          <>
            <Text>Connect to Fluentum</Text>
            <WalletConnectButton onConnect={connect} />
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}
```

## Deep Linking Setup

**Android:**  
Add to `android/app/src/main/AndroidManifest.xml`:
```xml
<intent-filter>
  <action android:name="android.intent.action.VIEW" />
  <category android:name="android.intent.category.DEFAULT" />
  <category android:name="android.intent.category.BROWSABLE" />
  <data android:scheme="flt" />
</intent-filter>
```

**iOS:**  
Add to `ios/FluentumMobileToolkit/Info.plist`:
```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleTypeRole</key>
    <string>Editor</string>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>flt</string>
    </array>
  </dict>
</array>
```

## License

MIT © Fluentum Technologies 