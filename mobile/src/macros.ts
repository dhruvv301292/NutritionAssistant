import { colors } from './theme';
import type { Goals, Nutrition, TrackedMacro } from './types/api';

// FDA/AHA general daily sodium guideline (2300mg) — sodium has no
// user-settable goal (see GoalsScreen), so this is a fixed reference bar
// rather than something the user can customize.
export const SODIUM_REFERENCE_MG = 2300;

type NumericGoalKey = 'calorie_goal' | 'protein_goal' | 'carb_goal' | 'fat_goal' | 'fiber_goal';

export type MacroDef = {
  key: TrackedMacro;
  label: string;
  color: string;
  unit: string;
  // goalKey is undefined for sodium — it has no per-user target, only the
  // fixed SODIUM_REFERENCE_MG reference.
  goalKey?: NumericGoalKey;
};

// Single source of truth for every macro's display metadata, used by
// GoalsScreen (toggles), TodayScreen/HistoryScreen (tiles), and
// NutritionTags (chat log tags) — previously duplicated as a separate
// MACRO_TILES const in each of the two screens.
export const MACROS: MacroDef[] = [
  { key: 'protein', label: 'PROTEIN', color: colors.accent700, unit: 'g', goalKey: 'protein_goal' },
  { key: 'carbs', label: 'CARBS', color: colors.accent2_500, unit: 'g', goalKey: 'carb_goal' },
  { key: 'fat', label: 'FAT', color: colors.neutral700, unit: 'g', goalKey: 'fat_goal' },
  { key: 'fiber', label: 'FIBER', color: colors.accent2_700, unit: 'g', goalKey: 'fiber_goal' },
  { key: 'sodium', label: 'SODIUM', color: colors.accent600, unit: 'mg' },
];

export function trackedMacroDefs(tracked: TrackedMacro[]): MacroDef[] {
  return MACROS.filter((m) => tracked.includes(m.key));
}

export function macroValue(total: Nutrition, key: TrackedMacro): number {
  return total[key];
}

export function macroGoal(m: MacroDef, goals: Goals): number {
  return m.goalKey ? goals[m.goalKey] : SODIUM_REFERENCE_MG;
}
