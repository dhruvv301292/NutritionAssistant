import { Alert, Pressable, StyleSheet } from 'react-native';
import { Feather } from '@expo/vector-icons';
import { useAuth } from '../auth/AuthContext';
import { colors, radius } from '../theme';

export default function AccountButton() {
  const { signOut } = useAuth();

  function handlePress() {
    Alert.alert('Sign out?', undefined, [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Sign out', style: 'destructive', onPress: signOut },
    ]);
  }

  return (
    <Pressable style={styles.button} onPress={handlePress}>
      <Feather name="user" size={17} color={colors.accent700} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    width: 36,
    height: 36,
    borderRadius: radius.pill,
    backgroundColor: colors.accent100,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
