export function dateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

export function formatDateLabel(d: Date): string {
  return d.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' });
}

export function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
}

export function last14Days(from: Date): Date[] {
  return Array.from({ length: 14 }, (_, i) => {
    const d = new Date(from);
    d.setDate(d.getDate() - (13 - i));
    return d;
  });
}

// Monday of the week containing `d`.
export function startOfWeek(d: Date): Date {
  const result = new Date(d);
  const day = result.getDay(); // 0 = Sunday .. 6 = Saturday
  const diff = day === 0 ? -6 : 1 - day;
  result.setDate(result.getDate() + diff);
  return result;
}

export function weekDates(weekStart: Date): Date[] {
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart);
    d.setDate(d.getDate() + i);
    return d;
  });
}
