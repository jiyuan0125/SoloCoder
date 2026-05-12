import Database from 'better-sqlite3';
import path from 'path';
import { ConfigHistory, Environment } from './types';

export const db = new Database(path.join(__dirname, '..', 'config-center.db'));

db.exec(`
  CREATE TABLE IF NOT EXISTS apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    createdAt TEXT NOT NULL DEFAULT (datetime('now')),
    updatedAt TEXT NOT NULL DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    appId INTEGER NOT NULL,
    environment TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    valueType TEXT NOT NULL,
    description TEXT DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    createdAt TEXT NOT NULL DEFAULT (datetime('now')),
    updatedAt TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (appId) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE(appId, environment, key)
  );

  CREATE TABLE IF NOT EXISTS config_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    configId INTEGER NOT NULL,
    version INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    valueType TEXT NOT NULL,
    description TEXT DEFAULT '',
    operator TEXT NOT NULL,
    modifiedAt TEXT NOT NULL DEFAULT (datetime('now')),
    oldValue TEXT,
    newValue TEXT,
    operation TEXT NOT NULL,
    FOREIGN KEY (configId) REFERENCES configs(id) ON DELETE CASCADE
  );

  CREATE INDEX IF NOT EXISTS idx_configs_app_env ON configs(appId, environment);
  CREATE INDEX IF NOT EXISTS idx_history_config_id ON config_history(configId);
`);

export interface AppDb {
  id: number;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface ConfigDb {
  id: number;
  appId: number;
  environment: Environment;
  key: string;
  value: string;
  valueType: string;
  description: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface ConfigHistoryDb {
  id: number;
  configId: number;
  version: number;
  key: string;
  value: string;
  valueType: string;
  description: string;
  operator: string;
  modifiedAt: string;
  oldValue: string | null;
  newValue: string | null;
  operation: string;
}

export function getAppByName(name: string): AppDb | undefined {
  return db.prepare('SELECT * FROM apps WHERE name = ?').get(name) as AppDb | undefined;
}

export function getAppById(id: number): AppDb | undefined {
  return db.prepare('SELECT * FROM apps WHERE id = ?').get(id) as AppDb | undefined;
}

export function getAllApps(): AppDb[] {
  return db.prepare('SELECT * FROM apps ORDER BY id DESC').all() as AppDb[];
}

export function createApp(name: string, description: string = ''): number {
  const result = db.prepare('INSERT INTO apps (name, description) VALUES (?, ?)').run(name, description);
  return result.lastInsertRowid as number;
}

export function updateApp(id: number, name?: string, description?: string): void {
  const updates: string[] = [];
  const params: (string | number)[] = [];

  if (name !== undefined) {
    updates.push('name = ?');
    params.push(name);
  }
  if (description !== undefined) {
    updates.push('description = ?');
    params.push(description);
  }

  if (updates.length > 0) {
    updates.push('updatedAt = datetime("now")');
    params.push(id);
    db.prepare(`UPDATE apps SET ${updates.join(', ')} WHERE id = ?`).run(...params);
  }
}

export function deleteApp(id: number): void {
  db.prepare('DELETE FROM apps WHERE id = ?').run(id);
}

export function getConfig(appId: number, environment: Environment, key: string): ConfigDb | undefined {
  return db.prepare('SELECT * FROM configs WHERE appId = ? AND environment = ? AND key = ?').get(appId, environment, key) as ConfigDb | undefined;
}

export function getConfigById(id: number): ConfigDb | undefined {
  return db.prepare('SELECT * FROM configs WHERE id = ?').get(id) as ConfigDb | undefined;
}

export function getConfigsByAppAndEnv(appId: number, environment: Environment): ConfigDb[] {
  return db.prepare('SELECT * FROM configs WHERE appId = ? AND environment = ? ORDER BY id DESC').all(appId, environment) as ConfigDb[];
}

export function getConfigsByApp(appId: number): ConfigDb[] {
  return db.prepare('SELECT * FROM configs WHERE appId = ? ORDER BY id DESC').all(appId) as ConfigDb[];
}

export function createConfig(
  appId: number,
  environment: Environment,
  key: string,
  value: string,
  valueType: string,
  description: string,
  operator: string
): number {
  const insertConfig = db.prepare(`
    INSERT INTO configs (appId, environment, key, value, valueType, description, version)
    VALUES (?, ?, ?, ?, ?, ?, 1)
  `);

  const insertHistory = db.prepare(`
    INSERT INTO config_history (configId, version, key, value, valueType, description, operator, oldValue, newValue, operation)
    VALUES (?, 1, ?, ?, ?, ?, ?, NULL, ?, 'create')
  `);

  const transaction = db.transaction(() => {
    const result = insertConfig.run(appId, environment, key, value, valueType, description);
    const configId = result.lastInsertRowid as number;
    insertHistory.run(configId, key, value, valueType, description, operator, value);
    return configId;
  });

  return transaction();
}

export function updateConfig(
  id: number,
  newValue: string,
  newValueType: string,
  newDescription: string,
  operator: string
): void {
  const getCurrent = db.prepare('SELECT * FROM configs WHERE id = ?');
  const updateConfig = db.prepare(`
    UPDATE configs SET value = ?, valueType = ?, description = ?, version = version + 1, updatedAt = datetime('now')
    WHERE id = ?
  `);
  const insertHistory = db.prepare(`
    INSERT INTO config_history (configId, version, key, value, valueType, description, operator, oldValue, newValue, operation)
    VALUES (?, (SELECT version + 1 FROM configs WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, 'update')
  `);

  const transaction = db.transaction(() => {
    const current = getCurrent.get(id) as ConfigDb;
    updateConfig.run(newValue, newValueType, newDescription, id);
    insertHistory.run(id, id, current.key, newValue, newValueType, newDescription, operator, current.value, newValue);
  });

  transaction();
}

export function deleteConfig(id: number, operator: string): void {
  const getCurrent = db.prepare('SELECT * FROM configs WHERE id = ?');
  const insertHistory = db.prepare(`
    INSERT INTO config_history (configId, version, key, value, valueType, description, operator, oldValue, newValue, operation)
    VALUES (?, (SELECT version FROM configs WHERE id = ?), ?, ?, ?, ?, ?, ?, NULL, 'delete')
  `);
  const deleteConfigStmt = db.prepare('DELETE FROM configs WHERE id = ?');

  const transaction = db.transaction(() => {
    const current = getCurrent.get(id) as ConfigDb;
    if (current) {
      insertHistory.run(id, id, current.key, current.value, current.valueType, current.description, operator, current.value);
      deleteConfigStmt.run(id);
    }
  });

  transaction();
}

export function getConfigHistory(configId: number): ConfigHistoryDb[] {
  return db.prepare('SELECT * FROM config_history WHERE configId = ? ORDER BY version DESC').all(configId) as ConfigHistoryDb[];
}

export function getConfigHistoryByVersion(configId: number, version: number): ConfigHistoryDb | undefined {
  return db.prepare('SELECT * FROM config_history WHERE configId = ? AND version = ?').get(configId, version) as ConfigHistoryDb | undefined;
}

export function rollbackConfig(
  configId: number,
  targetVersion: number,
  operator: string
): void {
  const getCurrent = db.prepare('SELECT * FROM configs WHERE id = ?');
  const getHistory = db.prepare('SELECT * FROM config_history WHERE configId = ? AND version = ?');
  const updateConfig = db.prepare(`
    UPDATE configs SET value = ?, valueType = ?, description = ?, version = version + 1, updatedAt = datetime('now')
    WHERE id = ?
  `);
  const insertHistory = db.prepare(`
    INSERT INTO config_history (configId, version, key, value, valueType, description, operator, oldValue, newValue, operation)
    VALUES (?, (SELECT version + 1 FROM configs WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, 'rollback')
  `);

  const transaction = db.transaction(() => {
    const current = getCurrent.get(configId) as ConfigDb;
    const history = getHistory.get(configId, targetVersion) as ConfigHistoryDb;
    
    if (current && history) {
      updateConfig.run(history.value, history.valueType, history.description, configId);
      insertHistory.run(configId, configId, history.key, history.value, history.valueType, history.description, operator, current.value, history.value);
    }
  });

  transaction();
}
