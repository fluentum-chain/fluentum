import React from 'react';
import { SafeAreaView, ScrollView, Text } from 'react-native';
import { useFluentumWallet } from './src/hooks/useFluentumWallet';
import { WalletConnectButton } from './src/components/WalletConnectButton';
import { NFTGallery } from './src/components/NFTGallery';
import { SwapTemplate } from './src/templates/SwapTemplate';

export default function App() {
  const { publicKey, balance, nfts, connect } = useFluentumWallet();

  return (
    <SafeAreaView style={{ flex: 1 }}>
      <ScrollView contentContainerStyle={{ padding: 16 }}>
        {publicKey ? (
          <>
            <Text style={{ fontSize: 18, marginBottom: 16 }}>
              Balance: {balance} FLT
            </Text>
            <SwapTemplate publicKey={publicKey} />
            <Text style={{ fontSize: 18, marginVertical: 16 }}>
              Your NFTs:
            </Text>
            <NFTGallery nfts={nfts} />
          </>
        ) : (
          <>
            <Text style={{ fontSize: 24, textAlign: 'center', marginBottom: 32 }}>
              Connect to Fluentum
            </Text>
            <WalletConnectButton onConnect={connect} />
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
} 