import type { FeatureSwitch, IncrementalUpdate, EvaluationContext, SwitchResponse } from '../types';
import { parseJsonValue } from '../utils/validators';
import { evaluateConditions, hasEnabledConditions } from '../utils/conditionEvaluator';

export interface ConfigClientOptions {
  pollIntervalMs?: number;
  maxFailedAttempts?: number;
}

interface ClientSwitchState {
  switch: FeatureSwitch;
}

export class FeatureSwitchClient {
  private cache = new Map<string, ClientSwitchState>();
  private currentVersion = 0;
  private consecutiveFailures = 0;
  private isStale = false;
  private timer: NodeJS.Timeout | null = null;
  private pollIntervalMs: number;
  private maxFailedAttempts: number;
  private serviceName: string;

  private fetchAll: () => Promise<{ switches: FeatureSwitch[]; version: number }>;
  private fetchIncremental: (sinceVersion: number) => Promise<IncrementalUpdate>;

  constructor(
    serviceName: string,
    fetchAll: () => Promise<{ switches: FeatureSwitch[]; version: number }>,
    fetchIncremental: (sinceVersion: number) => Promise<IncrementalUpdate>,
    options?: ConfigClientOptions
  ) {
    this.serviceName = serviceName;
    this.fetchAll = fetchAll;
    this.fetchIncremental = fetchIncremental;
    this.pollIntervalMs = options?.pollIntervalMs ?? 60000;
    this.maxFailedAttempts = options?.maxFailedAttempts ?? 3;
  }

  async initialize(): Promise<void> {
    await this.fullSync();
    this.startPolling();
  }

  private async fullSync(): Promise<void> {
    try {
      const result = await this.fetchAll();
      this.cache.clear();
      for (const sw of result.switches) {
        if (!sw.isDeleted) {
          this.cache.set(sw.key, { switch: sw });
        }
      }
      this.currentVersion = result.version;
      this.consecutiveFailures = 0;
      this.isStale = false;
    } catch (err) {
      this.consecutiveFailures++;
      if (this.consecutiveFailures > this.maxFailedAttempts) {
        this.isStale = true;
      }
      throw err;
    }
  }

  private async incrementalSync(): Promise<void> {
    try {
      const update = await this.fetchIncremental(this.currentVersion);
      for (const sw of update.switches) {
        if (sw.isDeleted) {
          this.cache.delete(sw.key);
        } else {
          this.cache.set(sw.key, { switch: sw });
        }
      }
      for (const key of update.deletedKeys) {
        this.cache.delete(key);
      }
      this.currentVersion = update.version;
      this.consecutiveFailures = 0;
      this.isStale = false;
    } catch (err) {
      this.consecutiveFailures++;
      if (this.consecutiveFailures > this.maxFailedAttempts) {
        this.isStale = true;
      }
      throw err;
    }
  }

  private startPolling(): void {
    this.stopPolling();
    this.timer = setInterval(async () => {
      try {
        await this.incrementalSync();
      } catch (err) {
        console.error('Incremental sync failed:', err);
      }
    }, this.pollIntervalMs);
  }

  stopPolling(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  getStaleStatus(): boolean {
    return this.isStale;
  }

  getVersion(): number {
    return this.currentVersion;
  }

  getAllKeys(): string[] {
    return Array.from(this.cache.keys());
  }

  evaluate(key: string, context?: EvaluationContext): SwitchResponse | null {
    const state = this.cache.get(key);
    if (!state) {
      return null;
    }

    const sw = state.switch;
    const response: SwitchResponse = {
      key: sw.key,
      name: sw.name,
      description: sw.description,
      valueType: sw.valueType,
      value: sw.valueType === 'json' && sw.jsonValue ? parseJsonValue(sw.jsonValue) : sw.booleanValue,
      conditions: sw.conditions,
      enabled: sw.valueType === 'json' ? true : sw.booleanValue,
      deleted: sw.isDeleted,
      updatedAt: sw.updatedAt,
      updatedBy: sw.updatedBy,
      version: sw.version
    };

    if (context && hasEnabledConditions(sw.conditions)) {
      const shouldEnable = evaluateConditions(sw.conditions, context);
      if (sw.valueType === 'boolean') {
        response.enabled = response.value as boolean && shouldEnable;
      } else {
        response.enabled = shouldEnable;
      }
    }

    return response;
  }

  evaluateBatch(keys: string[], context?: EvaluationContext): {
    results: Record<string, SwitchResponse | null>;
    stale: boolean;
  } {
    const results: Record<string, SwitchResponse | null> = {};
    for (const key of keys) {
      results[key] = this.evaluate(key, context);
    }
    return { results, stale: this.isStale };
  }
}
