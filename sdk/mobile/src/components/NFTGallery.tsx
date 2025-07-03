import React from 'react';
import { FlatList, Image, Text, View, StyleSheet } from 'react-native';

export const NFTGallery = ({ nfts }: { nfts: Array<any> }) => (
  <FlatList
    data={nfts}
    numColumns={2}
    renderItem={({ item }) => (
      <View style={styles.card}>
        <Image source={{ uri: item.image }} style={styles.image} />
        <Text style={styles.title}>{item.name}</Text>
      </View>
    )}
    keyExtractor={item => item.id}
  />
);

const styles = StyleSheet.create({
  card: { margin: 8, width: 160 },
  image: { width: 160, height: 160, borderRadius: 8 },
  title: { marginTop: 4, textAlign: 'center' }
}); 