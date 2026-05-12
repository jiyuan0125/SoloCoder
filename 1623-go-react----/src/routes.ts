import express, { Request, Response } from 'express';
import {
  getAppByName,
  getAppById,
  getAllApps,
  createApp,
  updateApp,
  deleteApp,
  getConfig,
  getConfigById,
  getConfigsByAppAndEnv,
  getConfigsByApp,
  createConfig as dbCreateConfig,
  updateConfig as dbUpdateConfig,
  deleteConfig as dbDeleteConfig,
  getConfigHistory,
  getConfigHistoryByVersion,
  rollbackConfig as dbRollbackConfig,
  ConfigDb,
  db
} from './database';
import {
  Environment,
  ValueType,
  AppCreateRequest,
  AppUpdateRequest,
  ConfigCreateRequest,
  ConfigUpdateRequest,
  BatchConfigRequest,
  BatchPublishResponse
} from './types';
import {
  validateEnvironment,
  validateValueType,
  validateValue,
  calculateMaxVersion
} from './utils';
import {
  notifyConfigChange,
  disconnectByApp,
  disconnectByConfig
} from './polling';

const router = express.Router();

router.use(express.json());

router.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

router.get('/apps', (req: Request, res: Response) => {
  const apps = getAllApps();
  res.json(apps);
});

router.post('/apps', (req: Request, res: Response) => {
  const { name, description } = req.body as AppCreateRequest;

  if (!name || typeof name !== 'string' || name.trim() === '') {
    return res.status(400).json({ error: 'App name is required' });
  }

  if (getAppByName(name.trim())) {
    return res.status(409).json({ error: `App '${name}' already exists` });
  }

  const id = createApp(name.trim(), description || '');
  const app = getAppById(id);
  res.status(201).json(app);
});

router.get('/apps/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(id);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  res.json(app);
});

router.put('/apps/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(id);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const { name, description } = req.body as AppUpdateRequest;

  if (name !== undefined) {
    if (typeof name !== 'string' || name.trim() === '') {
      return res.status(400).json({ error: 'App name cannot be empty' });
    }
    const existing = getAppByName(name.trim());
    if (existing && existing.id !== id) {
      return res.status(409).json({ error: `App '${name}' already exists` });
    }
  }

  updateApp(id, name?.trim(), description);
  const updated = getAppById(id);
  res.json(updated);
});

router.delete('/apps/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(id);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  disconnectByApp(id);
  deleteApp(id);
  res.status(204).end();
});

router.get('/apps/:appId/configs', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (environment) {
    const envValidation = validateEnvironment(environment);
    if (!envValidation.valid) {
      return res.status(400).json({ error: envValidation.error });
    }
    const configs = getConfigsByAppAndEnv(appId, environment);
    return res.json(configs);
  }

  const configs = getConfigsByApp(appId);
  res.json(configs);
});

router.post('/apps/:appId/configs', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const { environment, key, value, valueType, description, operator } = req.body as ConfigCreateRequest;

  if (!environment || !key || !value || !valueType || !operator) {
    return res.status(400).json({ error: 'environment, key, value, valueType, and operator are required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const typeValidation = validateValueType(valueType);
  if (!typeValidation.valid) {
    return res.status(400).json({ error: typeValidation.error });
  }

  const valueValidation = validateValue(value, valueType);
  if (!valueValidation.valid) {
    return res.status(400).json({
      error: valueValidation.error,
      errorPosition: valueValidation.errorPosition
    });
  }

  if (getConfig(appId, environment, key.trim())) {
    return res.status(409).json({ error: `Config '${key}' already exists in ${environment} environment` });
  }

  const configId = dbCreateConfig(
    appId,
    environment,
    key.trim(),
    value,
    valueType,
    description || '',
    operator
  );

  notifyConfigChange(appId, environment, 1);

  const config = getConfigById(configId);
  res.status(201).json(config);
});

router.get('/apps/:appId/configs/:key', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const config = getConfig(appId, environment, req.params.key);
  if (!config) {
    return res.status(404).json({ error: 'Config not found' });
  }

  res.json(config);
});

router.put('/apps/:appId/configs/:key', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const config = getConfig(appId, environment, req.params.key);
  if (!config) {
    return res.status(404).json({ error: 'Config not found' });
  }

  const { value, valueType, description, operator } = req.body as ConfigUpdateRequest;

  if (!operator) {
    return res.status(400).json({ error: 'operator is required' });
  }

  const newValue = value !== undefined ? value : config.value;
  const newValueType = valueType !== undefined ? valueType : config.valueType as ValueType;
  const newDescription = description !== undefined ? description : config.description;

  if (valueType !== undefined) {
    const typeValidation = validateValueType(valueType);
    if (!typeValidation.valid) {
      return res.status(400).json({ error: typeValidation.error });
    }
  }

  const valueValidation = validateValue(newValue, newValueType);
  if (!valueValidation.valid) {
    return res.status(400).json({
      error: valueValidation.error,
      errorPosition: valueValidation.errorPosition
    });
  }

  dbUpdateConfig(config.id, newValue, newValueType, newDescription, operator);

  const updatedConfig = getConfigById(config.id);
  if (updatedConfig) {
    notifyConfigChange(appId, environment, updatedConfig.version);
  }

  res.json(updatedConfig);
});

router.delete('/apps/:appId/configs/:key', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const config = getConfig(appId, environment, req.params.key);
  if (!config) {
    return res.status(404).json({ error: 'Config not found' });
  }

  const operator = req.query.operator as string | undefined;
  if (!operator) {
    return res.status(400).json({ error: 'operator query parameter is required' });
  }

  disconnectByConfig(appId, environment, req.params.key);
  dbDeleteConfig(config.id, operator);

  res.status(204).end();
});

router.get('/apps/:appId/configs/:key/history', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const config = getConfig(appId, environment, req.params.key);
  if (!config) {
    return res.status(404).json({ error: 'Config not found' });
  }

  const history = getConfigHistory(config.id);
  res.json(history);
});

router.post('/apps/:appId/configs/batch-publish', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const { environment, items, operator } = req.body as BatchConfigRequest;

  if (!environment || !items || !Array.isArray(items) || items.length === 0 || !operator) {
    return res.status(400).json({ error: 'environment, items (non-empty array), and operator are required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const failedItems: { index: number; key: string; error: string }[] = [];

  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    const { key, value, valueType } = item;

    if (!key || value === undefined) {
      failedItems.push({ index: i, key: key || `item-${i}`, error: 'key and value are required' });
      continue;
    }

    if (valueType) {
      const typeValidation = validateValueType(valueType);
      if (!typeValidation.valid) {
        failedItems.push({ index: i, key, error: typeValidation.error || 'Invalid value type' });
        continue;
      }
    }

    const existingConfig = getConfig(appId, environment, key.trim());
    const actualValueType = valueType || (existingConfig?.valueType as ValueType) || 'string';

    const valueValidation = validateValue(value, actualValueType);
    if (!valueValidation.valid) {
      failedItems.push({
        index: i,
        key,
        error: valueValidation.errorPosition !== undefined
          ? `${valueValidation.error} at position ${valueValidation.errorPosition}`
          : valueValidation.error || 'Invalid value'
      });
    }
  }

  if (failedItems.length > 0) {
    return res.status(400).json({
      success: false,
      failedItems
    } as BatchPublishResponse);
  }

  const versions: number[] = [];

  const transaction = db.transaction(() => {
    for (const item of items) {
      const { key, value, valueType, description } = item;
      const existingConfig = getConfig(appId, environment, key.trim());

      if (existingConfig) {
        const newDescription = description !== undefined ? description : existingConfig.description;
        const newType = valueType || existingConfig.valueType as ValueType;
        dbUpdateConfig(existingConfig.id, value, newType, newDescription, operator);
        const updated = getConfigById(existingConfig.id);
        if (updated) {
          versions.push(updated.version);
        }
      } else {
        const configId = dbCreateConfig(
          appId,
          environment,
          key.trim(),
          value,
          valueType || 'string',
          description || '',
          operator
        );
        const created = getConfigById(configId);
        if (created) {
          versions.push(created.version);
        }
      }
    }
  });

  try {
    transaction();
    const maxVersion = calculateMaxVersion(versions);
    notifyConfigChange(appId, environment, maxVersion);
    res.json({ success: true } as BatchPublishResponse);
  } catch (error) {
    res.status(500).json({ error: 'Batch publish failed', details: (error as Error).message });
  }
});

router.get('/apps/:appId/configs/pull', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const versionParam = req.query.version;
  const clientVersion = versionParam ? parseInt(versionParam as string) : undefined;

  const configs = getConfigsByAppAndEnv(appId, environment);
  const currentMaxVersion = calculateMaxVersion(configs.map(c => c.version));

  if (clientVersion === undefined || currentMaxVersion > clientVersion) {
    return res.json({
      version: currentMaxVersion,
      configs: configs.map(c => ({
        key: c.key,
        value: c.value,
        valueType: c.valueType
      }))
    });
  }

  const { addConnection, removeConnection } = require('./polling');
  const connectionKey = addConnection(appId, environment, clientVersion, res);

  req.on('close', () => {
    removeConnection(connectionKey);
  });
});

router.post('/apps/:appId/configs/:key/rollback/:version', (req: Request, res: Response) => {
  const appId = parseInt(req.params.appId);
  if (isNaN(appId)) {
    return res.status(400).json({ error: 'Invalid app ID' });
  }

  const version = parseInt(req.params.version);
  if (isNaN(version)) {
    return res.status(400).json({ error: 'Invalid version' });
  }

  const app = getAppById(appId);
  if (!app) {
    return res.status(404).json({ error: 'App not found' });
  }

  const environment = req.query.environment as Environment | undefined;
  if (!environment) {
    return res.status(400).json({ error: 'environment query parameter is required' });
  }

  const envValidation = validateEnvironment(environment);
  if (!envValidation.valid) {
    return res.status(400).json({ error: envValidation.error });
  }

  const config = getConfig(appId, environment, req.params.key);
  if (!config) {
    return res.status(404).json({ error: 'Config not found' });
  }

  const operator = req.body.operator as string | undefined;
  if (!operator) {
    return res.status(400).json({ error: 'operator is required in request body' });
  }

  const history = getConfigHistoryByVersion(config.id, version);
  if (!history) {
    return res.status(404).json({ error: `Version ${version} not found in history` });
  }

  dbRollbackConfig(config.id, version, operator);

  const updatedConfig = getConfigById(config.id);
  if (updatedConfig) {
    notifyConfigChange(appId, environment, updatedConfig.version);
  }

  res.json(updatedConfig);
});

export default router;
