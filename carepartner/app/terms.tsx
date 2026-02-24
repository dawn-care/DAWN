import React from "react";
import { SafeAreaView, StyleSheet, Text, View } from "react-native";
import { Stack } from "expo-router";

export default function Terms() {
  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />
      <SafeAreaView style={styles.safe}>
        <View style={styles.container}>
          <Text style={styles.title}>Terms & Conditions (Stub)</Text>
          <Text style={styles.text}>
            This screen will display the Terms and an Accept button.
          </Text>
        </View>
      </SafeAreaView>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#fff" },
  container: { flex: 1, alignItems: "center", justifyContent: "center", padding: 18 },
  title: { fontSize: 24, fontWeight: "800", marginBottom: 10 },
  text: { fontSize: 16, color: "rgba(0,0,0,0.65)", textAlign: "center" },
});