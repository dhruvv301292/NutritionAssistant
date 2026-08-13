import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Switch, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { getGoals, putGoals } from '../api/client';
import { useAuth } from '../auth/AuthContext';
import AccountButton from '../components/AccountButton';
import { MACROS } from '../macros';
import { colors, fonts, radius } from '../theme';
import type { Goals, TrackedMacro } from '../types/api';

type FieldKey = 'calorie_goal' | 'protein_goal' | 'fiber_goal';

const FIELDS: { key: FieldKey; label: string; step: number }[] = [
  { key: 'calorie_goal', label: 'Calories (kcal)', step: 100 },
  { key: 'protein_goal', label: 'Protein (g)', step: 10 },
  { key: 'fiber_goal', label: 'Fiber (g)', step: 5 },
];

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

  function adjust(key: FieldKey, delta: number) {
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
            <Text style={styles.kicker}>Daily targets</Text>
            <View style={{ gap: 12, marginBottom: 22 }}>
              {FIELDS.map((f) => (
                <View key={f.key} style={styles.field}>
                  <Text style={styles.fieldLabel}>{f.label}</Text>
                  <View style={styles.stepperRow}>
                    <Pressable style={styles.stepperButton} onPress={() => adjust(f.key, -f.step)}>
                      <Text style={styles.stepperButtonText}>−</Text>
                    </Pressable>
                    <Text style={styles.stepperValue}>{goals[f.key]}</Text>
                    <Pressable
                      style={[styles.stepperButton, styles.stepperButtonPrimary]}
                      onPress={() => adjust(f.key, f.step)}
                    >
                      <Text style={[styles.stepperButtonText, { color: colors.bg }]}>+</Text>
                    </Pressable>
                  </View>
                </View>
              ))}
            </View>

            <Text style={styles.kicker}>Tracked macros</Text>
            <View style={{ gap: 4, marginBottom: 22 }}>
              {MACROS.map((m) => (
                <View key={m.key} style={styles.toggleRow}>
                  <Text style={styles.toggleLabel}>{m.label.charAt(0) + m.label.slice(1).toLowerCase()}</Text>
                  <Switch
                    value={goals.tracked_macros.includes(m.key)}
                    onValueChange={() => toggleMacro(m.key)}
                    trackColor={{ true: colors.accent }}
                  />
                </View>
              ))}
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
  kicker: { fontSize: 19, fontFamily: fonts.heading, color: colors.text, marginBottom: 10 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14, marginBottom: 12 },
  field: { gap: 6 },
  toggleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
  },
  toggleLabel: { fontFamily: fonts.body, fontSize: 15, color: colors.text },
  fieldLabel: { fontFamily: fonts.body, fontSize: 12, color: colors.text, opacity: 0.7 },
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
  stepperButtonText: { fontFamily: fonts.heading, fontSize: 19, color: colors.text },
  stepperValue: { flex: 1, textAlign: 'center', fontFamily: fonts.heading, fontSize: 28, color: colors.text },
  accountEmail: { fontFamily: fonts.body, fontSize: 13, color: colors.text, opacity: 0.5, marginTop: 20, textAlign: 'center' },
});
