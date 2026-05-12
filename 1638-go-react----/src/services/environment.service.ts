import {
  Environment,
  EnvironmentType,
  EnvironmentStatus,
  EnvironmentConfig,
  ConfigVersion,
  CreateEnvironmentRequest,
  UpdateEnvironmentRequest,
  ApiError,
} from '../types';
import { environmentStore } from '../store';
import { STATE_TRANSITIONS } from '../config/quota';

function validateDestroyed(env: Environment): void {
  if (env.status === 'destroyed') {
    throw new ApiError(400, 'destroyed environment cannot be modified');
  }
}

function validateStateTransition(from: EnvironmentStatus, to: EnvironmentStatus): void {
  const allowed = STATE_TRANSITIONS[from] || [];
  if (!allowed.includes(to)) {
    throw new ApiError(400, `invalid state transition from ${from} to ${to}`);
  }
}

function validateType(type: string): asserts type is EnvironmentType {
  const validTypes: EnvironmentType[] = ['dev', 'test', 'staging', 'prod'];
  if (!validTypes.includes(type as EnvironmentType)) {
    throw new ApiError(400, `invalid environment type: ${type}`);
  }
}

function validateStatus(status: string): asserts status is EnvironmentStatus {
  const validStatuses: EnvironmentStatus[] = ['creating', 'initializing', 'running', 'maintenance', 'destroyed'];
  if (!validStatuses.includes(status as EnvironmentStatus)) {
    throw new ApiError(400, `invalid environment status: ${status}`);
  }
}

function createConfigVersion(config: EnvironmentConfig, version: number): ConfigVersion {
  return {
    version,
    config: { ...config },
    timestamp: new Date().toISOString(),
  };
}

export const environmentService = {
  create(request: CreateEnvironmentRequest): Environment {
    validateType(request.type);

    environmentStore.checkQuota(request.projectId, request.type);

    const initialConfigVersion = 1;
    const newEnv = environmentStore.create({
      name: request.name,
      type: request.type,
      projectId: request.projectId,
      status: 'creating',
      resources: request.resources,
      config: request.config,
      configHistory: [createConfigVersion(request.config, initialConfigVersion)],
      currentConfigVersion: initialConfigVersion,
    });

    return newEnv;
  },

  list(filters?: { projectId?: string; type?: string }): Environment[] {
    if (filters?.type) {
      validateType(filters.type);
    }
    return environmentStore.findAll(filters as { projectId?: string; type?: EnvironmentType });
  },

  getById(id: string): Environment {
    const env = environmentStore.findById(id);
    if (!env) {
      throw new ApiError(404, `environment ${id} not found`);
    }
    return env;
  },

  async update(id: string, request: UpdateEnvironmentRequest): Promise<Environment> {
    return environmentStore.withLock(id, () => {
      const env = environmentService.getById(id);
      validateDestroyed(env);

      const updates: Partial<Environment> = {};

      if (request.name !== undefined) {
        updates.name = request.name;
      }

      if (request.resources) {
        updates.resources = {
          ...env.resources,
          ...request.resources,
          database: request.resources.database
            ? { ...env.resources.database, ...request.resources.database }
            : env.resources.database,
          middleware: request.resources.middleware
            ? { ...env.resources.middleware, ...request.resources.middleware }
            : env.resources.middleware,
        };
      }

      if (request.config) {
        const newVersion = env.currentConfigVersion + 1;
        updates.config = request.config;
        updates.currentConfigVersion = newVersion;
        updates.configHistory = [...env.configHistory, createConfigVersion(request.config, newVersion)];
      }

      const updated = environmentStore.update(id, updates);
      if (!updated) {
        throw new ApiError(404, `environment ${id} not found`);
      }
      return updated;
    });
  },

  async delete(id: string): Promise<void> {
    await environmentService.updateStatus(id, 'destroyed');
  },

  async updateStatus(id: string, newStatus: EnvironmentStatus): Promise<Environment> {
    return environmentStore.withLock(id, () => {
      const env = environmentService.getById(id);
      validateDestroyed(env);
      validateStatus(newStatus);
      validateStateTransition(env.status, newStatus);

      const updated = environmentStore.update(id, { status: newStatus });
      if (!updated) {
        throw new ApiError(404, `environment ${id} not found`);
      }
      return updated;
    });
  },

  async clone(id: string, newName: string): Promise<Environment> {
    return environmentStore.withLock(id, () => {
      const sourceEnv = environmentService.getById(id);

      if (!newName || !newName.trim()) {
        throw new ApiError(400, 'clone environment name is required');
      }

      environmentStore.checkQuota(sourceEnv.projectId, sourceEnv.type);

      const initialConfigVersion = 1;
      const clonedEnv = environmentStore.create({
        name: newName,
        type: sourceEnv.type,
        projectId: sourceEnv.projectId,
        status: 'maintenance',
        resources: {
          serviceInstanceCount: sourceEnv.resources.serviceInstanceCount,
          database: { ...sourceEnv.resources.database },
          middleware: { ...sourceEnv.resources.middleware },
        },
        config: { ...sourceEnv.config },
        configHistory: [createConfigVersion(sourceEnv.config, initialConfigVersion)],
        currentConfigVersion: initialConfigVersion,
      });

      return clonedEnv;
    });
  },

  getConfigHistory(id: string): ConfigVersion[] {
    const env = environmentService.getById(id);
    return env.configHistory;
  },

  async rollback(id: string, version: number): Promise<Environment> {
    return environmentStore.withLock(id, () => {
      const env = environmentService.getById(id);
      validateDestroyed(env);

      const versionEntry = env.configHistory.find(v => v.version === version);
      if (!versionEntry) {
        throw new ApiError(404, `config version ${version} not found`);
      }

      if (env.resources.serviceInstanceCount > 0 && env.status === 'running') {
        throw new ApiError(400, 'active instances detected, manual restart required');
      }

      const newVersion = env.currentConfigVersion + 1;
      const updated = environmentStore.update(id, {
        config: { ...versionEntry.config },
        currentConfigVersion: newVersion,
        configHistory: [...env.configHistory, createConfigVersion(versionEntry.config, newVersion)],
      });

      if (!updated) {
        throw new ApiError(404, `environment ${id} not found`);
      }
      return updated;
    });
  },
};
