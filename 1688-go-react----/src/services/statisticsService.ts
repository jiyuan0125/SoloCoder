import { db } from '../database';

const FORTY_EIGHT_HOURS = 48 * 60 * 60 * 1000;

export interface LevelDistribution {
  normal: number;
  mild: number;
  moderate: number;
  severe: number;
}

export interface ScaleAverageTrend {
  scaleId: string;
  scaleName: string;
  averages: { date: string; averageScore: number }[];
}

export interface AlertTimeliness {
  totalAlerts: number;
  onTime: number;
  overdue: number;
  timelinessRate: number;
}

export const getLevelDistribution = (): Promise<LevelDistribution> => {
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT level, COUNT(*) as count FROM assessment_results GROUP BY level`,
      [],
      (err, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        const distribution: LevelDistribution = {
          normal: 0,
          mild: 0,
          moderate: 0,
          severe: 0
        };

        for (const row of rows) {
          if (row.level in distribution) {
            (distribution as any)[row.level] = row.count;
          }
        }

        resolve(distribution);
      }
    );
  });
};

export const getScaleAverageTrend = (days: number = 30): Promise<ScaleAverageTrend[]> => {
  return new Promise((resolve, reject) => {
    const cutoffTime = Date.now() - days * 24 * 60 * 60 * 1000;

    db.all(
      `SELECT ar.scale_id, s.name as scale_name, ar.completed_at, ar.standard_score
       FROM assessment_results ar
       JOIN scales s ON ar.scale_id = s.id
       WHERE ar.completed_at >= ?
       ORDER BY ar.scale_id, ar.completed_at`,
      [cutoffTime],
      (err, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        const result: { [key: string]: ScaleAverageTrend } = {};
        const dailyScores: { [key: string]: { [key: string]: number[] } } = {};

        for (const row of rows) {
          const scaleId = row.scale_id;
          const date = new Date(row.completed_at).toISOString().split('T')[0];

          if (!result[scaleId]) {
            result[scaleId] = {
              scaleId,
              scaleName: row.scale_name,
              averages: []
            };
            dailyScores[scaleId] = {};
          }

          if (!dailyScores[scaleId][date]) {
            dailyScores[scaleId][date] = [];
          }
          dailyScores[scaleId][date].push(row.standard_score);
        }

        for (const scaleId in dailyScores) {
          for (const date in dailyScores[scaleId]) {
            const scores = dailyScores[scaleId][date];
            const avg = scores.reduce((a, b) => a + b, 0) / scores.length;
            result[scaleId].averages.push({
              date,
              averageScore: Math.round(avg * 100) / 100
            });
          }
        }

        resolve(Object.values(result));
      }
    );
  });
};

export const getAlertTimeliness = (): Promise<AlertTimeliness> => {
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT created_at, resolved_at, status FROM alerts WHERE status = 'resolved'`,
      [],
      (err, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        let onTime = 0;
        let overdue = 0;

        for (const row of rows) {
          if (row.resolved_at) {
            const duration = row.resolved_at - row.created_at;
            if (duration <= FORTY_EIGHT_HOURS) {
              onTime++;
            } else {
              overdue++;
            }
          }
        }

        const total = onTime + overdue;
        const timelinessRate = total > 0 ? (onTime / total) * 100 : 0;

        resolve({
          totalAlerts: total,
          onTime,
          overdue,
          timelinessRate: Math.round(timelinessRate * 100) / 100
        });
      }
    );
  });
};
