import { StyleSheet, Text, View } from 'react-native';
import { colors, fonts, radius } from '../theme';
import type { Nutrition, TrackedMacro } from '../types/api';

const TAG_DEFS: { key: TrackedMacro; label: string; bg: string; color: string; outline?: boolean; decimals: number }[] = [
  { key: 'protein', label: 'P', bg: 'transparent', color: colors.accent, outline: true, decimals: 1 },
  { key: 'carbs', label: 'C', bg: colors.accent2_100, color: colors.accent2_800, decimals: 1 },
  { key: 'fat', label: 'F', bg: colors.neutral100, color: colors.neutral800, decimals: 1 },
  { key: 'fiber', label: 'Fi', bg: colors.accent2_100, color: colors.accent2_800, decimals: 1 },
  { key: 'sodium', label: 'Na', bg: colors.neutral100, color: colors.neutral800, decimals: 0 },
];

type Props = {
  nutrition: Nutrition;
  // Which macro tags to show alongside calories (always shown). Omit to
  // show the pre-tracked-macros default set (protein/carbs/fat), so
  // existing callers that haven't been updated yet don't break.
  trackedMacros?: TrackedMacro[];
};

export default function NutritionTags({ nutrition, trackedMacros = ['protein', 'carbs', 'fat'] }: Props) {
  return (
    <View style={styles.row}>
      <Tag text={`${Math.round(nutrition.calories)} kcal`} bg={colors.accent100} color={colors.accent800} />
      {TAG_DEFS.filter((t) => trackedMacros.includes(t.key)).map((t) => (
        <Tag
          key={t.key}
          text={`${t.label} ${nutrition[t.key].toFixed(t.decimals)}${t.key === 'sodium' ? 'mg' : 'g'}`}
          bg={t.bg}
          color={t.color}
          outline={t.outline}
        />
      ))}
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
