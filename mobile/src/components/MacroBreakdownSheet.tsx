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

function itemName(item: MealLog['items'][number]): string {
  return item.food?.name ?? `food #${item.food_id}`;
}

type ItemRow = { key: string; name: string; time: string; value: number };

function itemRows(meals: MealLog[], macroKey: MacroKey): ItemRow[] {
  const rows: ItemRow[] = [];
  for (const meal of meals) {
    for (const item of meal.items) {
      const value = item.nutrition?.[macroKey] ?? 0;
      if (value <= 0) continue;
      rows.push({ key: `${meal.id}-${item.id}`, name: itemName(item), time: formatTime(meal.logged_at), value });
    }
  }
  return rows.sort((a, b) => b.value - a.value);
}

export default function MacroBreakdownSheet({ visible, onClose, label, macroKey, unit, total, color, meals }: Props) {
  const rows = itemRows(meals, macroKey);

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
          {rows.map((row) => (
            <View key={row.key} style={styles.row}>
              <View style={{ flex: 1 }}>
                <Text style={styles.mealName}>{row.name}</Text>
                <Text style={styles.mealTime}>{row.time}</Text>
              </View>
              <Text style={[styles.mealValue, { color }]}>
                {Math.round(row.value)}
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
