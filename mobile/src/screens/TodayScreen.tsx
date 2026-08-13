import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Feather } from '@expo/vector-icons';
import { dailySummary, getGoals } from '../api/client';
import AccountButton from '../components/AccountButton';
import CaloriesCard from '../components/CaloriesCard';
import LogMealSheet from '../components/LogMealSheet';
import MacroBreakdownSheet from '../components/MacroBreakdownSheet';
import MacroTile from '../components/MacroTile';
import { MacroDef, macroGoal, trackedMacroDefs } from '../macros';
import { colors, fonts, radius, shadow } from '../theme';
import { dateKey, formatDateLabel } from '../dateUtils';
import type { DailySummary, Goals } from '../types/api';

export default function TodayScreen({ focused }: { focused: boolean }) {
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [goals, setGoals] = useState<Goals | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sheetOpen, setSheetOpen] = useState(false);
  const [activeMacro, setActiveMacro] = useState<MacroDef | null>(null);

  const load = useCallback(() => {
    setError(null);
    const today = dateKey(new Date());
    Promise.all([dailySummary(today), getGoals()])
      .then(([s, g]) => {
        setSummary(s);
        setGoals(g);
      })
      .catch(() => setError("Could not load today's data."));
  }, []);

  useEffect(() => { if (focused) load(); }, [focused, load]);

  const now = new Date();

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.headerRow}>
        <View>
          <Text style={styles.title}>Today</Text>
          <Text style={styles.dateLabel}>{formatDateLabel(now)}</Text>
        </View>
        <AccountButton />
      </View>

      <ScrollView contentContainerStyle={styles.scrollContent}>
        {error && <Text style={styles.errorText}>{error}</Text>}
        {!error && (!summary || !goals) && <ActivityIndicator color={colors.accent700} />}
        {summary && goals && (
          <>
            <CaloriesCard calories={Math.round(summary.total.calories)} goal={goals.calorie_goal} />
            <View style={styles.tileGrid}>
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
          </>
        )}
      </ScrollView>

      <Pressable onPress={() => setSheetOpen(true)} style={styles.promptBar}>
        <View style={styles.micIcon}>
          <Feather name="mic" size={16} color={colors.accent700} />
        </View>
        <Text style={styles.promptText}>Ready to log a meal?</Text>
      </Pressable>

      <LogMealSheet
        visible={sheetOpen}
        onClose={() => setSheetOpen(false)}
        onMealSaved={load}
        trackedMacros={goals?.tracked_macros ?? []}
      />

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

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingTop: 2,
    paddingBottom: 14,
  },
  title: { fontFamily: fonts.heading, fontSize: 30, color: colors.text },
  dateLabel: { fontFamily: fonts.body, fontSize: 12.5, color: colors.accent700, marginTop: 2 },
  scrollContent: { paddingHorizontal: 20, paddingBottom: 8, gap: 12 },
  tileGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 12 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14 },
  promptBar: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    backgroundColor: colors.surface,
    borderRadius: radius.pill,
    padding: 6,
    marginHorizontal: 16,
    marginBottom: 14,
    ...shadow.lg,
  },
  micIcon: {
    width: 40,
    height: 40,
    borderRadius: radius.pill,
    backgroundColor: colors.accent100,
    alignItems: 'center',
    justifyContent: 'center',
  },
  promptText: { fontFamily: fonts.body, color: colors.text, opacity: 0.6, fontSize: 14 },
});
