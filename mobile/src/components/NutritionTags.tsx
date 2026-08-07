import { StyleSheet, Text, View } from 'react-native';
import { colors, fonts, radius } from '../theme';
import type { Nutrition } from '../types/api';

export default function NutritionTags({ nutrition }: { nutrition: Nutrition }) {
  return (
    <View style={styles.row}>
      <Tag text={`${Math.round(nutrition.calories)} kcal`} bg={colors.accent100} color={colors.accent800} />
      <Tag text={`P ${nutrition.protein.toFixed(1)}g`} bg="transparent" color={colors.accent} outline />
      <Tag text={`C ${nutrition.carbs.toFixed(1)}g`} bg={colors.accent2_100} color={colors.accent2_800} />
      <Tag text={`F ${nutrition.fat.toFixed(1)}g`} bg={colors.neutral100} color={colors.neutral800} />
    </View>
  );
}

function Tag({ text, bg, color, outline }: { text: string; bg: string; color: string; outline?: boolean }) {
  return (
    <View style={[styles.tag, { backgroundColor: bg }, outline && { borderWidth: 1, borderColor: colors.accent }]}>
      <Text style={[styles.tagText, { color }]}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', gap: 6, flexWrap: 'wrap' },
  tag: { paddingHorizontal: 10, paddingVertical: 3, borderRadius: radius.md * 0.75 },
  tagText: { fontFamily: fonts.bodySemiBold, fontSize: 11, letterSpacing: 0.2 },
});
