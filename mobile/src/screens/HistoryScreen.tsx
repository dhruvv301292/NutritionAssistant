import { useCallback, useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import PagerView from 'react-native-pager-view';
import { SafeAreaView } from 'react-native-safe-area-context';
import { dailySummary, getGoals } from '../api/client';
import AccountButton from '../components/AccountButton';
import CaloriesCard from '../components/CaloriesCard';
import MacroBreakdownSheet from '../components/MacroBreakdownSheet';
import MacroTile from '../components/MacroTile';
import NutritionTags from '../components/NutritionTags';
import { MacroDef, macroGoal, trackedMacroDefs } from '../macros';
import { colors, fonts, radius, shadow } from '../theme';
import { dateKey, formatTime, startOfWeek, weekDates } from '../dateUtils';
import { SLOT_LABEL, SLOT_ORDER } from '../slots';
import type { DailySummary, Goals, MealLog, Nutrition } from '../types/api';

const WEEKDAY_LABELS = ['MO', 'TU', 'WE', 'TH', 'FR', 'SA', 'SU'];
const WEEK_PAGER_CENTER = 1;

export default function HistoryScreen({ focused }: { focused: boolean }) {
  const [selectedDate, setSelectedDate] = useState(() => dateKey(new Date()));
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()));
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [goals, setGoals] = useState<Goals | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activeMacro, setActiveMacro] = useState<MacroDef | null>(null);
  const weekPagerRef = useRef<PagerView>(null);

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

  useEffect(() => { if (focused) load(selectedDate); }, [focused, load, selectedDate]);

  const todayKey = dateKey(new Date());
  const currentWeekStart = startOfWeek(new Date());
  const isCurrentWeek = dateKey(weekStart) === dateKey(currentWeekStart);

  // The pager always shows [prev week, current week, next week] with the
  // current week centered. After a swipe settles on an edge page, shift
  // weekStart by 7 days and snap back to center without animating, so the
  // pager always has a page to swipe to in either direction. Swiping past
  // the current week is undone immediately — the "next" page is never a
  // real future week.
  function handlePageSelected(position: number) {
    if (position === WEEK_PAGER_CENTER) return;
    if (position > WEEK_PAGER_CENTER && isCurrentWeek) {
      requestAnimationFrame(() => weekPagerRef.current?.setPageWithoutAnimation(WEEK_PAGER_CENTER));
      return;
    }
    const delta = position === 0 ? -7 : 7;
    setWeekStart((prev) => {
      const next = new Date(prev);
      next.setDate(next.getDate() + delta);
      return next;
    });
    requestAnimationFrame(() => weekPagerRef.current?.setPageWithoutAnimation(WEEK_PAGER_CENTER));
  }

  const weeks = [-7, 0, 7].map((offset) => {
    const d = new Date(weekStart);
    d.setDate(d.getDate() + offset);
    return weekDates(d);
  });

  const mealsBySlot = groupBySlot(summary?.meals ?? []);

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.headerRow}>
        <Text style={styles.title}>History</Text>
        <AccountButton />
      </View>

      <View style={styles.weekdayRow}>
        {WEEKDAY_LABELS.map((label) => (
          <Text key={label} style={styles.weekdayLabel}>{label}</Text>
        ))}
      </View>

      <PagerView
        ref={weekPagerRef}
        style={styles.weekPager}
        initialPage={WEEK_PAGER_CENTER}
        onPageSelected={(e) => handlePageSelected(e.nativeEvent.position)}
      >
        {weeks.map((week, i) => (
          <View key={i} style={styles.dateRow}>
            {week.map((day) => {
              const key = dateKey(day);
              const isSelected = key === selectedDate;
              const isFuture = key > todayKey;
              return (
                <Pressable
                  key={key}
                  disabled={isFuture}
                  onPress={() => setSelectedDate(key)}
                  style={[
                    styles.dateChip,
                    { backgroundColor: isSelected ? colors.accent : colors.surface },
                    isFuture && styles.dateChipDisabled,
                  ]}
                >
                  <Text
                    style={[
                      styles.dateNum,
                      { color: isSelected ? colors.bg : colors.text },
                      isFuture && styles.dateNumDisabled,
                    ]}
                  >
                    {day.getDate()}
                  </Text>
                </Pressable>
              );
            })}
          </View>
        ))}
      </PagerView>

      <ScrollView contentContainerStyle={styles.scrollContent}>
        {error && <Text style={styles.errorText}>{error}</Text>}
        {!error && (!summary || !goals) && <ActivityIndicator color={colors.accent700} />}
        {summary && goals && (
          <>
            <CaloriesCard calories={Math.round(summary.total.calories)} goal={goals.calorie_goal} />
            <View style={[styles.tileGrid, { marginBottom: 18 }]}>
              {trackedMacroDefs(goals.tracked_macros).map((m) => (
                <MacroTile
                  key={m.key}
                  label={m.label}
                  value={Math.round(summary.total[m.key])}
                  goal={macroGoal(m, goals)}
                  unit={m.unit}
                  color={m.color}
                  onPress={() => setActiveMacro(m)}
                />
              ))}
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
                    <NutritionTags nutrition={sumMealNutrition(meal)} trackedMacros={goals.tracked_macros} />
                  </View>
                ))}
              </View>
            ))}
          </>
        )}
      </ScrollView>

      {summary && activeMacro && (
        <MacroBreakdownSheet
          visible={!!activeMacro}
          onClose={() => setActiveMacro(null)}
          label={activeMacro.label}
          macroKey={activeMacro.key}
          unit={activeMacro.unit}
          total={summary.total[activeMacro.key]}
          color={activeMacro.color}
          meals={summary.meals}
        />
      )}
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
    const slot = meal.slot;
    result[slot] = result[slot] ?? [];
    result[slot]!.push(meal);
  }
  return result;
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    marginBottom: 10,
  },
  title: { fontFamily: fonts.heading, fontSize: 30, color: colors.text },
  weekdayRow: { flexDirection: 'row', paddingHorizontal: 20, marginBottom: 8 },
  weekdayLabel: { flex: 1, textAlign: 'center', fontFamily: fonts.bodySemiBold, fontSize: 12, color: colors.text, opacity: 0.5, textTransform: 'uppercase' },
  weekPager: { height: 66, marginBottom: 16 },
  dateRow: { flexDirection: 'row', paddingHorizontal: 20, gap: 6 },
  dateChip: { flex: 1, alignItems: 'center', justifyContent: 'center', paddingVertical: 10, borderRadius: radius.lg },
  dateChipDisabled: { opacity: 0.35 },
  dateNum: { fontFamily: fonts.heading, fontSize: 23 },
  dateNumDisabled: { color: colors.neutral500 },
  scrollContent: { paddingHorizontal: 20, paddingBottom: 24 },
  kicker: { fontFamily: fonts.bodySemiBold, fontSize: 10, letterSpacing: 1, color: colors.accent, marginBottom: 8, textTransform: 'uppercase' },
  tileGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 12, marginBottom: 12 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14 },
  emptyText: { fontFamily: fonts.body, fontSize: 13, color: colors.text, opacity: 0.5 },
  mealCard: { backgroundColor: colors.surface, borderRadius: 26, padding: 14, marginBottom: 10, gap: 8 },
  mealHeaderRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' },
  mealName: { fontFamily: fonts.heading, fontSize: 15, color: colors.text, flex: 1, marginRight: 8 },
  mealTime: { fontFamily: fonts.body, fontSize: 12, color: colors.text, opacity: 0.5 },
});
