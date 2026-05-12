export function parseDateTime(date: string, time: string): Date {
  return new Date(`${date}T${time}`);
}

export function isDatePassed(dateStr: string): boolean {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const date = new Date(dateStr);
  return date < today;
}

export function isWithin24Hours(dateTime: Date): boolean {
  const now = new Date();
  const diffMs = dateTime.getTime() - now.getTime();
  const diffHours = diffMs / (1000 * 60 * 60);
  return diffHours < 24;
}

export function timeOverlap(
  start1: string, end1: string,
  start2: string, end2: string
): boolean {
  return start1 < end2 && start2 < end1;
}

export function calculateServiceMinutes(
  actualStart: Date,
  actualEnd: Date,
  overlaps: Array<{ start: Date; end: Date }> = []
): number {
  let totalMinutes = (actualEnd.getTime() - actualStart.getTime()) / (1000 * 60);
  
  for (const overlap of overlaps) {
    const overlapStart = actualStart > overlap.start ? actualStart : overlap.start;
    const overlapEnd = actualEnd < overlap.end ? actualEnd : overlap.end;
    if (overlapStart < overlapEnd) {
      const overlapMinutes = (overlapEnd.getTime() - overlapStart.getTime()) / (1000 * 60);
      totalMinutes -= overlapMinutes;
    }
  }
  
  return Math.max(0, totalMinutes);
}

export function roundServiceMinutes(minutes: number): number {
  if (minutes < 30) return 0;
  if (minutes < 60) return 30;
  return Math.floor(minutes / 30) * 30;
}

export function getBadge(annualHours: number): string {
  if (annualHours >= 300) return 'gold';
  if (annualHours >= 100) return 'silver';
  if (annualHours >= 50) return 'bronze';
  return 'none';
}
