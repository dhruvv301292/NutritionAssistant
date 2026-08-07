import { useCallback, useState } from 'react';
import { useFocusEffect } from '@react-navigation/native';
import { ActivityIndicator, FlatList, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { dailySummary, getGoals } from '../api/client';
import CaloriesCard from '../components/CaloriesCard';
import MacroTile from '../components/MacroTile';
import NutritionTags from '../components/NutritionTags';
import { colors, fonts, radius, shadow } from '../theme';
import { dateKey, formatTime, last14Days } from '../dateUtils';
import { SLOT_LABEL, SLOT_ORDER, slotForTime } from '../slots';
import type { DailySummary, Goals, MealLog, Nutrition } from '../types/api';

export default function HistoryScreen() {
  const [selectedDate, setSelectedDate] = useState(() => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    return dateKey(yesterday);
  });
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [goals, setGoals] = useState<Goals | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback((date: string) => {
    setError(null);
    setSummary(null);
    Promise.all([dailySummary(date), getGoals()])
      .then(([s, g]) => {
        setSummary(s);
        setGoals(g);
      })
      .catch(() => setError('Could not load that day.'));
  }, []);

  useFocusEffect(useCallback(() => { load(selectedDate); }, [load, selectedDate]));

  const dates = last14Days(new Date());

  const mealsBySlot = groupBySlot(summary?.meals ?? []);

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Text style={styles.title}>History</Text>

      <FlatList
        horizontal
        data={dates}
        keyExtractor={(d) => dateKey(d)}
        showsHorizontalScrollIndicator={false}
        style={styles.dateStrip}
        contentContainerStyle={{ gap: 8, paddingHorizontal: 20 }}
        initialScrollIndex={dates.length - 1}
        getItemLayout={(_, index) => ({ length: 60, offset: 60 * index, index })}
        renderItem={({ item }) => {
          const key = dateKey(item);
          const isSelected = key === selectedDate;
          return (
            <Pressable
              onPress={() => setSelectedDate(key)}
              style={[styles.dateChip, { backgroundColor: isSelected ? colors.accent : colors.surface }]}
            >
              <Text style={[styles.dateWeekday, { color: isSelected ? colors.bg : colors.text }]}>
                {item.toLocaleDateString(undefined, { weekday: 'short' }).slice(0, 2)}
              </Text>
              <Text style={[styles.dateNum, { color: isSelected ? colors.bg : colors.text }]}>{item.getDate()}</Text>
            </Pressable>
          );
        }}
      />

      <ScrollView contentContainerStyle={styles.scrollContent}>
        {error && <Text style={styles.errorText}>{error}</Text>}
        {!error && (!summary || !goals) && <ActivityIndicator color={colors.accent700} />}
        {summary && goals && (
          <>
            <CaloriesCard calories={Math.round(summary.total.calories)} goal={goals.calorie_goal} />
            <View style={styles.tileGrid}>
              <MacroTile label="PROTEIN" value={Math.round(summary.total.protein)} goal={goals.protein_goal} color={colors.accent700} />
              <MacroTile label="CARBS" value={Math.round(summary.total.carbs)} goal={goals.carb_goal} color={colors.accent2_500} />
            </View>
            <View style={[styles.tileGrid, { marginBottom: 18 }]}>
              <MacroTile label="FAT" value={Math.round(summary.total.fat)} goal={goals.fat_goal} color={colors.neutral700} />
              <MacroTile label="FIBER" value={Math.round(summary.total.fiber)} goal={goals.fiber_goal} color={colors.accent2_700} />
            </View>

            {SLOT_ORDER.map((slot) => (
              <View key={slot} style={{ marginBottom: 18 }}>
                <Text style={styles.kicker}>{SLOT_LABEL[slot]}</Text>
                {(mealsBySlot[slot] ?? []).length === 0 && (
                  <Text style={styles.emptyText}>Nothing logged yet.</Text>
                )}
                {(mealsBySlot[slot] ?? []).map((meal) => (
                  <View key={meal.id} style={[styles.mealCard, shadow.sm]}>
                    <View style={styles.mealHeaderRow}>
                      <Text style={styles.mealName}>
                        {meal.items.map((i) => i.food?.name ?? `food #${i.food_id}`).join(', ')}
                      </Text>
                      <Text style={styles.mealTime}>{formatTime(meal.logged_at)}</Text>
                    </View>
                    <NutritionTags nutrition={sumMealNutrition(meal)} />
                  </View>
                ))}
              </View>
            ))}
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

function sumMealNutrition(meal: MealLog): Nutrition {
  const total: Nutrition = { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0, sodium: 0 };
  for (const item of meal.items) {
    if (!item.nutrition) continue;
    total.calories += item.nutrition.calories;
    total.protein += item.nutrition.protein;
    total.carbs += item.nutrition.carbs;
    total.fat += item.nutrition.fat;
    total.fiber += item.nutrition.fiber;
    total.sodium += item.nutrition.sodium;
  }
  return total;
}

function groupBySlot(meals: MealLog[]): Partial<Record<string, MealLog[]>> {
  const result: Partial<Record<string, MealLog[]>> = {};
  for (const meal of meals) {
    const slot = slotForTime(new Date(meal.logged_at));
    result[slot] = result[slot] ?? [];
    result[slot]!.push(meal);
  }
  return result;
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  title: { fontFamily: fonts.heading, fontSize: 30, color: colors.text, paddingHorizontal: 20, marginBottom: 10 },
  dateStrip: { flexGrow: 0, marginBottom: 16 },
  dateChip: { width: 52, alignItems: 'center', gap: 5, paddingVertical: 10, borderRadius: radius.lg },
  dateWeekday: { fontFamily: fonts.bodySemiBold, fontSize: 13, textTransform: 'uppercase', opacity: 0.75 },
  dateNum: { fontFamily: fonts.heading, fontSize: 23 },
  scrollContent: { paddingHorizontal: 20, paddingBottom: 24 },
  kicker: { fontFamily: fonts.bodySemiBold, fontSize: 10, letterSpacing: 1, color: colors.accent, marginBottom: 8, textTransform: 'uppercase' },
  tileGrid: { flexDirection: 'row', gap: 12, marginBottom: 12 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14 },
  emptyText: { fontFamily: fonts.body, fontSize: 13, color: colors.text, opacity: 0.5 },
  mealCard: { backgroundColor: colors.surface, borderRadius: 26, padding: 14, marginBottom: 10, gap: 8 },
  mealHeaderRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  mealName: { fontFamily: fonts.heading, fontSize: 15, color: colors.text, flex: 1, marginRight: 8 },
  mealTime: { fontFamily: fonts.body, fontSize: 12, color: colors.text, opacity: 0.5 },
});
