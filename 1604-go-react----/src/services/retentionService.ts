import { 
  RetentionAnalysisQuery, 
  RetentionAnalysisResult, 
  RetentionDayResult 
} from '../types';
import { memoryStore } from '../storage/memoryStore';

const MS_PER_DAY = 24 * 60 * 60 * 1000;

function getDefaultTimeRange(): { start: Date; end: Date } {
  const end = new Date();
  const start = new Date(end.getTime() - 7 * MS_PER_DAY);
  return { start, end };
}

function roundToTwo(num: number): number {
  return Math.round(num * 100) / 100;
}

export function analyzeRetention(
  query: RetentionAnalysisQuery
): RetentionAnalysisResult {
  const timeRange = {
    start: query.startTime || getDefaultTimeRange().start,
    end: query.endTime || getDefaultTimeRange().end,
  };

  const baselineEvents = memoryStore.getEvents({
    eventName: query.baselineEvent,
    startTime: timeRange.start,
    endTime: timeRange.end,
  });

  const baselineUserDays = new Map<string, number>();
  for (const event of baselineEvents) {
    if (baselineUserDays.has(event.userId)) {
      continue;
    }
    const dayStart = new Date(event.occurredAt);
    dayStart.setHours(0, 0, 0, 0);
    const dayMs = dayStart.getTime();
    baselineUserDays.set(event.userId, dayMs);
  }

  const baselineUsers = new Set(baselineUserDays.keys());

  const result: RetentionAnalysisResult = {
    baselineEvent: query.baselineEvent,
    baselineUsers: baselineUsers.size,
    queryTime: new Date(),
    timeRange,
    retention: [],
  };

  for (const day of query.days.sort((a, b) => a - b)) {
    let retainedUsers = 0;
    const retentionEventNames = query.retentionEvents.length > 0 
      ? query.retentionEvents 
      : [query.baselineEvent];

    for (const userId of baselineUsers) {
      const firstDayMs = baselineUserDays.get(userId);
      if (firstDayMs === undefined) continue;

      const targetDayStart = new Date(firstDayMs + day * MS_PER_DAY);
      const targetDayEnd = new Date(firstDayMs + (day + 1) * MS_PER_DAY);

      for (const eventName of retentionEventNames) {
        const userRetentionEvents = memoryStore.getEvents({
          userId,
          eventName,
          startTime: targetDayStart,
          endTime: targetDayEnd,
        });

        if (userRetentionEvents.length > 0) {
          retainedUsers++;
          break;
        }
      }
    }

    const retentionRate = baselineUsers.size === 0 
      ? 0 
      : roundToTwo((retainedUsers / baselineUsers.size) * 100);

    result.retention.push({
      day,
      retainedUsers,
      retentionRate,
    });
  }

  return result;
}
