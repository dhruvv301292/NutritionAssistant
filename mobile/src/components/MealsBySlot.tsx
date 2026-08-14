import { StyleSheet, Text, View } from 'react-native';
import { colors, fonts, shadow } from '../theme';
import { formatTime } from '../dateUtils';
import { SLOT_LABEL, SLOT_ORDER } from '../slots';
import NutritionTags from './NutritionTags';
import type { MealLog, Nutrition, TrackedMacro } from '../types/api';

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
    const slot = meal.slot;
    result[slot] = result[slot] ?? [];
    result[slot]!.push(meal);
  }
  return result;
}

// Only slots with at least one logged meal are shown — an empty "Dinner"
// section before dinner has happened yet is just noise.
export default function MealsBySlot({ meals, trackedMacros }: { meals: MealLog[]; trackedMacros: TrackedMacro[] }) {
  const mealsBySlot = groupBySlot(meals);
  const slotsWithMeals = SLOT_ORDER.filter((slot) => (mealsBySlot[slot] ?? []).length > 0);

  return (
    <>
      {slotsWithMeals.map((slot) => (
        <View key={slot} style={{ marginBottom: 18 }}>
          <Text style={styles.kicker}>{SLOT_LABEL[slot]}</Text>
          {(mealsBySlot[slot] ?? []).map((meal) => (
            <View key={meal.id} style={[styles.mealCard, shadow.sm]}>
              <View style={styles.mealHeaderRow}>
                <Text style={styles.mealName}>
                  {meal.items.map((i) => i.food?.name ?? `food #${i.food_id}`).join(', ')}
                </Text>
                <Text style={styles.mealTime}>{formatTime(meal.logged_at)}</Text>
              </View>
              <NutritionTags nutrition={sumMealNutrition(meal)} trackedMacros={trackedMacros} />
            </View>
          ))}
        </View>
      ))}
    </>
  );
}

const styles = StyleSheet.create({
  kicker: { fontFamily: fonts.bodySemiBold, fontSize: 10, letterSpacing: 1, color: colors.accent, marginBottom: 8, textTransform: 'uppercase' },
  mealCard: { backgroundColor: colors.surface, borderRadius: 26, padding: 14, marginBottom: 10, gap: 8 },
  mealHeaderRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  mealName: { fontFamily: fonts.heading, fontSize: 15, color: colors.text, flex: 1, marginRight: 8 },
  mealTime: { fontFamily: fonts.body, fontSize: 12, color: colors.text, opacity: 0.5 },
});
