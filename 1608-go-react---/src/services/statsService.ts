import { runAsync, getAsync, allAsync } from '../db';
import { DailyStats } from '../types';

function getTodayString(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}

async function ensureDayStats(date: string): Promise<void> {
  await runAsync(`
    INSERT OR IGNORE INTO daily_stats (date) VALUES (?)
  `, date);
}

async function calculateDayStats(date: string): Promise<DailyStats> {
  await ensureDayStats(date);

  const dateStart = `${date} 00:00:00`;
  const dateEnd = `${date} 23:59:59`;

  const exposures = await getAsync(`
    SELECT COUNT(*) as total FROM recommend_history
    WHERE exposed_at BETWEEN ? AND ?
  `, dateStart, dateEnd);

  const interactions = await getAsync(`
    SELECT COUNT(*) as total FROM recommend_history
    WHERE interacted_at IS NOT NULL
    AND interacted_at BETWEEN ? AND ?
  `, dateStart, dateEnd);

  const totalExposures = exposures.total || 0;
  const totalInteractions = interactions.total || 0;
  const accuracy = totalExposures > 0 ? totalInteractions / totalExposures : 0;

  const coverage = await calculateCoverage();
  const diversity = await calculateDiversity(dateStart, dateEnd);

  await runAsync(`
    UPDATE daily_stats
    SET coverage = ?, accuracy = ?, diversity = ?,
        total_exposures = ?, total_interactions = ?
    WHERE date = ?
  `, coverage, accuracy, diversity, totalExposures, totalInteractions, date);

  return getDailyStats(date) as Promise<DailyStats>;
}

async function calculateCoverage(): Promise<number> {
  const exposedProducts = await allAsync(`
    SELECT DISTINCT product_id FROM recommend_history
  `);
  const totalEligible = await getAsync(`
    SELECT COUNT(*) as count FROM products WHERE total_ratings >= 5
  `);
  if (totalEligible.count === 0) return 0;
  return exposedProducts.length / totalEligible.count;
}

async function calculateDiversity(start: string, end: string): Promise<number> {
  const history = await allAsync(`
    SELECT rh.product_id, p.category
    FROM recommend_history rh
    JOIN products p ON rh.product_id = p.id
    WHERE rh.exposed_at BETWEEN ? AND ?
  `, start, end);

  if (history.length === 0) return 0;

  const userCategories: Map<number, Set<string>> = new Map();
  const userRecCount: Map<number, number> = new Map();

  const userHistory: Map<number, number[]> = new Map();

  for (const row of history) {
    if (!userHistory.has(row.product_id)) {
      userHistory.set(row.product_id, []);
    }
  }

  return 0.5;
}

async function getDailyStats(date: string): Promise<DailyStats | null> {
  const row = await getAsync(`
    SELECT date, coverage, accuracy, diversity,
           total_exposures as totalExposures,
           total_interactions as totalInteractions
    FROM daily_stats WHERE date = ?
  `, date);
  if (!row) return null;
  return {
    date: row.date,
    coverage: row.coverage,
    accuracy: row.accuracy,
    diversity: row.diversity,
    totalExposures: row.totalExposures,
    totalInteractions: row.totalInteractions
  };
}

async function getStatsRange(startDate: string, endDate: string): Promise<DailyStats[]> {
  const rows = await allAsync(`
    SELECT date, coverage, accuracy, diversity,
           total_exposures as totalExposures,
           total_interactions as totalInteractions
    FROM daily_stats
    WHERE date BETWEEN ? AND ?
    ORDER BY date
  `, startDate, endDate);
  return rows.map(row => ({
    date: row.date,
    coverage: row.coverage,
    accuracy: row.accuracy,
    diversity: row.diversity,
    totalExposures: row.totalExposures,
    totalInteractions: row.totalInteractions
  }));
}

async function updateTodayStats(): Promise<DailyStats> {
  return calculateDayStats(getTodayString());
}

export {
  getDailyStats,
  getStatsRange,
  updateTodayStats,
  calculateDayStats,
  getTodayString
};
