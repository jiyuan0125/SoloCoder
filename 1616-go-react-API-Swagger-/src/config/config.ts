import { getDb } from '../database/db';
import { Config } from '../types';

const DEFAULT_CONFIG: Config = {
  proxy_timeout: 30000,
  max_request_body: 1048576,
};

let currentConfig: Config = { ...DEFAULT_CONFIG };

const loadConfig = (): void => {
  const db = getDb();
  const rows = db.prepare('SELECT key, value FROM config').all() as { key: string; value: string }[];
  
  for (const row of rows) {
    try {
      const value = JSON.parse(row.value);
      if (row.key in currentConfig) {
        (currentConfig as any)[row.key] = value;
      }
    } catch {
    }
  }
};

const getConfig = (): Config => {
  return { ...currentConfig };
};

const updateConfig = (newConfig: Partial<Config>): Config => {
  const db = getDb();
  const stmt = db.prepare('INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)');
  
  const updateStmt = db.transaction((config: Partial<Config>) => {
    for (const [key, value] of Object.entries(config)) {
      if (key in DEFAULT_CONFIG) {
        stmt.run(key, JSON.stringify(value));
        (currentConfig as any)[key] = value;
      }
    }
  });
  
  updateStmt(newConfig);
  return getConfig();
};

const initializeConfig = (): void => {
  const db = getDb();
  const stmt = db.prepare('INSERT OR IGNORE INTO config (key, value) VALUES (?, ?)');
  
  for (const [key, value] of Object.entries(DEFAULT_CONFIG)) {
    stmt.run(key, JSON.stringify(value));
  }
  
  loadConfig();
};

export { getConfig, updateConfig, initializeConfig, DEFAULT_CONFIG };
