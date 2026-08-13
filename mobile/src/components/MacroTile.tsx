import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors, fonts, radius, shadow } from '../theme';

type Props = {
  label: string;
  value: number;
  goal: number;
  unit?: string;
  color: string;
  onPress?: () => void;
};

export default function MacroTile({ label, value, goal, unit = 'g', color, onPress }: Props) {
  const pct = Math.max(4, Math.min(100, Math.round((value / goal) * 100)));
  return (
    <Pressable onPress={onPress} style={[styles.tile, shadow.sm]}>
      <Text style={styles.label}>{label}</Text>
      <Text style={styles.value}>
        {value}
        {unit}
        <Text style={styles.goal}> / {goal}{unit}</Text>
      </Text>
      <View style={styles.track}>
        <View style={[styles.fill, { width: `${pct}%`, backgroundColor: color }]} />
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  tile: {
    flexBasis: '47%',
    flexGrow: 1,
    backgroundColor: colors.surface,
    borderRadius: 26,
    padding: 14,
    gap: 8,
  },
  label: { fontFamily: fonts.bodySemiBold, fontSize: 10, letterSpacing: 1, color: colors.text, opacity: 0.6 },
  value: { fontFamily: fonts.heading, fontSize: 30, color: colors.text },
  goal: { fontFamily: fonts.body, fontSize: 14, opacity: 0.55 },
  track: { height: 6, borderRadius: radius.pill, backgroundColor: colors.neutral200, overflow: 'hidden' },
  fill: { height: '100%', borderRadius: radius.pill },
});
