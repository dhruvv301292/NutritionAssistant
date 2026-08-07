// The backend has no concept of meal "slots" (breakfast/lunch/dinner) —
// meal_logs only stores a timestamp. Same approach as the design mockup:
// derive a slot label client-side from the hour logged.
export type Slot = 'breakfast' | 'lunch' | 'dinner' | 'snack';

export const SLOT_ORDER: Slot[] = ['breakfast', 'lunch', 'dinner', 'snack'];
export const SLOT_LABEL: Record<Slot, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snacks',
};

export function slotForTime(d: Date): Slot {
  const h = d.getHours();
  if (h >= 5 && h < 11) return 'breakfast';
  if (h >= 11 && h < 16) return 'lunch';
  if (h >= 16 && h < 22) return 'dinner';
  return 'snack';
}
