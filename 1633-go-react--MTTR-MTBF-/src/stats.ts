import { listIncidents } from './repository';
import { Incident, MTTRResult, MTBFResult, GroupBy } from './types';

export async function calculateMTTR(
  service?: string,
  startTimeFrom?: string,
  startTimeTo?: string,
  groupBy?: GroupBy
): Promise<any> {
  const query = { service, startTimeFrom, startTimeTo };
  const incidents = await listIncidents(query);

  if (groupBy) {
    return calculateGroupedMTTR(incidents, groupBy);
  }

  const recovered = incidents.filter((i) => i.recoveredTime !== null);
  if (recovered.length === 0) {
    return { mttr: null, unit: 'minutes', count: 0 };
  }

  const totalMinutes = recovered.reduce((sum, i) => {
    const durationMs = i.recoveredTime!.getTime() - i.startTime.getTime();
    return sum + durationMs / (1000 * 60);
  }, 0);

  const avg = totalMinutes / recovered.length;
  return { mttr: Math.round(avg * 100) / 100, unit: 'minutes', count: recovered.length };
}

export async function calculateMTBF(
  service?: string,
  startTimeFrom?: string,
  startTimeTo?: string,
  groupBy?: GroupBy
): Promise<any> {
  const query = { service, startTimeFrom, startTimeTo };
  const incidents = await listIncidents(query);

  const sorted = incidents.sort((a, b) => a.startTime.getTime() - b.startTime.getTime());

  if (groupBy) {
    return calculateGroupedMTBF(sorted, groupBy);
  }

  if (sorted.length < 2) {
    return {
      mtbf: null,
      unit: 'hours',
      count: sorted.length,
      message: 'insufficient data for MTBF calculation',
    };
  }

  let totalHours = 0;
  let gapCount = 0;

  for (let i = 1; i < sorted.length; i++) {
    const gapMs = sorted[i].startTime.getTime() - sorted[i - 1].startTime.getTime();
    totalHours += gapMs / (1000 * 60 * 60);
    gapCount++;
  }

  const avg = totalHours / gapCount;
  return { mtbf: Math.round(avg * 100) / 100, unit: 'hours', count: sorted.length };
}

function getGroupKey(date: Date, groupBy: GroupBy): string {
  const year = date.getFullYear();
  if (groupBy === 'week') {
    const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
    const dayNum = d.getUTCDay() || 7;
    d.setUTCDate(d.getUTCDate() + 4 - dayNum);
    const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
    const weekNo = Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
    return `${year}-W${weekNo.toString().padStart(2, '0')}`;
  } else if (groupBy === 'month') {
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    return `${year}-${month}`;
  } else {
    const quarter = Math.ceil((date.getMonth() + 1) / 3);
    return `${year}-Q${quarter}`;
  }
}

function calculateGroupedMTTR(incidents: Incident[], groupBy: GroupBy): any {
  const groups: Record<string, Incident[]> = {};

  incidents.forEach((i) => {
    const key = getGroupKey(i.startTime, groupBy);
    if (!groups[key]) groups[key] = [];
    groups[key].push(i);
  });

  const result: any[] = [];
  Object.keys(groups)
    .sort()
    .forEach((key) => {
      const groupIncidents = groups[key].filter((i) => i.recoveredTime !== null);
      if (groupIncidents.length === 0) {
        result.push({ group: key, mttr: null, count: 0 });
      } else {
        const totalMinutes = groupIncidents.reduce((sum, i) => {
          const durationMs = i.recoveredTime!.getTime() - i.startTime.getTime();
          return sum + durationMs / (1000 * 60);
        }, 0);
        const avg = totalMinutes / groupIncidents.length;
        result.push({
          group: key,
          mttr: Math.round(avg * 100) / 100,
          count: groupIncidents.length,
        });
      }
    });

  return { groups: result, unit: 'minutes' };
}

function calculateGroupedMTBF(sorted: Incident[], groupBy: GroupBy): any {
  const groups: Record<string, Incident[]> = {};

  sorted.forEach((i) => {
    const key = getGroupKey(i.startTime, groupBy);
    if (!groups[key]) groups[key] = [];
    groups[key].push(i);
  });

  const result: any[] = [];
  Object.keys(groups)
    .sort()
    .forEach((key) => {
      const groupIncidents = groups[key].sort(
        (a, b) => a.startTime.getTime() - b.startTime.getTime()
      );

      if (groupIncidents.length < 2) {
        result.push({
          group: key,
          mtbf: null,
          count: groupIncidents.length,
          message: 'insufficient data for MTBF calculation',
        });
      } else {
        let totalHours = 0;
        let gapCount = 0;
        for (let i = 1; i < groupIncidents.length; i++) {
          const gapMs =
            groupIncidents[i].startTime.getTime() - groupIncidents[i - 1].startTime.getTime();
          totalHours += gapMs / (1000 * 60 * 60);
          gapCount++;
        }
        const avg = totalHours / gapCount;
        result.push({
          group: key,
          mtbf: Math.round(avg * 100) / 100,
          count: groupIncidents.length,
        });
      }
    });

  return { groups: result, unit: 'hours' };
}
