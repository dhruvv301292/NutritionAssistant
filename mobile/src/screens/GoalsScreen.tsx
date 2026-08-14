import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Switch, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { getGoals, putGoals } from '../api/client';
import { useAuth } from '../auth/AuthContext';
import AccountButton from '../components/AccountButton';
import { MACROS } from '../macros';
import { colors, fonts, radius } from '../theme';
import type { Goals, TrackedMacro } from '../types/api';

type NumericGoalKey = 'calorie_goal' | 'protein_goal' | 'carb_goal' | 'fat_goal' | 'fiber_goal' | 'sodium_goal';

const STEP: Record<NumericGoalKey, number> = {
  calorie_goal: 100,
  protein_goal: 10,
  carb_goal: 10,
  fat_goal: 5,
  fiber_goal: 5,
  sodium_goal: 100,
};

export default function GoalsScreen({ focused }: { focused: boolean }) {
  const { state } = useAuth();
  const [goals, setGoals] = useState<Goals | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    setError(null);
    getGoals()
      .then(setGoals)
      .catch(() => setError('Could not load goals.'));
  }, []);

  useEffect(() => { if (focused) load(); }, [focused, load]);

  async function persist(next: Goals) {
    setGoals(next);
    setSaving(true);
    setError(null);
    try {
      const saved = await putGoals(next);
      setGoals(saved);
    } catch {
      setError('Could not save goals.');
    } finally {
      setSaving(false);
    }
  }

  function adjust(key: NumericGoalKey, delta: number) {
    if (!goals) return;
    persist({ ...goals, [key]: Math.max(0, goals[key] + delta) });
  }

  function toggleMacro(key: TrackedMacro) {
    if (!goals) return;
    const tracked = goals.tracked_macros.includes(key)
      ? goals.tracked_macros.filter((m) => m !== key)
      : [...goals.tracked_macros, key];
    persist({ ...goals, tracked_macros: tracked });
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.headerRow}>
        <Text style={styles.title}>Goals</Text>
        <AccountButton />
      </View>

      <ScrollView contentContainerStyle={styles.scrollContent}>
        {error && <Text style={styles.errorText}>{error}</Text>}
        {!goals && !error && <ActivityIndicator color={colors.accent700} />}

        {goals && (
          <>
            <Text style={styles.kicker}>Daily targets and limits</Text>
            <View style={{ gap: 12, marginBottom: 22 }}>
              <View style={[styles.field, styles.fieldCard]}>
                <View style={styles.fieldHeaderRow}>
                  <View style={styles.toggleSlot} />
                  <Text style={[styles.fieldLabel, styles.fieldLabelCentered]}>Calories (kcal)</Text>
                  <View style={styles.toggleSlot} />
                </View>
                <View style={styles.stepperRow}>
                  <Pressable style={styles.stepperButton} onPress={() => adjust('calorie_goal', -STEP.calorie_goal)}>
                    <Text style={styles.stepperButtonText}>−</Text>
                  </Pressable>
                  <Text style={styles.stepperValue}>{goals.calorie_goal}</Text>
                  <Pressable
                    style={[styles.stepperButton, styles.stepperButtonPrimary]}
                    onPress={() => adjust('calorie_goal', STEP.calorie_goal)}
                  >
                    <Text style={[styles.stepperButtonText, { color: colors.bg }]}>+</Text>
                  </Pressable>
                </View>
              </View>
              {[...MACROS]
                .sort((a, b) => {
                  const aOn = goals.tracked_macros.includes(a.key);
                  const bOn = goals.tracked_macros.includes(b.key);
                  return aOn === bOn ? 0 : aOn ? -1 : 1;
                })
                .map((m) => {
                const on = goals.tracked_macros.includes(m.key);
                return (
                  <View key={m.key} style={[styles.field, styles.fieldCard]}>
                    <View style={styles.fieldHeaderRow}>
                      <View style={styles.toggleSlot}>
                        <Switch
                          style={styles.toggleSmall}
                          value={on}
                          onValueChange={() => toggleMacro(m.key)}
                          trackColor={{ true: colors.accent }}
                        />
                      </View>
                      <Text style={[styles.fieldLabel, styles.fieldLabelCentered]}>
                        {m.label.charAt(0) + m.label.slice(1).toLowerCase()} ({m.unit})
                      </Text>
                      <View style={styles.toggleSlot} />
                    </View>
                    <View style={styles.stepperRow}>
                      <Pressable
                        style={[styles.stepperButton, !on && styles.stepperButtonDisabled]}
                        disabled={!on}
                        onPress={() => adjust(m.goalKey, -STEP[m.goalKey])}
                      >
                        <Text style={[styles.stepperButtonText, !on && styles.stepperButtonTextDisabled]}>−</Text>
                      </Pressable>
                      <Text style={[styles.stepperValue, !on && styles.stepperValueDisabled]}>{goals[m.goalKey]}</Text>
                      <Pressable
                        style={[styles.stepperButton, styles.stepperButtonPrimary, !on && styles.stepperButtonDisabled]}
                        disabled={!on}
                        onPress={() => adjust(m.goalKey, STEP[m.goalKey])}
                      >
                        <Text style={[styles.stepperButtonText, { color: colors.bg }, !on && styles.stepperButtonTextDisabled]}>+</Text>
                      </Pressable>
                    </View>
                  </View>
                );
              })}
            </View>

            {saving && <ActivityIndicator color={colors.accent700} style={{ marginBottom: 8 }} />}

            {state.status === 'signedIn' && (
              <Text style={styles.accountEmail}>{state.user.email}</Text>
            )}
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
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
  scrollContent: { paddingHorizontal: 20, paddingBottom: 24 },
  kicker: { fontSize: 19, fontFamily: fonts.body, fontWeight: '400', color: colors.text, marginBottom: 10 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14, marginBottom: 12 },
  field: { gap: 10 },
  fieldCard: {
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: 14,
  },
  fieldHeaderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  fieldLabel: { fontFamily: fonts.body, fontSize: 16, color: colors.text, opacity: 0.85 },
  fieldLabelCentered: { flex: 1, textAlign: 'center' },
  toggleSlot: { width: 38, alignItems: 'flex-start', justifyContent: 'center' },
  toggleSmall: { transform: [{ translateX: -13 }, { scale: 0.5 }] },
  stepperRow: { flexDirection: 'row', alignItems: 'center', gap: 14 },
  stepperButton: {
    width: 38,
    height: 38,
    borderRadius: radius.pill,
    backgroundColor: colors.neutral200,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stepperButtonPrimary: { backgroundColor: colors.accent },
  stepperButtonDisabled: { backgroundColor: colors.neutral100, opacity: 0.4 },
  stepperButtonText: { fontFamily: fonts.heading, fontSize: 19, color: colors.text },
  stepperButtonTextDisabled: { opacity: 0.4 },
  stepperValue: { flex: 1, textAlign: 'center', fontFamily: fonts.heading, fontSize: 28, color: colors.text },
  stepperValueDisabled: { opacity: 0.4 },
  accountEmail: { fontFamily: fonts.body, fontSize: 13, color: colors.text, opacity: 0.5, marginTop: 20, textAlign: 'center' },
});
