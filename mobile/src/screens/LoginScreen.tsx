import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuth } from '../auth/AuthContext';
import { useGoogleSignIn } from '../auth/googleSignIn';
import { colors, fonts, radius, shadow } from '../theme';

export default function LoginScreen() {
  const { signInWithGoogleIdToken } = useAuth();
  const [request, response, promptAsync] = useGoogleSignIn();
  const [signingIn, setSigningIn] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (response?.type !== 'success') return;
    const idToken = response.params.id_token;
    setSigningIn(true);
    setError(null);
    signInWithGoogleIdToken(idToken)
      .catch(() => setError('Could not sign in. Try again.'))
      .finally(() => setSigningIn(false));
  }, [response, signInWithGoogleIdToken]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>NutriChat</Text>
        <Text style={styles.subtitle}>Log meals in plain English, track your macros.</Text>

        <Pressable
          style={[styles.button, shadow.md]}
          disabled={!request || signingIn}
          onPress={() => promptAsync()}
        >
          {signingIn ? (
            <ActivityIndicator color={colors.bg} />
          ) : (
            <Text style={styles.buttonText}>Continue with Google</Text>
          )}
        </Pressable>

        {error && <Text style={styles.errorText}>{error}</Text>}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { flex: 1, justifyContent: 'center', alignItems: 'center', paddingHorizontal: 32, gap: 10 },
  title: { fontFamily: fonts.heading, fontSize: 36, color: colors.text, marginBottom: 4 },
  subtitle: { fontFamily: fonts.body, fontSize: 14, color: colors.text, opacity: 0.6, textAlign: 'center', marginBottom: 28 },
  button: {
    backgroundColor: colors.accent,
    borderRadius: radius.pill,
    paddingVertical: 14,
    paddingHorizontal: 28,
    alignItems: 'center',
    minWidth: 220,
  },
  buttonText: { color: colors.bg, fontFamily: fonts.heading, fontSize: 15 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 13, marginTop: 12 },
});
