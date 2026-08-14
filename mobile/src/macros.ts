import { colors } from './theme';
import type { Goals, Nutrition, TrackedMacro } from './types/api';

type NumericGoalKey = 'calorie_goal' | 'protein_goal' | 'carb_goal' | 'fat_goal' | 'fiber_goal' | 'sodium_goal';

export type MacroDef = {
  key: TrackedMacro;
  label: string;
  color: string;
  unit: string;
  goalKey: NumericGoalKey;
};

// Same shape as MacroDef but for calories, which aren't toggleable/tracked
// like the others — kept separate so CaloriesCard can open
// MacroBreakdownSheet the same way MacroTile does.
export type CaloriesDef = { key: 'calories'; label: string; color: string; unit: string; goalKey: 'calorie_goal' };

export const CALORIES_MACRO: CaloriesDef = {
  key: 'calories',
  label: 'CALORIES',
  color: colors.accent500,
  unit: ' kcal',
  goalKey: 'calorie_goal',
};

// Single source of truth for every macro's display metadata, used by
// GoalsScreen (toggles + stepper fields), TodayScreen/HistoryScreen
// (tiles), and NutritionTags (chat log tags) — previously duplicated as a
// separate MACRO_TILES const in each of the two screens.
export const MACROS: MacroDef[] = [
  { key: 'protein', label: 'PROTEIN', color: colors.accent700, unit: 'g', goalKey: 'protein_goal' },
  { key: 'carbs', label: 'CARBS', color: colors.accent2_500, unit: 'g', goalKey: 'carb_goal' },
  { key: 'fat', label: 'FAT', color: colors.neutral700, unit: 'g', goalKey: 'fat_goal' },
  { key: 'fiber', label: 'FIBER', color: colors.accent2_700, unit: 'g', goalKey: 'fiber_goal' },
  { key: 'sodium', label: 'SODIUM', color: colors.accent600, unit: 'mg', goalKey: 'sodium_goal' },
];

export function trackedMacroDefs(tracked: TrackedMacro[]): MacroDef[] {
  return MACROS.filter((m) => tracked.includes(m.key));
}

export function macroValue(total: Nutrition, key: TrackedMacro): number {
  return total[key];
}

export function macroGoal(m: MacroDef, goals: Goals): number {
  return goals[m.goalKey];
}
