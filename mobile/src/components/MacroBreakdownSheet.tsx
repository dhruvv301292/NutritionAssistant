import { Modal, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { colors, fonts, radius, shadow } from '../theme';
import { formatTime } from '../dateUtils';
import type { MealLog, Nutrition } from '../types/api';

type MacroKey = keyof Nutrition;

type Props = {
  visible: boolean;
  onClose: () => void;
  label: string;
  macroKey: MacroKey;
  unit: string;
  total: number;
  color: string;
  meals: MealLog[];
};

function mealMacroTotal(meal: MealLog, macroKey: MacroKey): number {
  return meal.items.reduce((sum, item) => sum + (item.nutrition?.[macroKey] ?? 0), 0);
}

function mealName(meal: MealLog): string {
  return meal.items.map((i) => i.food?.name ?? `food #${i.food_id}`).join(', ');
}

export default function MacroBreakdownSheet({ visible, onClose, label, macroKey, unit, total, color, meals }: Props) {
  const rows = meals
    .map((meal) => ({ meal, value: mealMacroTotal(meal, macroKey) }))
    .filter((row) => row.value > 0)
    .sort((a, b) => b.value - a.value);

  return (
    <Modal visible={visible} animationType="slide" transparent onRequestClose={onClose}>
      <Pressable style={styles.backdrop} onPress={onClose} />
      <View style={styles.sheet}>
        <View style={styles.header}>
          <Text style={styles.title}>{label} breakdown</Text>
          <Pressable onPress={onClose} style={styles.closeButton}>
            <Text style={styles.closeButtonText}>✕</Text>
          </Pressable>
        </View>

        <ScrollView contentContainerStyle={{ gap: 8, paddingBottom: 10 }}>
          {rows.length === 0 && <Text style={styles.emptyText}>Nothing logged yet.</Text>}
          {rows.map(({ meal, value }) => (
            <View key={meal.id} style={styles.row}>
              <View style={{ flex: 1 }}>
                <Text style={styles.mealName}>{mealName(meal)}</Text>
                <Text style={styles.mealTime}>{formatTime(meal.logged_at)}</Text>
              </View>
              <Text style={[styles.mealValue, { color }]}>
                {Math.round(value)}
                {unit}
              </Text>
            </View>
          ))}
        </ScrollView>

        <View style={styles.totalRow}>
          <Text style={styles.totalLabel}>Total</Text>
          <Text style={styles.totalValue}>
            {Math.round(total)}
            {unit}
          </Text>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: { flex: 1, backgroundColor: 'rgba(32,30,29,0.4)' },
  sheet: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: 0,
    maxHeight: '70%',
    backgroundColor: colors.bg,
    borderTopLeftRadius: 28,
    borderTopRightRadius: 28,
    padding: 18,
    ...shadow.lg,
  },
  header: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 },
  title: { fontFamily: fonts.heading, fontSize: 17, color: colors.text },
  closeButton: {
    width: 30,
    height: 30,
    borderRadius: radius.pill,
    backgroundColor: colors.neutral200,
    alignItems: 'center',
    justifyContent: 'center',
  },
  closeButtonText: { fontFamily: fonts.body, fontSize: 14, color: colors.text },
  emptyText: { fontFamily: fonts.body, fontSize: 13, color: colors.text, opacity: 0.5 },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: colors.surface,
    borderRadius: 18,
    padding: 12,
  },
  mealName: { fontFamily: fonts.bodySemiBold, fontSize: 14, color: colors.text },
  mealTime: { fontFamily: fonts.body, fontSize: 11.5, color: colors.text, opacity: 0.5, marginTop: 2 },
  mealValue: { fontFamily: fonts.heading, fontSize: 16 },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.divider,
    marginTop: 10,
    paddingTop: 10,
  },
  totalLabel: { fontFamily: fonts.bodySemiBold, fontSize: 13, color: colors.text, opacity: 0.7 },
  totalValue: { fontFamily: fonts.bodyBold, fontSize: 13, color: colors.text },
});
