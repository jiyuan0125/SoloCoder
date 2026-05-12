import Database from 'better-sqlite3';
import path from 'path';

const db = new Database(path.join(__dirname, '..', 'foundation.db'));

db.pragma('journal_mode = WAL');

function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS projects (
      id TEXT PRIMARY KEY,
      organization_name TEXT NOT NULL,
      credit_code TEXT NOT NULL,
      contact_person TEXT NOT NULL,
      project_name TEXT NOT NULL,
      category TEXT NOT NULL,
      budget_amount REAL NOT NULL,
      used_budget REAL NOT NULL DEFAULT 0,
      implementation_period INTEGER NOT NULL,
      status TEXT NOT NULL,
      submit_count INTEGER NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS experts (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      field TEXT NOT NULL,
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS reviews (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      expert_id TEXT NOT NULL,
      score INTEGER NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (project_id) REFERENCES projects(id),
      FOREIGN KEY (expert_id) REFERENCES experts(id)
    );

    CREATE TABLE IF NOT EXISTS audit_logs (
      id TEXT PRIMARY KEY,
      operator TEXT NOT NULL,
      operation_time TEXT NOT NULL,
      operation_type TEXT NOT NULL,
      content TEXT NOT NULL,
      before_snapshot TEXT,
      after_snapshot TEXT,
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS todos (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      title TEXT NOT NULL,
      description TEXT NOT NULL,
      deadline TEXT NOT NULL,
      status TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (project_id) REFERENCES projects(id)
    );

    CREATE TABLE IF NOT EXISTS budget_alerts (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      usage_rate REAL NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (project_id) REFERENCES projects(id)
    );

    CREATE TABLE IF NOT EXISTS fund_allocations (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      amount REAL NOT NULL,
      operator TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (project_id) REFERENCES projects(id)
    );

    CREATE TABLE IF NOT EXISTS credit_codes (
      code TEXT PRIMARY KEY,
      organization_name TEXT NOT NULL
    );
  `);

  const expertCount = db.prepare('SELECT COUNT(*) as count FROM experts').get() as { count: number };
  if (expertCount.count === 0) {
    const insertExpert = db.prepare(`
      INSERT INTO experts (id, name, field, created_at)
      VALUES (?, ?, ?, ?)
    `);

    const experts = [
      { id: 'exp1', name: '张明', field: '公益项目评估' },
      { id: 'exp2', name: '李华', field: '社会服务' },
      { id: 'exp3', name: '王芳', field: '教育公益' },
      { id: 'exp4', name: '赵强', field: '环境保护' },
      { id: 'exp5', name: '刘静', field: '医疗公益' },
      { id: 'exp6', name: '陈伟', field: '社区发展' }
    ];

    const now = new Date().toISOString();
    for (const expert of experts) {
      insertExpert.run(expert.id, expert.name, expert.field, now);
    }
  }
}

export { db, initDatabase };
