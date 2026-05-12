import { Environment, EnvironmentType, ApiError } from '../types';
import { ENVIRONMENT_QUOTAS } from '../config/quota';

class EnvironmentStore {
  private environments: Map<string, Environment> = new Map();
  private idCounter: number = 1;

  generateId(): string {
    return `env-${this.idCounter++}`;
  }

  create(environment: Omit<Environment, 'id' | 'isLocked' | 'createdAt' | 'updatedAt'>): Environment {
    const id = this.generateId();
    const now = new Date().toISOString();
    const newEnv: Environment = {
      ...environment,
      id,
      isLocked: false,
      createdAt: now,
      updatedAt: now,
    };
    this.environments.set(id, newEnv);
    return newEnv;
  }

  findAll(filters?: { projectId?: string; type?: EnvironmentType }): Environment[] {
    let result = Array.from(this.environments.values());
    if (filters?.projectId) {
      result = result.filter(e => e.projectId === filters.projectId);
    }
    if (filters?.type) {
      result = result.filter(e => e.type === filters.type);
    }
    return result;
  }

  findById(id: string): Environment | undefined {
    return this.environments.get(id);
  }

  update(id: string, updates: Partial<Environment>): Environment | undefined {
    const env = this.environments.get(id);
    if (!env) {
      return undefined;
    }
    const updatedEnv: Environment = {
      ...env,
      ...updates,
      updatedAt: new Date().toISOString(),
    };
    this.environments.set(id, updatedEnv);
    return updatedEnv;
  }

  delete(id: string): boolean {
    return this.environments.delete(id);
  }

  countByProjectAndType(projectId: string, type: EnvironmentType): number {
    let count = 0;
    for (const env of this.environments.values()) {
      if (env.projectId === projectId && env.type === type && env.status !== 'destroyed') {
        count++;
      }
    }
    return count;
  }

  checkQuota(projectId: string, type: EnvironmentType): void {
    const current = this.countByProjectAndType(projectId, type);
    const limit = ENVIRONMENT_QUOTAS[type];
    if (current >= limit) {
      throw new ApiError(403, `environment quota exceeded for project ${projectId} type ${type}`);
    }
  }

  acquireLock(id: string): void {
    const env = this.findById(id);
    if (!env) {
      throw new ApiError(404, `environment ${id} not found`);
    }
    if (env.isLocked) {
      throw new ApiError(409, `another change is in progress for environment ${id}`);
    }
    this.update(id, { isLocked: true });
  }

  releaseLock(id: string): void {
    this.update(id, { isLocked: false });
  }

  async withLock<T>(id: string, operation: () => T | Promise<T>): Promise<T> {
    this.acquireLock(id);
    try {
      const result = await operation();
      return result;
    } finally {
      this.releaseLock(id);
    }
  }
}

export const environmentStore = new EnvironmentStore();
