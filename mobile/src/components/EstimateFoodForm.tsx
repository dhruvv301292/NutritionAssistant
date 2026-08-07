import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { createFood, estimateFood } from '../api/client';
import { colors, fonts, radius } from '../theme';
import type { Food, NutritionEstimate } from '../types/api';

export const NOT_MATCHED_ERROR = 'no matching food found';

export function needsEstimate(error: string | undefined): boolean {
  return error === NOT_MATCHED_ERROR;
}

type Props = {
  foodName: string;
  onSaved: () => void;
  // When set, this is an unconfirmed match from an external food database
  // (USDA/FatSecret) rather than a total miss — pre-fills the edit form
  // instead of requiring a separate "suggest values with AI" step.
  externalMatch?: Food;
};

type FieldKey = keyof Pick<NutritionEstimate, 'calories' | 'protein' | 'carbs' | 'fat' | 'fiber'>;

const FIELDS: { key: FieldKey; label: string }[] = [
  { key: 'calories', label: 'Calories' },
  { key: 'protein', label: 'Protein (g)' },
  { key: 'carbs', label: 'Carbs (g)' },
  { key: 'fat', label: 'Fat (g)' },
  { key: 'fiber', label: 'Fiber (g)' },
];

function toEstimate(food: Food): NutritionEstimate {
  return {
    name: food.name,
    calories: food.calories,
    protein: food.protein,
    carbs: food.carbs,
    fat: food.fat,
    fiber: food.fiber,
    sodium: food.sodium,
    unit: food.unit,
    unitquantity: food.unitquantity,
  };
}

export default function EstimateFoodForm({ foodName, onSaved, externalMatch }: Props) {
  const [estimate, setEstimate] = useState<NutritionEstimate | null>(externalMatch ? toEstimate(externalMatch) : null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleEstimate() {
    setLoading(true);
    setError(null);
    try {
      setEstimate(await estimateFood(foodName));
    } catch {
      setError('Could not get an estimate.');
    } finally {
      setLoading(false);
    }
  }

  function updateField(key: FieldKey, value: string) {
    setEstimate((prev) => (prev ? { ...prev, [key]: parseFloat(value) || 0 } : prev));
  }

  async function handleSave() {
    if (!estimate) return;
    setSaving(true);
    setError(null);
    try {
      await createFood(estimate);
      onSaved();
    } catch {
      setError('Could not save this food.');
    } finally {
      setSaving(false);
    }
  }

  if (!estimate) {
    return (
      <View style={styles.container}>
        <Text style={styles.warningText}>"{foodName}" isn't in our database yet.</Text>
        <Pressable style={styles.button} onPress={handleEstimate} disabled={loading}>
          {loading ? <ActivityIndicator color={colors.bg} /> : <Text style={styles.buttonText}>suggest values with AI</Text>}
        </Pressable>
        {error && <Text style={styles.errorText}>{error}</Text>}
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.warningText}>
        {externalMatch
          ? 'Found in an external food database — review and edit before saving:'
          : 'AI-suggested values — review and edit before saving:'}
      </Text>
      {FIELDS.map((f) => (
        <View key={f.key} style={styles.fieldRow}>
          <Text style={styles.fieldLabel}>{f.label}</Text>
          <TextInput
            style={styles.fieldInput}
            keyboardType="numeric"
            value={String(estimate[f.key])}
            onChangeText={(v) => updateField(f.key, v)}
          />
        </View>
      ))}
      <Pressable style={styles.button} onPress={handleSave} disabled={saving}>
        {saving ? <ActivityIndicator color={colors.bg} /> : <Text style={styles.buttonText}>save & add to database</Text>}
      </Pressable>
      {error && <Text style={styles.errorText}>{error}</Text>}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 8, marginTop: 4 },
  warningText: { fontFamily: fonts.body, fontSize: 12, color: '#d68910' },
  errorText: { fontFamily: fonts.body, fontSize: 12, color: '#c0392b' },
  fieldRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  fieldLabel: { fontFamily: fonts.body, fontSize: 13, color: colors.text },
  fieldInput: {
    fontFamily: fonts.body,
    fontSize: 13,
    color: colors.text,
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    paddingHorizontal: 8,
    paddingVertical: 4,
    minWidth: 70,
    textAlign: 'right',
  },
  button: {
    backgroundColor: colors.accent,
    borderRadius: radius.pill,
    paddingVertical: 8,
    alignItems: 'center',
  },
  buttonText: { color: colors.bg, fontSize: 12, fontFamily: fonts.heading },
});
