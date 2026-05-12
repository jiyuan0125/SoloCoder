import Database from 'better-sqlite3';
import path from 'path';
import {
  WorkflowDefinition,
  WorkflowNode,
  NodeStatus,
  InstanceStatus,
  NodeInstance,
  WorkflowInstance,
  WorkflowVariable,
  HistoryRecord
} from './types';

const DB_PATH = path.join(__dirname, '..', 'workflow.db');

const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');

function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS workflows (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      start_node_id TEXT NOT NULL,
      nodes TEXT NOT NULL,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS instances (
      id TEXT PRIMARY KEY,
      workflow_id TEXT NOT NULL,
      status TEXT NOT NULL,
      current_node_ids TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (workflow_id) REFERENCES workflows(id)
    );

    CREATE TABLE IF NOT EXISTS node_instances (
      id TEXT PRIMARY KEY,
      instance_id TEXT NOT NULL,
      node_id TEXT NOT NULL,
      status TEXT NOT NULL,
      approver TEXT,
      completed_at INTEGER,
      parallel_branch_index INTEGER,
      FOREIGN KEY (instance_id) REFERENCES instances(id)
    );

    CREATE TABLE IF NOT EXISTS variables (
      instance_id TEXT NOT NULL,
      key TEXT NOT NULL,
      value TEXT NOT NULL,
      PRIMARY KEY (instance_id, key),
      FOREIGN KEY (instance_id) REFERENCES instances(id)
    );

    CREATE TABLE IF NOT EXISTS history (
      id TEXT PRIMARY KEY,
      instance_id TEXT NOT NULL,
      node_id TEXT NOT NULL,
      action TEXT NOT NULL,
      operator TEXT NOT NULL,
      timestamp INTEGER NOT NULL,
      details TEXT,
      FOREIGN KEY (instance_id) REFERENCES instances(id)
    );

    CREATE INDEX IF NOT EXISTS idx_node_instances_instance ON node_instances(instance_id);
    CREATE INDEX IF NOT EXISTS idx_variables_instance ON variables(instance_id);
    CREATE INDEX IF NOT EXISTS idx_history_instance ON history(instance_id);
  `);
}

initDatabase();

export const workflowRepo = {
  create(def: Omit<WorkflowDefinition, 'createdAt'>): WorkflowDefinition {
    const now = Date.now();
    const stmt = db.prepare(`
      INSERT INTO workflows (id, name, start_node_id, nodes, created_at)
      VALUES (?, ?, ?, ?, ?)
    `);
    stmt.run(def.id, def.name, def.startNodeId, JSON.stringify(def.nodes), now);
    return { ...def, createdAt: now };
  },

  getById(id: string): WorkflowDefinition | null {
    const row = db.prepare('SELECT * FROM workflows WHERE id = ?').get(id) as any;
    if (!row) return null;
    return {
      id: row.id,
      name: row.name,
      startNodeId: row.start_node_id,
      nodes: JSON.parse(row.nodes) as WorkflowNode[],
      createdAt: row.created_at
    };
  },

  getAll(): WorkflowDefinition[] {
    const rows = db.prepare('SELECT * FROM workflows ORDER BY created_at DESC').all() as any[];
    return rows.map((row: any) => ({
      id: row.id,
      name: row.name,
      startNodeId: row.start_node_id,
      nodes: JSON.parse(row.nodes) as WorkflowNode[],
      createdAt: row.created_at
    }));
  }
};

export const instanceRepo = {
  create(inst: Omit<WorkflowInstance, 'createdAt'>): WorkflowInstance {
    const now = Date.now();
    const stmt = db.prepare(`
      INSERT INTO instances (id, workflow_id, status, current_node_ids, created_at)
      VALUES (?, ?, ?, ?, ?)
    `);
    stmt.run(inst.id, inst.workflowId, inst.status, JSON.stringify(inst.currentNodeIds), now);
    return { ...inst, createdAt: now };
  },

  getById(id: string): WorkflowInstance | null {
    const row = db.prepare('SELECT * FROM instances WHERE id = ?').get(id) as any;
    if (!row) return null;
    return {
      id: row.id,
      workflowId: row.workflow_id,
      status: row.status as InstanceStatus,
      currentNodeIds: JSON.parse(row.current_node_ids),
      createdAt: row.created_at
    };
  },

  update(inst: WorkflowInstance): void {
    const stmt = db.prepare(`
      UPDATE instances
      SET status = ?, current_node_ids = ?
      WHERE id = ?
    `);
    stmt.run(inst.status, JSON.stringify(inst.currentNodeIds), inst.id);
  }
};

export const nodeInstanceRepo = {
  create(nodeInst: Omit<NodeInstance, 'id'> & { id?: string }): NodeInstance {
    const stmt = db.prepare(`
      INSERT INTO node_instances (id, instance_id, node_id, status, approver, completed_at, parallel_branch_index)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);
    const id = nodeInst.id || require('uuid').v4();
    stmt.run(
      id,
      nodeInst.instanceId,
      nodeInst.nodeId,
      nodeInst.status,
      nodeInst.approver || null,
      nodeInst.completedAt || null,
      nodeInst.parallelBranchIndex ?? null
    );
    return { ...nodeInst, id } as NodeInstance;
  },

  update(nodeInst: NodeInstance): void {
    const stmt = db.prepare(`
      UPDATE node_instances
      SET status = ?, approver = ?, completed_at = ?, parallel_branch_index = ?
      WHERE id = ?
    `);
    stmt.run(
      nodeInst.status,
      nodeInst.approver || null,
      nodeInst.completedAt || null,
      nodeInst.parallelBranchIndex ?? null,
      nodeInst.id
    );
  },

  getById(id: string): NodeInstance | null {
    const row = db.prepare('SELECT * FROM node_instances WHERE id = ?').get(id) as any;
    if (!row) return null;
    return {
      id: row.id,
      instanceId: row.instance_id,
      nodeId: row.node_id,
      status: row.status as NodeStatus,
      approver: row.approver,
      completedAt: row.completed_at,
      parallelBranchIndex: row.parallel_branch_index
    };
  },

  getByInstanceAndNode(instanceId: string, nodeId: string, parallelBranchIndex?: number): NodeInstance | null {
    let sql = 'SELECT * FROM node_instances WHERE instance_id = ? AND node_id = ?';
    const params: any[] = [instanceId, nodeId];
    if (parallelBranchIndex !== undefined) {
      sql += ' AND parallel_branch_index = ?';
      params.push(parallelBranchIndex);
    }
    const row = db.prepare(sql).get(...params) as any;
    if (!row) return null;
    return {
      id: row.id,
      instanceId: row.instance_id,
      nodeId: row.node_id,
      status: row.status as NodeStatus,
      approver: row.approver,
      completedAt: row.completed_at,
      parallelBranchIndex: row.parallel_branch_index
    };
  },

  getByInstance(instanceId: string): NodeInstance[] {
    const rows = db.prepare('SELECT * FROM node_instances WHERE instance_id = ?').all(instanceId) as any[];
    return rows.map((row: any) => ({
      id: row.id,
      instanceId: row.instance_id,
      nodeId: row.node_id,
      status: row.status as NodeStatus,
      approver: row.approver,
      completedAt: row.completed_at,
      parallelBranchIndex: row.parallel_branch_index
    }));
  },

  resetNodesAfter(instanceId: string, targetNodeId: string): void {
    const nodeInstances = this.getByInstance(instanceId);
    const target = nodeInstances.find(n => n.nodeId === targetNodeId);
    if (!target) return;

    const toReset = nodeInstances.filter(n =>
      n.completedAt && target.completedAt && n.completedAt > target.completedAt
    );

    const stmt = db.prepare(`
      UPDATE node_instances
      SET status = 'pending', approver = NULL, completed_at = NULL
      WHERE id = ?
    `);

    for (const node of toReset) {
      stmt.run(node.id);
    }
  }
};

export const variableRepo = {
  set(instanceId: string, key: string, value: string): void {
    const stmt = db.prepare(`
      INSERT OR REPLACE INTO variables (instance_id, key, value)
      VALUES (?, ?, ?)
    `);
    stmt.run(instanceId, key, value);
  },

  setMany(instanceId: string, variables: Record<string, string>): void {
    for (const [key, value] of Object.entries(variables)) {
      this.set(instanceId, key, value);
    }
  },

  get(instanceId: string, key: string): string | null {
    const row = db.prepare('SELECT value FROM variables WHERE instance_id = ? AND key = ?').get(instanceId, key) as any;
    return row ? row.value : null;
  },

  getAll(instanceId: string): Record<string, string> {
    const rows = db.prepare('SELECT key, value FROM variables WHERE instance_id = ?').all(instanceId) as any[];
    const result: Record<string, string> = {};
    for (const row of rows) {
      result[row.key] = row.value;
    }
    return result;
  }
};

export const historyRepo = {
  create(record: Omit<HistoryRecord, 'id' | 'timestamp'>): HistoryRecord {
    const uuid = require('uuid');
    const id = uuid.v4();
    const now = Date.now();
    const stmt = db.prepare(`
      INSERT INTO history (id, instance_id, node_id, action, operator, timestamp, details)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(id, record.instanceId, record.nodeId, record.action, record.operator, now, record.details || null);
    return { ...record, id, timestamp: now };
  },

  getByInstance(instanceId: string): HistoryRecord[] {
    const rows = db.prepare(`
      SELECT * FROM history
      WHERE instance_id = ?
      ORDER BY timestamp ASC
    `).all(instanceId) as any[];

    return rows.map((row: any) => ({
      id: row.id,
      instanceId: row.instance_id,
      nodeId: row.node_id,
      action: row.action,
      operator: row.operator,
      timestamp: row.timestamp,
      details: row.details
    }));
  }
};

export const dbInstance = db;
