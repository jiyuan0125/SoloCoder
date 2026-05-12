import Database, { Database as DatabaseType } from 'better-sqlite3';

export const db: DatabaseType = new Database('./contracts.db');

export function initDatabase(): void {
  db.pragma('journal_mode = WAL');

  db.exec(`
    CREATE TABLE IF NOT EXISTS contract_templates (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      type TEXT NOT NULL CHECK(type IN ('采购', '销售', '服务协议', '保密协议')),
      version TEXT NOT NULL,
      content TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(name, version)
    );

    CREATE TABLE IF NOT EXISTS contracts (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      template_id INTEGER NOT NULL,
      template_name TEXT NOT NULL,
      template_version TEXT NOT NULL,
      template_type TEXT NOT NULL,
      template_content TEXT NOT NULL,
      variables TEXT NOT NULL,
      content TEXT NOT NULL,
      amount REAL NOT NULL,
      status TEXT NOT NULL DEFAULT '草稿' CHECK(status IN ('草稿', '审批中', '已批准', '签署中', '已签署', '已归档', '已作废')),
      is_supplement INTEGER DEFAULT 0,
      parent_contract_id INTEGER,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      approved_amount REAL,
      needs_manual_processing INTEGER DEFAULT 0,
      FOREIGN KEY (parent_contract_id) REFERENCES contracts(id),
      FOREIGN KEY (template_id) REFERENCES contract_templates(id)
    );

    CREATE TABLE IF NOT EXISTS approval_records (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      contract_id INTEGER NOT NULL,
      approver_role TEXT NOT NULL,
      approver_name TEXT,
      approved INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (contract_id) REFERENCES contracts(id)
    );

    CREATE TABLE IF NOT EXISTS signing_tasks (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      contract_id INTEGER NOT NULL,
      contract_amount REAL NOT NULL,
      status TEXT NOT NULL DEFAULT '待签署' CHECK(status IN ('待签署', '签署中', '已签署')),
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (contract_id) REFERENCES contracts(id)
    );

    CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
    CREATE INDEX IF NOT EXISTS idx_contracts_template ON contracts(template_id);
    CREATE INDEX IF NOT EXISTS idx_signing_tasks_contract ON signing_tasks(contract_id);
  `);
}

export type ContractTemplate = {
  id: number;
  name: string;
  type: '采购' | '销售' | '服务协议' | '保密协议';
  version: string;
  content: string;
  created_at: string;
};

export type ContractStatus = '草稿' | '审批中' | '已批准' | '签署中' | '已签署' | '已归档' | '已作废';

export type Contract = {
  id: number;
  template_id: number;
  template_name: string;
  template_version: string;
  template_type: string;
  template_content: string;
  variables: string;
  content: string;
  amount: number;
  status: ContractStatus;
  is_supplement: number;
  parent_contract_id: number | null;
  created_at: string;
  approved_amount: number | null;
  needs_manual_processing: number;
};

export type SigningTask = {
  id: number;
  contract_id: number;
  contract_amount: number;
  status: '待签署' | '签署中' | '已签署';
  created_at: string;
};
