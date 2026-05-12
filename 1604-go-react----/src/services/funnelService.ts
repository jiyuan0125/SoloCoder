import { 
  Funnel, 
  FunnelAnalysisQuery, 
  FunnelAnalysisResult, 
  Event, 
  FunnelStepResult 
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

export function analyzeFunnel(
  funnel: Funnel,
  query: FunnelAnalysisQuery
): FunnelAnalysisResult {
  const timeRange = {
    start: query.startTime || getDefaultTimeRange().start,
    end: query.endTime || getDefaultTimeRange().end,
  };
  const timeWindowDays = query.timeWindowDays || 7;
  const timeWindowMs = timeWindowDays * MS_PER_DAY;

  const steps = [...funnel.steps].sort((a, b) => a.order - b.order);
  const events = memoryStore.getEvents({
    startTime: timeRange.start,
    endTime: timeRange.end,
  });

  const eventsByUser = new Map<string, Event[]>();
  for (const event of events) {
    if (!eventsByUser.has(event.userId)) {
      eventsByUser.set(event.userId, []);
    }
    eventsByUser.get(event.userId)!.push(event);
  }

  const stepUserCounts: number[] = new Array(steps.length).fill(0);

  for (const [userId, userEvents] of eventsByUser.entries()) {
    const sortedEvents = [...userEvents].sort(
      (a, b) => a.occurredAt.getTime() - b.occurredAt.getTime()
    );

    const firstStepEvent = sortedEvents.find(
      e => e.eventName === steps[0].eventName
    );
    if (!firstStepEvent) continue;

    const funnelStartTime = firstStepEvent.occurredAt.getTime();
    const funnelEndTime = funnelStartTime + timeWindowMs;

    let currentStepIndex = 0;
    const userCompletedSteps = new Set<number>();

    for (const event of sortedEvents) {
      const eventTime = event.occurredAt.getTime();
      if (eventTime < funnelStartTime || eventTime > funnelEndTime) continue;

      if (event.eventName === steps[currentStepIndex].eventName) {
        userCompletedSteps.add(currentStepIndex);
        if (currentStepIndex < steps.length - 1) {
          currentStepIndex++;
        } else {
          break;
        }
      }
    }

    for (const stepIdx of userCompletedSteps) {
      stepUserCounts[stepIdx]++;
    }
  }

  const stepResults: FunnelStepResult[] = [];
  for (let i = 0; i < steps.length; i++) {
    const conversionRate = i === 0 
      ? 100.0 
      : stepUserCounts[i - 1] === 0 
        ? 0 
        : roundToTwo((stepUserCounts[i] / stepUserCounts[i - 1]) * 100);
    
    stepResults.push({
      step: steps[i].order,
      eventName: steps[i].eventName,
      uniqueUsers: stepUserCounts[i],
      conversionRate,
    });
  }

  const overallConversionRate = stepUserCounts[0] === 0
    ? 0
    : roundToTwo(
        (stepUserCounts[stepUserCounts.length - 1] / stepUserCounts[0]) * 100
      );

  return {
    funnelId: funnel.id,
    funnelVersion: funnel.version,
    queryTime: new Date(),
    timeRange,
    steps: stepResults,
    overallConversionRate,
  };
}
