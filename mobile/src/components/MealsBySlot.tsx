import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
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

function ingredientList(meal: MealLog): string {
  return meal.items.map((i) => i.food?.name ?? `food #${i.food_id}`).join(', ');
}

function itemQuantityLabel(item: MealLog['items'][number]): string {
  const name = item.food?.name ?? `food #${item.food_id}`;
  const qty = Number.isInteger(item.quantity) ? item.quantity : item.quantity.toFixed(1);
  return `${qty} ${item.unit} ${name}`;
}

function MealCard({ meal, trackedMacros }: { meal: MealLog; trackedMacros: TrackedMacro[] }) {
  const [expanded, setExpanded] = useState(false);
  const heading = meal.title || ingredientList(meal);

  return (
    <Pressable style={[styles.mealCard, shadow.sm]} onPress={() => setExpanded((e) => !e)}>
      <View style={styles.mealHeaderRow}>
        <Text style={styles.mealName}>{heading}</Text>
        <Text style={styles.mealTime}>{formatTime(meal.logged_at)}</Text>
      </View>
      {expanded && (
        <View style={styles.mealItemsList}>
          {meal.items.map((item) => (
            <Text key={item.id} style={styles.mealIngredients}>
              {itemQuantityLabel(item)}
            </Text>
          ))}
        </View>
      )}
      <NutritionTags nutrition={sumMealNutrition(meal)} trackedMacros={trackedMacros} />
    </Pressable>
  );
}

// Only slots with at least one logged meal are shown — an empty "Dinner"
// section before dinner has happened yet is just noise.
export default function MealsBySlot({ meals, trackedMacros }: { meals: MealLog[]; trackedMacros: TrackedMacro[] }) {
  const mealsBySlot = groupBySlot(meals);
  const slotsWithMeals = SLOT_ORDER.filter((slot) => (mealsBySlot[slot] ?? []).length > 0);

  return (
    <>
      {slotsWithMeals.map((slot, i) => (
        <View key={slot} style={{ marginTop: i === 0 ? 26 : 6 }}>
          <Text style={styles.kicker}>{SLOT_LABEL[slot]}</Text>
          {(mealsBySlot[slot] ?? []).map((meal) => (
            <MealCard key={meal.id} meal={meal} trackedMacros={trackedMacros} />
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
  mealName: { fontFamily: fonts.bodyBold, fontSize: 15, color: colors.text, flex: 1, marginRight: 8 },
  mealItemsList: { gap: 2 },
  mealIngredients: { fontFamily: fonts.body, fontSize: 12.5, color: colors.text, opacity: 0.6 },
  mealTime: { fontFamily: fonts.body, fontSize: 12, color: colors.text, opacity: 0.5 },
});
