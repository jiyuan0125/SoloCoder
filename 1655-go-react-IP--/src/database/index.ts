import Database from 'better-sqlite3';

const dbInstance = new Database('./security-gateway.db');

dbInstance.pragma('journal_mode = WAL');
dbInstance.pragma('synchronous = NORMAL');

dbInstance.exec(`
  CREATE TABLE IF NOT EXISTS rules (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('whitelist', 'blacklist')),
    ip_pattern TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER,
    enabled INTEGER NOT NULL DEFAULT 1,
    UNIQUE(type, ip_pattern)
  );

  CREATE INDEX IF NOT EXISTS idx_rules_type ON rules(type);
  CREATE INDEX IF NOT EXISTS idx_rules_enabled ON rules(enabled);

  CREATE TABLE IF NOT EXISTS rule_conditions (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    operator TEXT NOT NULL CHECK(operator IN ('equals', 'contains', 'regex')),
    FOREIGN KEY (rule_id) REFERENCES rules(id) ON DELETE CASCADE
  );

  CREATE INDEX IF NOT EXISTS idx_conditions_rule ON rule_conditions(rule_id);

  CREATE TABLE IF NOT EXISTS request_logs (
    id TEXT PRIMARY KEY,
    ip TEXT NOT NULL,
    path TEXT NOT NULL,
    method TEXT NOT NULL,
    timestamp INTEGER NOT NULL
  );

  CREATE INDEX IF NOT EXISTS idx_logs_ip_time ON request_logs(ip, timestamp);
  CREATE INDEX IF NOT EXISTS idx_logs_time ON request_logs(timestamp);
`);

const db = dbInstance as any;
export default db;
