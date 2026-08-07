import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors, fonts, radius, shadow } from '../theme';

type Props = {
  calories: number;
  goal: number;
  onPress?: () => void;
};

export default function CaloriesCard({ calories, goal, onPress }: Props) {
  const remaining = goal - calories;
  const remainingText = remaining >= 0 ? `${remaining} kcal left` : `${Math.abs(remaining)} kcal over`;
  const pct = Math.max(4, Math.min(100, Math.round((calories / goal) * 100)));

  return (
    <Pressable onPress={onPress} style={[styles.card, shadow.sm]}>
      <View style={styles.headerRow}>
        <Text style={styles.kicker}>CALORIES</Text>
        <Text style={styles.remaining}>{remainingText}</Text>
      </View>
      <View style={styles.valueRow}>
        <Text style={styles.value}>{calories}</Text>
        <Text style={styles.goal}> / {goal} kcal</Text>
      </View>
      <View style={styles.track}>
        <View style={[styles.fill, { width: `${pct}%` }]} />
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderRadius: 32,
    padding: 18,
    marginBottom: 12,
  },
  headerRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  kicker: { fontFamily: fonts.bodySemiBold, fontSize: 10, letterSpacing: 1, color: colors.accent },
  remaining: { fontFamily: fonts.body, fontSize: 12, color: colors.accent700 },
  valueRow: { flexDirection: 'row', alignItems: 'baseline', marginVertical: 10, gap: 6 },
  value: { fontFamily: fonts.heading, fontSize: 44, color: colors.text },
  goal: { fontFamily: fonts.body, fontSize: 16, color: colors.neutral800, opacity: 0.65 },
  track: { height: 10, borderRadius: radius.pill, backgroundColor: colors.neutral200, overflow: 'hidden' },
  fill: { height: '100%', borderRadius: radius.pill, backgroundColor: colors.accent500 },
});
