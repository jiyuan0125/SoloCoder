import { 
  PathAnalysisQuery, 
  PathAnalysisResult, 
  PathResult 
} from '../types';
import { memoryStore } from '../storage/memoryStore';

const MS_PER_DAY = 24 * 60 * 60 * 1000;

function getDefaultTimeRange(): { start: Date; end: Date } {
  const end = new Date();
  const start = new Date(end.getTime() - 7 * MS_PER_DAY);
  return { start, end };
}

export function analyzePaths(
  query: PathAnalysisQuery
): PathAnalysisResult {
  const timeRange = {
    start: query.startTime || getDefaultTimeRange().start,
    end: query.endTime || getDefaultTimeRange().end,
  };
  const maxDepth = query.maxDepth || 10;

  const events = memoryStore.getEvents({
    startTime: timeRange.start,
    endTime: timeRange.end,
    userId: query.userId,
  });

  const eventsByUser = new Map<string, string[]>();
  for (const event of events) {
    if (!eventsByUser.has(event.userId)) {
      eventsByUser.set(event.userId, []);
    }
    eventsByUser.get(event.userId)!.push(event.pagePath || event.eventName);
  }

  const pathCounts = new Map<string, PathResult>();

  for (const [userId, userPath] of eventsByUser.entries()) {
    const normalizedPath: string[] = [];
    const loopCounts: number[] = [];

    if (userPath.length === 0) continue;

    let currentPath = userPath[0];
    let currentLoop = 1;

    for (let i = 1; i < userPath.length; i++) {
      if (userPath[i] === currentPath) {
        currentLoop++;
      } else {
        normalizedPath.push(currentPath);
        loopCounts.push(currentLoop);
        currentPath = userPath[i];
        currentLoop = 1;
      }
    }
    normalizedPath.push(currentPath);
    loopCounts.push(currentLoop);

    const trimmedPath = normalizedPath.slice(0, maxDepth);
    const trimmedLoops = loopCounts.slice(0, maxDepth);
    const pathKey = trimmedPath.join('|');

    if (!pathCounts.has(pathKey)) {
      pathCounts.set(pathKey, {
        path: trimmedPath,
        loopCounts: trimmedLoops,
        userCount: 0,
      });
    }
    pathCounts.get(pathKey)!.userCount++;
  }

  const paths = Array.from(pathCounts.values()).sort(
    (a, b) => b.userCount - a.userCount
  );

  return {
    paths,
    totalPaths: paths.length,
    queryTime: new Date(),
    timeRange,
  };
}
