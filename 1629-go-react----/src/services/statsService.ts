import Database from 'better-sqlite3';
import { getDatabase } from '../database';
import { RequestStats, TriggerConditionType } from '../types';
import { NotFoundError } from './circuitBreakerService';

export interface StatsResponse {
  circuitBreakerId: string;
  minute: string;
  successCount: number;
  failureCount: number;
  totalCount: number;
  errorRate: number;
  averageResponseTimeMs: number;
}

export class StatsService {
  private db: Database.Database;

  constructor() {
    this.db = getDatabase();
  }

  private getMinuteKey(date: Date): string {
    const d = new Date(date);
    d.setSeconds(0, 0);
    return d.toISOString();
  }

  recordRequest(
    circuitBreakerId: string,
    isSuccess: boolean,
    responseTimeMs: number
  ): void {
    const minute = this.getMinuteKey(new Date());

    const existing = this.db
      .prepare(
        `SELECT * FROM request_stats
         WHERE circuit_breaker_id = ? AND minute = ?`
      )
      .get(circuitBreakerId, minute) as any;

    if (existing) {
      this.db
        .prepare(
          `UPDATE request_stats
           SET success_count = success_count + ?,
               failure_count = failure_count + ?,
               total_response_time_ms = total_response_time_ms + ?
           WHERE circuit_breaker_id = ? AND minute = ?`
        )
        .run(
          isSuccess ? 1 : 0,
          isSuccess ? 0 : 1,
          responseTimeMs,
          circuitBreakerId,
          minute
        );
    } else {
      this.db
        .prepare(
          `INSERT INTO request_stats
           (circuit_breaker_id, minute, success_count, failure_count, total_response_time_ms)
           VALUES (?, ?, ?, ?, ?)`
        )
        .run(
          circuitBreakerId,
          minute,
          isSuccess ? 1 : 0,
          isSuccess ? 0 : 1,
          responseTimeMs
        );
    }
  }

  getStatsForCircuitBreaker(circuitBreakerId: string): StatsResponse[] {
    const exists = this.db
      .prepare('SELECT id FROM circuit_breakers WHERE id = ?')
      .get(circuitBreakerId);

    if (!exists) {
      throw new NotFoundError(`Circuit breaker with id ${circuitBreakerId} not found`);
    }

    const stats = this.db
      .prepare(
        `SELECT * FROM request_stats
         WHERE circuit_breaker_id = ?
         ORDER BY minute DESC`
      )
      .all(circuitBreakerId) as any[];

    return stats.map((row) => this.mapToResponse(row));
  }

  getStatsForMinute(circuitBreakerId: string, minute: Date): StatsResponse | null {
    const minuteKey = this.getMinuteKey(minute);
    const row = this.db
      .prepare(
        `SELECT * FROM request_stats
         WHERE circuit_breaker_id = ? AND minute = ?`
      )
      .get(circuitBreakerId, minuteKey) as any;

    if (!row) {
      return null;
    }

    return this.mapToResponse(row);
  }

  getStatsForTimeWindow(
    circuitBreakerId: string,
    windowSeconds: number
  ): {
    successCount: number;
    failureCount: number;
    totalCount: number;
    errorRate: number;
    averageResponseTimeMs: number;
  } {
    const cutoffDate = new Date();
    cutoffDate.setSeconds(cutoffDate.getSeconds() - windowSeconds);
    const cutoffMinute = this.getMinuteKey(cutoffDate);

    const result = this.db
      .prepare(
        `SELECT
           SUM(success_count) as success_count,
           SUM(failure_count) as failure_count,
           SUM(total_response_time_ms) as total_response_time_ms
         FROM request_stats
         WHERE circuit_breaker_id = ? AND minute >= ?`
      )
      .get(circuitBreakerId, cutoffMinute) as any;

    const successCount = result.success_count || 0;
    const failureCount = result.failure_count || 0;
    const totalCount = successCount + failureCount;
    const totalTime = result.total_response_time_ms || 0;

    return {
      successCount,
      failureCount,
      totalCount,
      errorRate: totalCount > 0 ? (failureCount / totalCount) * 100 : 0,
      averageResponseTimeMs: totalCount > 0 ? totalTime / totalCount : 0,
    };
  }

  shouldTriggerOpen(
    circuitBreakerId: string,
    windowSeconds: number,
    triggerConditions: Array<{ type: TriggerConditionType; threshold: number }>
  ): { shouldTrigger: boolean; reason?: string } {
    const stats = this.getStatsForTimeWindow(circuitBreakerId, windowSeconds);

    if (stats.totalCount === 0) {
      return { shouldTrigger: false };
    }

    for (const condition of triggerConditions) {
      if (condition.type === TriggerConditionType.ERROR_RATE) {
        if (stats.errorRate >= condition.threshold) {
          return {
            shouldTrigger: true,
            reason: `Error rate ${stats.errorRate.toFixed(2)}% exceeds threshold ${condition.threshold}%`,
          };
        }
      } else if (condition.type === TriggerConditionType.RESPONSE_TIME) {
        if (stats.averageResponseTimeMs >= condition.threshold) {
          return {
            shouldTrigger: true,
            reason: `Average response time ${stats.averageResponseTimeMs.toFixed(2)}ms exceeds threshold ${condition.threshold}ms`,
          };
        }
      }
    }

    return { shouldTrigger: false };
  }

  private mapToResponse(row: any): StatsResponse {
    const successCount = row.success_count;
    const failureCount = row.failure_count;
    const totalCount = successCount + failureCount;

    return {
      circuitBreakerId: row.circuit_breaker_id,
      minute: row.minute,
      successCount,
      failureCount,
      totalCount,
      errorRate: totalCount > 0 ? (failureCount / totalCount) * 100 : 0,
      averageResponseTimeMs:
        totalCount > 0 ? row.total_response_time_ms / totalCount : 0,
    };
  }
}
