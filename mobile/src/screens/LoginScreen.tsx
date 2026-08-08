import { useEffect, useState } from 'react';
import { ActivityIndicator, Platform, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import * as AppleAuthentication from 'expo-apple-authentication';
import { useAuth } from '../auth/AuthContext';
import { useGoogleSignIn } from '../auth/googleSignIn';
import { colors, fonts, radius, shadow } from '../theme';

export default function LoginScreen() {
  const { signInWithGoogleIdToken, signInWithAppleIdToken } = useAuth();
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

  async function handleApplePress() {
    setError(null);
    try {
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });
      if (!credential.identityToken) {
        setError('Could not sign in. Try again.');
        return;
      }
      setSigningIn(true);
      // Apple only includes the name in this initial credential, on the
      // user's very first-ever sign-in — never again after that.
      const name = [credential.fullName?.givenName, credential.fullName?.familyName]
        .filter(Boolean)
        .join(' ');
      await signInWithAppleIdToken(credential.identityToken, name);
    } catch (err: any) {
      if (err?.code === 'ERR_REQUEST_CANCELED') return;
      setError('Could not sign in. Try again.');
    } finally {
      setSigningIn(false);
    }
  }

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

        {Platform.OS === 'ios' && (
          <AppleAuthentication.AppleAuthenticationButton
            buttonType={AppleAuthentication.AppleAuthenticationButtonType.CONTINUE}
            buttonStyle={AppleAuthentication.AppleAuthenticationButtonStyle.BLACK}
            cornerRadius={radius.pill}
            style={styles.appleButton}
            onPress={handleApplePress}
          />
        )}

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
  appleButton: { width: 220, height: 48, marginTop: 10 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 13, marginTop: 12 },
});
