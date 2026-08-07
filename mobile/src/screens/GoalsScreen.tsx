import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { getGoals, putGoals, CURRENT_USER_ID } from '../api/client';
import { colors, fonts, radius, shadow } from '../theme';
import type { Goals } from '../types/api';

type FieldKey = 'calorie_goal' | 'protein_goal' | 'fiber_goal';

const FIELDS: { key: FieldKey; label: string; step: number }[] = [
  { key: 'calorie_goal', label: 'Calories (kcal)', step: 100 },
  { key: 'protein_goal', label: 'Protein (g)', step: 10 },
  { key: 'fiber_goal', label: 'Fiber (g)', step: 5 },
];

export default function GoalsScreen({ focused }: { focused: boolean }) {
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

  function adjust(key: FieldKey, delta: number) {
    setGoals((prev) => (prev ? { ...prev, [key]: Math.max(0, prev[key] + delta) } : prev));
  }

  async function handleSave() {
    if (!goals) return;
    setSaving(true);
    try {
      const saved = await putGoals({ ...goals, user_id: CURRENT_USER_ID });
      setGoals(saved);
    } catch {
      setError('Could not save goals.');
    } finally {
      setSaving(false);
    }
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Text style={styles.title}>Goals</Text>

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

            <Pressable style={[styles.saveButton, shadow.sm]} onPress={handleSave} disabled={saving}>
              <Text style={styles.saveButtonText}>{saving ? 'Saving…' : 'Save goals'}</Text>
            </Pressable>
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  title: { fontFamily: fonts.heading, fontSize: 30, color: colors.text, paddingHorizontal: 20, marginBottom: 10 },
  scrollContent: { paddingHorizontal: 20, paddingBottom: 24 },
  kicker: { fontSize: 14, fontFamily: fonts.heading, color: colors.text, marginBottom: 10 },
  errorText: { fontFamily: fonts.body, color: '#c0392b', fontSize: 14, marginBottom: 12 },
  field: { gap: 6 },
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
  saveButton: {
    backgroundColor: colors.accent,
    borderRadius: radius.pill,
    paddingVertical: 14,
    alignItems: 'center',
  },
  saveButtonText: { color: colors.bg, fontFamily: fonts.heading, fontSize: 14 },
});
