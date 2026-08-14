import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { estimateFood } from '../api/client';
import { colors, fonts, radius } from '../theme';
import type { Food, NutritionEstimate, TrackedMacro } from '../types/api';

export const NOT_MATCHED_ERROR = 'no matching food found';

export function needsEstimate(error: string | undefined): boolean {
  return error === NOT_MATCHED_ERROR;
}

type Props = {
  foodName: string;
  // Threaded through to the estimate request and stamped onto the created
  // food, so a branded item never gets saved with brand=null — without
  // this, brand-based matching/dedup (see backend foods/repository.go)
  // can never find this food again and silently creates a duplicate row
  // every time the same branded product is logged.
  brand?: string;
  // Controlled by the parent (LogMealSheet) rather than owned here, so
  // "log this meal" can save whatever's currently in the form — including
  // unsaved edits — for every pending item in one pass, instead of
  // requiring a separate "save & add to database" tap per item first.
  estimate: NutritionEstimate | null;
  onEstimateChange: (estimate: NutritionEstimate) => void;
  // When set, this is an unconfirmed match from an external food database
  // rather than a total miss — pre-fills the edit form instead of
  // requiring a separate "suggest values with AI" step.
  externalMatch?: Food;
  trackedMacros?: TrackedMacro[];
};

type FieldKey = keyof Pick<NutritionEstimate, 'calories' | 'protein' | 'carbs' | 'fat' | 'fiber' | 'sodium'>;

const BASE_FIELDS: { key: FieldKey; label: string; macro?: TrackedMacro }[] = [
  { key: 'calories', label: 'Calories' },
  { key: 'protein', label: 'Protein (g)', macro: 'protein' },
  { key: 'carbs', label: 'Carbs (g)', macro: 'carbs' },
  { key: 'fat', label: 'Fat (g)', macro: 'fat' },
  { key: 'fiber', label: 'Fiber (g)', macro: 'fiber' },
  { key: 'sodium', label: 'Sodium (mg)', macro: 'sodium' },
];

function toEstimate(food: Food): NutritionEstimate {
  return {
    name: food.name,
    brand: food.brand,
    calories: food.calories,
    protein: food.protein,
    carbs: food.carbs,
    fat: food.fat,
    fiber: food.fiber,
    sodium: food.sodium,
    unit: food.unit,
    unitquantity: food.unitquantity,
    grams_per_unit: food.grams_per_unit,
  };
}

export default function EstimateFoodForm({ foodName, brand, estimate, onEstimateChange, externalMatch, trackedMacros }: Props) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const FIELDS = BASE_FIELDS.filter((f) => !f.macro || !trackedMacros || trackedMacros.includes(f.macro));

  // externalMatch arrives already resolved (from the chat response), so
  // seed the parent's estimate state with it as soon as this item shows up
  // — no separate fetch needed for that path.
  useEffect(() => {
    if (externalMatch && !estimate) {
      onEstimateChange(toEstimate(externalMatch));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [externalMatch]);

  async function handleEstimate() {
    setLoading(true);
    setError(null);
    try {
      onEstimateChange(await estimateFood(foodName, brand));
    } catch {
      setError('Could not get an estimate.');
    } finally {
      setLoading(false);
    }
  }

  function updateField(key: FieldKey, value: string) {
    if (!estimate) return;
    onEstimateChange({ ...estimate, [key]: parseFloat(value) || 0 });
  }

  function updateUnitQuantity(value: string) {
    if (!estimate) return;
    onEstimateChange({ ...estimate, unitquantity: parseFloat(value) || 0 });
  }

  function updateGramsPerUnit(value: string) {
    if (!estimate) return;
    const parsed = parseFloat(value);
    onEstimateChange({ ...estimate, grams_per_unit: Number.isFinite(parsed) ? parsed : undefined });
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
      <Text style={styles.warningText}>Please review before logging the meal:</Text>
      <View style={styles.fieldRow}>
        <Text style={styles.fieldLabel}>Values are per</Text>
        <View style={styles.basisRow}>
          <TextInput
            style={[styles.fieldInput, styles.basisQuantityInput]}
            keyboardType="numeric"
            value={String(estimate.unitquantity)}
            onChangeText={updateUnitQuantity}
          />
          <Text style={styles.basisUnitText}>{estimate.unit}</Text>
        </View>
      </View>
      {estimate.unit === 'count' && (
        <View style={styles.fieldRow}>
          <Text style={styles.fieldLabel}>1 {estimate.unit} =</Text>
          <View style={styles.basisRow}>
            <TextInput
              style={[styles.fieldInput, styles.basisQuantityInput]}
              keyboardType="numeric"
              placeholder="?"
              value={estimate.grams_per_unit != null ? String(estimate.grams_per_unit) : ''}
              onChangeText={updateGramsPerUnit}
            />
            <Text style={styles.basisUnitText}>grams</Text>
          </View>
        </View>
      )}
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
      {error && <Text style={styles.errorText}>{error}</Text>}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 8 },
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
  basisRow: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  basisQuantityInput: { minWidth: 50 },
  basisUnitText: { fontFamily: fonts.body, fontSize: 13, color: colors.text },
  button: {
    backgroundColor: colors.accent,
    borderRadius: radius.pill,
    paddingVertical: 8,
    alignItems: 'center',
  },
  buttonText: { color: colors.bg, fontSize: 12, fontFamily: fonts.heading },
});
