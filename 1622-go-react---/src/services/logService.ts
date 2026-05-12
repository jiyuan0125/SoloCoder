import { db } from '../database';
import { LogEntry, LogLevel, VALID_LOG_LEVELS, LogQueryParams, AggregationStats } from '../types';

export function insertLog(log: LogEntry): Promise<void> {
  return new Promise((resolve, reject) => {
    db.run(
      `INSERT INTO logs (timestamp, level, serviceName, traceId, message, tags) 
       VALUES (?, ?, ?, ?, ?, ?)`,
      [
        log.timestamp,
        log.level,
        log.serviceName,
        log.traceId,
        log.message,
        JSON.stringify(log.tags),
      ],
      (err) => {
        if (err) {
          reject(err);
        } else {
          resolve();
        }
      }
    );
  });
}

export function queryLogs(params: LogQueryParams & { serviceName?: string }): Promise<{ logs: LogEntry[]; total: number }> {
  return new Promise((resolve, reject) => {
    const conditions: string[] = [];
    const values: (number | string)[] = [];

    if (params.startTime !== undefined) {
      conditions.push('timestamp >= ?');
      values.push(params.startTime);
    }
    if (params.endTime !== undefined) {
      conditions.push('timestamp <= ?');
      values.push(params.endTime);
    }
    if (params.level) {
      conditions.push('level = ?');
      values.push(params.level);
    }
    if (params.serviceName) {
      conditions.push('serviceName = ?');
      values.push(params.serviceName);
    }
    if (params.keyword) {
      conditions.push('message LIKE ?');
      values.push(`%${params.keyword}%`);
    }

    const whereClause = conditions.length > 0 ? `WHERE ${conditions.join(' AND ')}` : '';

    db.get(`SELECT COUNT(*) as total FROM logs ${whereClause}`, values, (err, row: { total: number } | undefined) => {
      if (err) {
        reject(err);
        return;
      }

      const total = row?.total || 0;
      const page = params.page || 1;
      const pageSize = params.pageSize || 50;
      const offset = (page - 1) * pageSize;

      db.all(
        `SELECT * FROM logs ${whereClause} ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
        [...values, pageSize, offset],
        (err, rows: any[]) => {
          if (err) {
            reject(err);
            return;
          }

          const logs: LogEntry[] = rows.map(row => ({
            id: row.id,
            timestamp: row.timestamp,
            level: row.level as LogLevel,
            serviceName: row.serviceName,
            traceId: row.traceId,
            message: row.message,
            tags: JSON.parse(row.tags),
          }));

          resolve({ logs, total });
        }
      );
    });
  });
}

export function queryTraceLogs(traceId: string): Promise<LogEntry[]> {
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT * FROM logs WHERE traceId = ? ORDER BY timestamp ASC`,
      [traceId],
      (err, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        const logs: LogEntry[] = rows.map(row => ({
          id: row.id,
          timestamp: row.timestamp,
          level: row.level as LogLevel,
          serviceName: row.serviceName,
          traceId: row.traceId,
          message: row.message,
          tags: JSON.parse(row.tags),
        }));

        resolve(logs);
      }
    );
  });
}

export function getAggregationStats(startTime: number, endTime: number): Promise<AggregationStats> {
  return new Promise((resolve, reject) => {
    const duration = endTime - startTime;
    let granularity = 'day';
    let dateFormat = '%Y-%m-%d';

    if (duration <= 60 * 60 * 1000) {
      granularity = 'minute';
      dateFormat = '%Y-%m-%d %H:%M';
    } else if (duration <= 24 * 60 * 60 * 1000) {
      granularity = 'hour';
      dateFormat = '%Y-%m-%d %H:00';
    }

    const levelCounts: Record<LogLevel, number> = {
      DEBUG: 0,
      INFO: 0,
      WARN: 0,
      ERROR: 0,
    };

    db.all(
      `SELECT level, COUNT(*) as count FROM logs WHERE timestamp >= ? AND timestamp <= ? GROUP BY level`,
      [startTime, endTime],
      (err, levelRows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        levelRows.forEach(row => {
          if (VALID_LOG_LEVELS.includes(row.level as LogLevel)) {
            levelCounts[row.level as LogLevel] = row.count;
          }
        });

        db.all(
          `SELECT serviceName, 
                  COUNT(*) as totalCount, 
                  SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) as errorCount 
           FROM logs 
           WHERE timestamp >= ? AND timestamp <= ? 
           GROUP BY serviceName`,
          [startTime, endTime],
          (err, serviceRows: any[]) => {
            if (err) {
              reject(err);
              return;
            }

            const serviceErrorRates: Record<string, number> = {};
            serviceRows.forEach(row => {
              if (row.totalCount > 0) {
                serviceErrorRates[row.serviceName] = (row.errorCount || 0) / row.totalCount;
              }
            });

            db.all(
              `SELECT message, COUNT(*) as count 
               FROM logs 
               WHERE level = 'ERROR' AND timestamp >= ? AND timestamp <= ? 
               GROUP BY message 
               ORDER BY count DESC 
               LIMIT 10`,
              [startTime, endTime],
              (err, errorRows: any[]) => {
                if (err) {
                  reject(err);
                  return;
                }

                const topErrors = errorRows.map(row => ({
                  message: row.message,
                  count: row.count,
                }));

                db.all(
                  `SELECT strftime('${dateFormat}', datetime(timestamp/1000, 'unixepoch')) as time, 
                          level, 
                          COUNT(*) as count 
                   FROM logs 
                   WHERE timestamp >= ? AND timestamp <= ? 
                   GROUP BY time, level 
                   ORDER BY time ASC`,
                  [startTime, endTime],
                  (err, timelineRows: any[]) => {
                    if (err) {
                      reject(err);
                      return;
                    }

                    const timeline = timelineRows
                      .filter(row => VALID_LOG_LEVELS.includes(row.level as LogLevel))
                      .map(row => ({
                        time: row.time,
                        level: row.level as LogLevel,
                        count: row.count,
                      }));

                    resolve({
                      timeRange: {
                        startTime,
                        endTime,
                        granularity,
                      },
                      levelCounts,
                      serviceErrorRates,
                      topErrors,
                      timeline,
                    });
                  }
                );
              }
            );
          }
        );
      }
    );
  });
}

export function isValidLogLevel(level: string): level is LogLevel {
  return VALID_LOG_LEVELS.includes(level as LogLevel);
}
