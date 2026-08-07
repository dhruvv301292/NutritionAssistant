import type { Slot } from './types/api';

export type { Slot };

export const SLOT_ORDER: Slot[] = ['breakfast', 'lunch', 'dinner', 'snack'];
export const SLOT_LABEL: Record<Slot, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snacks',
};

// Sensible default for the slot picker, based on time of day — the user can
// always override it before saving.
export function suggestedSlot(d: Date): Slot {
  const h = d.getHours();
  if (h >= 5 && h < 11) return 'breakfast';
  if (h >= 11 && h < 16) return 'lunch';
  if (h >= 16 && h < 22) return 'dinner';
  return 'snack';
}
