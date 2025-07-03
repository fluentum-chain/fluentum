import React from 'react';
import { Button, View, StyleSheet } from 'react-native';

export const WalletConnectButton = ({ onConnect }: { onConnect: (wallet: string) => void }) => (
  <View style={styles.container}>
    <Button 
      title="Connect Phantom" 
      onPress={() => onConnect('phantom')} 
      color="#4A22B0" 
    />
    <Button 
      title="Connect Backpack" 
      onPress={() => onConnect('backpack')} 
      color="#1A1A1A" 
    />
  </View>
);

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    padding: 16
  }
}); 