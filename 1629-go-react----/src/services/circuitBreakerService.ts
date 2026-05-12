import Database from 'better-sqlite3';
import { getDatabase } from '../database';
import {
  CircuitBreaker,
  CircuitBreakerConfig,
  CircuitBreakerState,
  CreateCircuitBreakerRequest,
  StateTransitionLog,
  TriggerCondition,
  TriggerConditionType,
  UpdateCircuitBreakerRequest,
} from '../types';

function generateId(): string {
  return 'cb_' + Date.now().toString(36) + '_' + Math.random().toString(36).substring(2, 8);
}

export class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ValidationError';
  }
}

export class NotFoundError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NotFoundError';
  }
}

export class InvalidStateTransitionError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'InvalidStateTransitionError';
  }
}

export class CircuitBreakerService {
  private db: Database.Database;

  constructor() {
    this.db = getDatabase();
  }

  validateCreateRequest(req: CreateCircuitBreakerRequest): void {
    if (!req.name || req.name.trim().length === 0) {
      throw new ValidationError('Name is required');
    }
    if (!req.endpoint || req.endpoint.trim().length === 0) {
      throw new ValidationError('Endpoint is required');
    }
    if (!req.triggerConditions || req.triggerConditions.length === 0) {
      throw new ValidationError('At least one trigger condition is required');
    }

    for (const condition of req.triggerConditions) {
      if (!Object.values(TriggerConditionType).includes(condition.type)) {
        throw new ValidationError(`Invalid trigger condition type: ${condition.type}`);
      }
      if (typeof condition.threshold !== 'number' || condition.threshold <= 0) {
        throw new ValidationError('Trigger condition threshold must be a positive number');
      }
      if (condition.type === TriggerConditionType.ERROR_RATE && (condition.threshold < 0 || condition.threshold > 100)) {
        throw new ValidationError('Error rate threshold must be between 0 and 100');
      }
    }

    if (!req.degradationAction) {
      throw new ValidationError('Degradation action is required');
    }

    const validValues = ['DEFAULT_VALUE', 'RETURN_CACHE', 'REJECT'];
    if (!validValues.includes(req.degradationAction as string)) {
      throw new ValidationError(`Invalid degradation action: ${req.degradationAction}`);
    }

    if (req.timeWindowSeconds !== undefined && (typeof req.timeWindowSeconds !== 'number' || req.timeWindowSeconds <= 0)) {
      throw new ValidationError('Time window must be a positive number');
    }

    if (req.openDurationSeconds !== undefined && (typeof req.openDurationSeconds !== 'number' || req.openDurationSeconds <= 0)) {
      throw new ValidationError('Open duration must be a positive number');
    }

    if (req.halfOpenRequestLimit !== undefined && (typeof req.halfOpenRequestLimit !== 'number' || req.halfOpenRequestLimit <= 0)) {
      throw new ValidationError('Half-open request limit must be a positive number');
    }
  }

  validateUpdateRequest(req: UpdateCircuitBreakerRequest): void {
    if (req.name !== undefined && req.name.trim().length === 0) {
      throw new ValidationError('Name cannot be empty');
    }
    if (req.endpoint !== undefined && req.endpoint.trim().length === 0) {
      throw new ValidationError('Endpoint cannot be empty');
    }

    if (req.triggerConditions !== undefined) {
      if (req.triggerConditions.length === 0) {
        throw new ValidationError('At least one trigger condition is required');
      }
      for (const condition of req.triggerConditions) {
        if (!Object.values(TriggerConditionType).includes(condition.type)) {
          throw new ValidationError(`Invalid trigger condition type: ${condition.type}`);
        }
        if (typeof condition.threshold !== 'number' || condition.threshold <= 0) {
          throw new ValidationError('Trigger condition threshold must be a positive number');
        }
        if (condition.type === TriggerConditionType.ERROR_RATE && (condition.threshold < 0 || condition.threshold > 100)) {
          throw new ValidationError('Error rate threshold must be between 0 and 100');
        }
      }
    }

    if (req.degradationAction !== undefined) {
      const validValues = ['DEFAULT_VALUE', 'RETURN_CACHE', 'REJECT'];
      if (!validValues.includes(req.degradationAction as string)) {
        throw new ValidationError(`Invalid degradation action: ${req.degradationAction}`);
      }
    }

    if (req.timeWindowSeconds !== undefined && (typeof req.timeWindowSeconds !== 'number' || req.timeWindowSeconds <= 0)) {
      throw new ValidationError('Time window must be a positive number');
    }

    if (req.openDurationSeconds !== undefined && (typeof req.openDurationSeconds !== 'number' || req.openDurationSeconds <= 0)) {
      throw new ValidationError('Open duration must be a positive number');
    }

    if (req.halfOpenRequestLimit !== undefined && (typeof req.halfOpenRequestLimit !== 'number' || req.halfOpenRequestLimit <= 0)) {
      throw new ValidationError('Half-open request limit must be a positive number');
    }
  }

  createCircuitBreaker(req: CreateCircuitBreakerRequest): { config: CircuitBreakerConfig; circuitBreaker: CircuitBreaker } {
    this.validateCreateRequest(req);

    const now = new Date();
    const configId = generateId();
    const circuitBreakerId = generateId();

    const config: CircuitBreakerConfig = {
      id: configId,
      name: req.name,
      endpoint: req.endpoint,
      triggerConditions: req.triggerConditions,
      degradationAction: req.degradationAction,
      defaultValue: req.defaultValue,
      timeWindowSeconds: req.timeWindowSeconds || 60,
      openDurationSeconds: req.openDurationSeconds || 30,
      halfOpenRequestLimit: req.halfOpenRequestLimit || 5,
      createdAt: now,
      updatedAt: now,
    };

    const circuitBreaker: CircuitBreaker = {
      id: circuitBreakerId,
      configId: configId,
      state: CircuitBreakerState.CLOSED,
      stateChangedAt: now,
    };

    const stmtConfig = this.db.prepare(`
      INSERT INTO circuit_breaker_configs (
        id, name, endpoint, trigger_conditions, degradation_action, default_value,
        time_window_seconds, open_duration_seconds, half_open_request_limit,
        created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    stmtConfig.run(
      config.id,
      config.name,
      config.endpoint,
      JSON.stringify(config.triggerConditions),
      config.degradationAction,
      config.defaultValue !== undefined ? JSON.stringify(config.defaultValue) : null,
      config.timeWindowSeconds,
      config.openDurationSeconds,
      config.halfOpenRequestLimit,
      config.createdAt.toISOString(),
      config.updatedAt.toISOString()
    );

    const stmtBreaker = this.db.prepare(`
      INSERT INTO circuit_breakers (id, config_id, state, state_changed_at)
      VALUES (?, ?, ?, ?)
    `);

    stmtBreaker.run(
      circuitBreaker.id,
      circuitBreaker.configId,
      circuitBreaker.state,
      circuitBreaker.stateChangedAt.toISOString()
    );

    return { config, circuitBreaker };
  }

  updateCircuitBreaker(id: string, req: UpdateCircuitBreakerRequest): { config: CircuitBreakerConfig; circuitBreaker: CircuitBreaker } {
    this.validateUpdateRequest(req);

    const result = this.db
      .prepare(
        `SELECT cb.*, cbc.* FROM circuit_breakers cb
         JOIN circuit_breaker_configs cbc ON cb.config_id = cbc.id
         WHERE cb.id = ?`
      )
      .get(id) as any;

    if (!result) {
      throw new NotFoundError(`Circuit breaker with id ${id} not found`);
    }

    const now = new Date();
    const existingConfig = this.rowToConfig(result);

    const updatedConfig: CircuitBreakerConfig = {
      ...existingConfig,
      name: req.name ?? existingConfig.name,
      endpoint: req.endpoint ?? existingConfig.endpoint,
      triggerConditions: req.triggerConditions ?? existingConfig.triggerConditions,
      degradationAction: req.degradationAction ?? existingConfig.degradationAction,
      defaultValue: req.defaultValue !== undefined ? req.defaultValue : existingConfig.defaultValue,
      timeWindowSeconds: req.timeWindowSeconds ?? existingConfig.timeWindowSeconds,
      openDurationSeconds: req.openDurationSeconds ?? existingConfig.openDurationSeconds,
      halfOpenRequestLimit: req.halfOpenRequestLimit ?? existingConfig.halfOpenRequestLimit,
      updatedAt: now,
    };

    const stmt = this.db.prepare(`
      UPDATE circuit_breaker_configs
      SET name = ?, endpoint = ?, trigger_conditions = ?, degradation_action = ?,
          default_value = ?, time_window_seconds = ?, open_duration_seconds = ?,
          half_open_request_limit = ?, updated_at = ?
      WHERE id = ?
    `);

    stmt.run(
      updatedConfig.name,
      updatedConfig.endpoint,
      JSON.stringify(updatedConfig.triggerConditions),
      updatedConfig.degradationAction,
      updatedConfig.defaultValue !== undefined ? JSON.stringify(updatedConfig.defaultValue) : null,
      updatedConfig.timeWindowSeconds,
      updatedConfig.openDurationSeconds,
      updatedConfig.halfOpenRequestLimit,
      updatedConfig.updatedAt.toISOString(),
      updatedConfig.id
    );

    const circuitBreaker = this.rowToCircuitBreaker(result);
    return { config: updatedConfig, circuitBreaker };
  }

  getCircuitBreaker(id: string): { config: CircuitBreakerConfig; circuitBreaker: CircuitBreaker } | null {
    const result = this.db
      .prepare(
        `SELECT cb.*, cbc.* FROM circuit_breakers cb
         JOIN circuit_breaker_configs cbc ON cb.config_id = cbc.id
         WHERE cb.id = ?`
      )
      .get(id) as any;

    if (!result) {
      return null;
    }

    return {
      config: this.rowToConfig(result),
      circuitBreaker: this.rowToCircuitBreaker(result),
    };
  }

  listCircuitBreakers(): Array<{ config: CircuitBreakerConfig; circuitBreaker: CircuitBreaker }> {
    const results = this.db
      .prepare(
        `SELECT cb.*, cbc.* FROM circuit_breakers cb
         JOIN circuit_breaker_configs cbc ON cb.config_id = cbc.id
         ORDER BY cbc.created_at DESC`
      )
      .all() as any[];

    return results.map((row) => ({
      config: this.rowToConfig(row),
      circuitBreaker: this.rowToCircuitBreaker(row),
    }));
  }

  deleteCircuitBreaker(id: string): boolean {
    const result = this.getCircuitBreaker(id);
    if (!result) {
      return false;
    }

    const transaction = this.db.transaction(() => {
      this.db.prepare('DELETE FROM request_stats WHERE circuit_breaker_id = ?').run(id);
      this.db.prepare('DELETE FROM state_transition_logs WHERE circuit_breaker_id = ?').run(id);
      this.db.prepare('DELETE FROM circuit_breakers WHERE id = ?').run(id);
      this.db.prepare('DELETE FROM circuit_breaker_configs WHERE id = ?').run(result.config.id);
    });

    transaction();
    return true;
  }

  canTransitionState(from: CircuitBreakerState, to: CircuitBreakerState): boolean {
    const allowedTransitions: Record<CircuitBreakerState, CircuitBreakerState[]> = {
      [CircuitBreakerState.CLOSED]: [CircuitBreakerState.OPEN],
      [CircuitBreakerState.OPEN]: [CircuitBreakerState.HALF_OPEN],
      [CircuitBreakerState.HALF_OPEN]: [CircuitBreakerState.CLOSED, CircuitBreakerState.OPEN],
    };

    return allowedTransitions[from]?.includes(to) || false;
  }

  transitionState(id: string, toState: CircuitBreakerState, reason: string): CircuitBreaker {
    const result = this.db
      .prepare(
        `SELECT cb.*, cbc.* FROM circuit_breakers cb
         JOIN circuit_breaker_configs cbc ON cb.config_id = cbc.id
         WHERE cb.id = ?`
      )
      .get(id) as any;

    if (!result) {
      throw new NotFoundError(`Circuit breaker with id ${id} not found`);
    }

    const circuitBreaker = this.rowToCircuitBreaker(result);
    const fromState = circuitBreaker.state;

    if (fromState === toState) {
      console.log(`[State Transition Log] Circuit breaker ${id} is already in state ${toState}. Duplicate trigger logged. Reason: ${reason}`);
      return circuitBreaker;
    }

    if (!this.canTransitionState(fromState, toState)) {
      throw new InvalidStateTransitionError(
        `Invalid state transition: ${fromState} -> ${toState}`
      );
    }

    const now = new Date();
    const transaction = this.db.transaction(() => {
      this.db
        .prepare(
          'UPDATE circuit_breakers SET state = ?, state_changed_at = ?, last_open_reason = ? WHERE id = ?'
        )
        .run(
          toState,
          now.toISOString(),
          toState === CircuitBreakerState.OPEN ? reason : circuitBreaker.lastOpenReason || null,
          id
        );

      this.db
        .prepare(
          `INSERT INTO state_transition_logs (circuit_breaker_id, from_state, to_state, reason, timestamp)
           VALUES (?, ?, ?, ?, ?)`
        )
        .run(id, fromState, toState, reason, now.toISOString());
    });

    transaction();

    console.log(
      `[State Transition Log] Circuit breaker ${id} transitioned: ${fromState} -> ${toState}. Reason: ${reason}. Timestamp: ${now.toISOString()}`
    );

    return {
      ...circuitBreaker,
      state: toState,
      stateChangedAt: now,
      lastOpenReason: toState === CircuitBreakerState.OPEN ? reason : circuitBreaker.lastOpenReason,
    };
  }

  getStateTransitionLogs(circuitBreakerId: string): StateTransitionLog[] {
    const logs = this.db
      .prepare(
        `SELECT * FROM state_transition_logs
         WHERE circuit_breaker_id = ?
         ORDER BY timestamp DESC`
      )
      .all(circuitBreakerId) as any[];

    return logs.map((row) => ({
      id: row.id,
      circuitBreakerId: row.circuit_breaker_id,
      fromState: row.from_state as CircuitBreakerState,
      toState: row.to_state as CircuitBreakerState,
      reason: row.reason,
      timestamp: new Date(row.timestamp),
    }));
  }

  private rowToConfig(row: any): CircuitBreakerConfig {
    return {
      id: row.id,
      name: row.name,
      endpoint: row.endpoint,
      triggerConditions: JSON.parse(row.trigger_conditions),
      degradationAction: row.degradation_action,
      defaultValue: row.default_value !== null ? JSON.parse(row.default_value) : undefined,
      timeWindowSeconds: row.time_window_seconds,
      openDurationSeconds: row.open_duration_seconds,
      halfOpenRequestLimit: row.half_open_request_limit,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at),
    };
  }

  private rowToCircuitBreaker(row: any): CircuitBreaker {
    return {
      id: row.id,
      configId: row.config_id,
      state: row.state as CircuitBreakerState,
      stateChangedAt: new Date(row.state_changed_at),
      lastOpenReason: row.last_open_reason || undefined,
    };
  }
}
