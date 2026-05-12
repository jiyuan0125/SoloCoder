import { EntityConfig } from '../models/types';

const entityConfigs: Map<string, EntityConfig> = new Map();
const configByCode: Map<string, EntityConfig> = new Map();

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
}

export function createEntityConfig(
  type: EntityConfig['type'],
  code: string,
  name: string,
  config: Record<string, unknown>
): EntityConfig {
  if (configByCode.has(`${type}:${code}`)) {
    throw new Error(`配置代码 ${code} 已存在于类型 ${type}`);
  }

  const id = generateId();
  const entityConfig: EntityConfig = {
    id,
    type,
    code,
    name,
    config,
    active: true,
    createdAt: new Date(),
  };

  entityConfigs.set(id, entityConfig);
  configByCode.set(`${type}:${code}`, entityConfig);
  return entityConfig;
}

export function getEntityConfig(id: string): EntityConfig | undefined {
  return entityConfigs.get(id);
}

export function getEntityConfigByCode(
  type: EntityConfig['type'],
  code: string
): EntityConfig | undefined {
  return configByCode.get(`${type}:${code}`);
}

export function getEntityConfigsByType(
  type: EntityConfig['type']
): EntityConfig[] {
  return Array.from(entityConfigs.values())
    .filter(c => c.type === type && c.active)
    .sort((a, b) => a.code.localeCompare(b.code));
}

export function getAllEntityConfigs(): EntityConfig[] {
  return Array.from(entityConfigs.values());
}

export function updateEntityConfig(
  id: string,
  updates: Partial<Pick<EntityConfig, 'name' | 'config' | 'active'>>
): EntityConfig | undefined {
  const config = entityConfigs.get(id);
  if (!config) return undefined;

  if (updates.name !== undefined) {
    config.name = updates.name;
  }
  if (updates.config !== undefined) {
    config.config = { ...config.config, ...updates.config };
  }
  if (updates.active !== undefined) {
    config.active = updates.active;
  }

  entityConfigs.set(id, config);
  configByCode.set(`${config.type}:${config.code}`, config);
  return config;
}

export function deleteEntityConfig(id: string): boolean {
  const config = entityConfigs.get(id);
  if (!config) return false;

  entityConfigs.delete(id);
  configByCode.delete(`${config.type}:${config.code}`);
  return true;
}

export function initDefaultEntityConfigs(): void {
  createEntityConfig(
    'chronic_disease',
    'HTN',
    '高血压',
    {
      thresholds: {
        systolicBPMax: 140,
        diastolicBPMax: 90,
      },
      followUpTemplate: 'HTN_FOLLOWUP',
    }
  );

  createEntityConfig(
    'chronic_disease',
    'DM',
    '糖尿病',
    {
      thresholds: {
        fastingBloodSugarMax: 7.0,
        HbA1cMax: 7.0,
      },
      followUpTemplate: 'DM_FOLLOWUP',
    }
  );

  createEntityConfig(
    'chronic_disease',
    'HLP',
    '高血脂',
    {
      thresholds: {
        LDLMax: 3.4,
        triglyceridesMax: 1.7,
      },
      followUpTemplate: 'HLP_FOLLOWUP',
    }
  );

  createEntityConfig(
    'medication',
    'ANTIHYPERTENSIVE',
    '抗高血压药',
    {
      category: 'cardiovascular',
      commonTypes: ['ACEI', 'ARB', 'CCB', 'BetaBlocker', 'Diuretic'],
    }
  );

  createEntityConfig(
    'medication',
    'ANTIDIABETIC',
    '抗糖尿病药',
    {
      category: 'endocrine',
      commonTypes: ['Metformin', 'Sulfonylurea', 'Insulin', 'GLP1'],
    }
  );

  createEntityConfig(
    'followup_template',
    'HTN_FOLLOWUP',
    '高血压随访模板',
    {
      requiredMetrics: ['systolicBP', 'diastolicBP', 'weight'],
      symptoms: ['头痛', '头晕', '心悸', '胸闷'],
      lifestyleAspects: ['smoking', 'alcohol', 'exercise', 'diet'],
    }
  );

  createEntityConfig(
    'followup_template',
    'DM_FOLLOWUP',
    '糖尿病随访模板',
    {
      requiredMetrics: ['fastingBloodSugar', 'HbA1c', 'weight'],
      symptoms: ['多饮', '多尿', '多食', '乏力', '视力模糊'],
      lifestyleAspects: ['smoking', 'alcohol', 'exercise', 'diet', 'sleepHours'],
    }
  );
}
