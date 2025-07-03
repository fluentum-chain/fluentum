import React, { useState } from 'react';
import { View, TextInput, Button, StyleSheet, Text } from 'react-native';
import { FluentumRPC } from '../services/FluentumRPC';

export const SwapTemplate = ({ publicKey }: { publicKey: string }) => {
  const [fromToken, setFromToken] = useState('FLT');
  const [toToken, setToToken] = useState('USDC');
  const [amount, setAmount] = useState('');

  const executeSwap = async () => {
    await FluentumRPC.executeSwap(
      publicKey, 
      fromToken, 
      toToken, 
      parseFloat(amount)
    );
  };

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.input}
        value={amount}
        onChangeText={setAmount}
        placeholder="Amount"
        keyboardType="numeric"
      />
      <View style={styles.tokenRow}>
        <TextInput
          style={[styles.input, styles.tokenInput]}
          value={fromToken}
          onChangeText={setFromToken}
        />
        <Text>→</Text>
        <TextInput
          style={[styles.input, styles.tokenInput]}
          value={toToken}
          onChangeText={setToToken}
        />
      </View>
      <Button title="Execute Swap" onPress={executeSwap} />
    </View>
  );
};

const styles = StyleSheet.create({
  container: { padding: 16 },
  input: { 
    borderWidth: 1, 
    padding: 12, 
    marginBottom: 16,
    borderRadius: 8
  },
  tokenInput: { flex: 1 },
  tokenRow: { 
    flexDirection: 'row', 
    alignItems: 'center', 
    gap: 8 
  }
}); 