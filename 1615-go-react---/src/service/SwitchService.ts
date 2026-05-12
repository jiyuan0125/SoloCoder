import { SwitchRepository } from '../repository/SwitchRepository';
import type {
  CreateSwitchRequest,
  UpdateSwitchRequest,
  EvaluationContext,
  SwitchResponse,
  FullSwitchResponse,
  IncrementalUpdate
} from '../types';
import { validateCreateRequest, validateUpdateRequest, parseJsonValue } from '../utils/validators';
import { evaluateConditions, hasEnabledConditions } from '../utils/conditionEvaluator';
import { ConflictError, NotFoundError, BadRequestError } from '../utils/errors';

export class SwitchService {
  private repo = new SwitchRepository();

  private toResponse(sw: {
    key: string; name: string; description: string; valueType: string;
    booleanValue: boolean; jsonValue: string | null; conditions: any;
    isDeleted: boolean; updatedAt: number; updatedBy: string; version: number;
    id?: number; createdBy?: string; createdAt?: number;
  }): SwitchResponse | FullSwitchResponse {
    const base: SwitchResponse = {
      key: sw.key,
      name: sw.name,
      description: sw.description,
      valueType: sw.valueType as 'boolean' | 'json',
      value: sw.valueType === 'json' && sw.jsonValue ? parseJsonValue(sw.jsonValue) : sw.booleanValue,
      conditions: sw.conditions,
      enabled: sw.valueType === 'json' ? true : sw.booleanValue,
      deleted: sw.isDeleted,
      updatedAt: sw.updatedAt,
      updatedBy: sw.updatedBy,
      version: sw.version
    };
    if (sw.id !== undefined && sw.createdBy && sw.createdAt !== undefined) {
      return { ...base, id: sw.id, createdBy: sw.createdBy, createdAt: sw.createdAt } as FullSwitchResponse;
    }
    return base;
  }

  create(req: Partial<CreateSwitchRequest>): FullSwitchResponse {
    const validated = validateCreateRequest(req);
    const existing = this.repo.findByKey(validated.key);
    if (existing && !existing.isDeleted) {
      throw new ConflictError(`Switch with key "${validated.key}" already exists`);
    }

    const booleanValue = validated.valueType === 'boolean' ? validated.value as boolean : false;
    const jsonValue = validated.valueType === 'json' ? JSON.stringify(validated.value) : null;
    const conditions = JSON.stringify(validated.conditions || {});

    if (existing && existing.isDeleted) {
      const updated = this.repo.update(validated.key, {
        name: validated.name,
        description: validated.description ?? '',
        booleanValue,
        jsonValue,
        conditions,
        operator: validated.operator
      });
      if (!updated) throw new Error('Failed to restore deleted switch');
      return this.toResponse(updated) as FullSwitchResponse;
    }

    const sw = this.repo.create({
      key: validated.key,
      name: validated.name,
      description: validated.description ?? '',
      valueType: validated.valueType,
      booleanValue,
      jsonValue,
      conditions,
      operator: validated.operator
    });
    return this.toResponse(sw) as FullSwitchResponse;
  }

  getByKey(key: string, context?: EvaluationContext, serviceName?: string): SwitchResponse {
    const sw = this.repo.findByKey(key);
    if (!sw) {
      throw new NotFoundError(`Switch "${key}" not found`);
    }

    if (serviceName) {
      this.repo.upsertServiceReference(serviceName, key);
    }

    const response = this.toResponse(sw) as SwitchResponse;

    if (sw.isDeleted) {
      return response;
    }

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

  getAll(): FullSwitchResponse[] {
    const all = this.repo.findAll();
    return all.map(sw => this.toResponse(sw) as FullSwitchResponse);
  }

  update(key: string, req: Partial<UpdateSwitchRequest>): FullSwitchResponse {
    const existing = this.repo.findByKey(key);
    if (!existing) {
      throw new NotFoundError(`Switch "${key}" not found`);
    }

    const validated = validateUpdateRequest(req, existing.valueType);

    const updates: {
      name?: string;
      description?: string;
      booleanValue?: boolean;
      jsonValue?: string | null;
      conditions?: string;
      operator: string;
    } = { operator: validated.operator };

    if (validated.name !== undefined) updates.name = validated.name;
    if (validated.description !== undefined) updates.description = validated.description;
    if (validated.value !== undefined) {
      if (existing.valueType === 'boolean') {
        updates.booleanValue = validated.value as boolean;
      } else {
        updates.jsonValue = JSON.stringify(validated.value);
      }
    }
    if (validated.conditions !== undefined) {
      updates.conditions = JSON.stringify(validated.conditions);
    }

    const updated = this.repo.update(key, updates);
    if (!updated) throw new Error('Update failed');
    return this.toResponse(updated) as FullSwitchResponse;
  }

  delete(key: string, operator: string): void {
    if (!operator || typeof operator !== 'string') {
      throw new BadRequestError('operator is required');
    }
    const existing = this.repo.findByKey(key);
    if (!existing) {
      throw new NotFoundError(`Switch "${key}" not found`);
    }
    if (existing.isDeleted) return;

    if (this.repo.hasActiveReferences(key)) {
      throw new ConflictError(`Cannot delete switch "${key}": there are active service references`, 'ACTIVE_REFERENCES');
    }

    this.repo.softDelete(key, operator);
  }

  batchGet(keys: string[], context?: EvaluationContext, serviceName?: string): {
    results: Record<string, SwitchResponse | null>;
    stale: boolean;
  } {
    if (!Array.isArray(keys)) {
      throw new BadRequestError('keys must be an array');
    }
    const switches = this.repo.findByKeys(keys);
    const byKey = new Map(switches.map(s => [s.key, s]));

    const results: Record<string, SwitchResponse | null> = {};
    for (const key of keys) {
      const sw = byKey.get(key);
      if (!sw) {
        results[key] = null;
        continue;
      }

      if (serviceName) {
        this.repo.upsertServiceReference(serviceName, key);
      }

      const response = this.toResponse(sw) as SwitchResponse;

      if (!sw.isDeleted && context && hasEnabledConditions(sw.conditions)) {
        const shouldEnable = evaluateConditions(sw.conditions, context);
        if (sw.valueType === 'boolean') {
          response.enabled = response.value as boolean && shouldEnable;
        } else {
          response.enabled = shouldEnable;
        }
      }

      results[key] = response;
    }

    return { results, stale: false };
  }

  getIncremental(sinceVersion: number): IncrementalUpdate {
    const currentVersion = this.repo.getCurrentVersion();
    const { updated, deleted } = this.repo.findByVersionAfter(sinceVersion);

    return {
      version: currentVersion,
      switches: updated,
      deletedKeys: deleted.map(s => s.key)
    };
  }

  getCurrentVersion(): number {
    return this.repo.getCurrentVersion();
  }

  getHistory(key: string, limit: number = 50) {
    return this.repo.getHistory(key, limit);
  }
}
