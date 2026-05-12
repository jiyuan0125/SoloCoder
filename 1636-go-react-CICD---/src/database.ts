import Database from 'better-sqlite3';
import { v4 as uuidv4 } from 'uuid';
import {
  Pipeline,
  PipelineExecution,
  Stage,
  Task,
  Notification,
  PipelineStatus,
  StageStatus,
  TaskStatus,
  TriggerType,
  TaskType,
} from './types';

const db = new Database('./cicd.db');

db.exec(`
CREATE TABLE IF NOT EXISTS pipelines (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  definition TEXT NOT NULL,
  allowed_webhook_sources TEXT,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS pipeline_executions (
  id TEXT PRIMARY KEY,
  pipeline_id TEXT NOT NULL,
  status TEXT NOT NULL,
  trigger_type TEXT NOT NULL,
  trigger_time INTEGER NOT NULL,
  start_time INTEGER,
  end_time INTEGER,
  FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
);

CREATE TABLE IF NOT EXISTS stages (
  id TEXT PRIMARY KEY,
  execution_id TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  "order" INTEGER NOT NULL,
  FOREIGN KEY (execution_id) REFERENCES pipeline_executions(id)
);

CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  execution_id TEXT NOT NULL,
  stage_id TEXT NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  environment TEXT,
  status TEXT NOT NULL,
  start_time INTEGER,
  end_time INTEGER,
  logs TEXT NOT NULL,
  "order" INTEGER NOT NULL,
  FOREIGN KEY (execution_id) REFERENCES pipeline_executions(id),
  FOREIGN KEY (stage_id) REFERENCES stages(id)
);

CREATE TABLE IF NOT EXISTS notifications (
  id TEXT PRIMARY KEY,
  pipeline_id TEXT NOT NULL,
  execution_id TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  FOREIGN KEY (pipeline_id) REFERENCES pipelines(id),
  FOREIGN KEY (execution_id) REFERENCES pipeline_executions(id)
);

CREATE INDEX IF NOT EXISTS idx_executions_pipeline ON pipeline_executions(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_stages_execution ON stages(execution_id);
CREATE INDEX IF NOT EXISTS idx_tasks_stage ON tasks(stage_id);
CREATE INDEX IF NOT EXISTS idx_notifications_execution ON notifications(execution_id);
`);

export const pipelineRepo = {
  create(pipeline: Omit<Pipeline, 'id' | 'createdAt'>): Pipeline {
    const id = uuidv4();
    const createdAt = Date.now();
    const stmt = db.prepare(`
      INSERT INTO pipelines (id, name, definition, allowed_webhook_sources, created_at)
      VALUES (?, ?, ?, ?, ?)
    `);
    stmt.run(id, pipeline.name, pipeline.definition, pipeline.allowedWebhookSources, createdAt);
    return { id, name: pipeline.name, definition: pipeline.definition, allowedWebhookSources: pipeline.allowedWebhookSources, createdAt };
  },

  findById(id: string): Pipeline | undefined {
    const row = db.prepare('SELECT * FROM pipelines WHERE id = ?').get(id) as any;
    if (!row) return undefined;
    return {
      id: row.id,
      name: row.name,
      definition: row.definition,
      allowedWebhookSources: row.allowed_webhook_sources,
      createdAt: row.created_at,
    };
  },
};

export const executionRepo = {
  create(execution: Omit<PipelineExecution, 'id'>): PipelineExecution {
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO pipeline_executions (id, pipeline_id, status, trigger_type, trigger_time, start_time, end_time)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(id, execution.pipelineId, execution.status, execution.triggerType, execution.triggerTime, execution.startTime, execution.endTime);
    return { id, ...execution };
  },

  findLatest(pipelineId: string): PipelineExecution | undefined {
    const row = db.prepare('SELECT * FROM pipeline_executions WHERE pipeline_id = ? ORDER BY trigger_time DESC LIMIT 1').get(pipelineId) as any;
    if (!row) return undefined;
    return {
      id: row.id,
      pipelineId: row.pipeline_id,
      status: row.status as PipelineStatus,
      triggerType: row.trigger_type as TriggerType,
      triggerTime: row.trigger_time,
      startTime: row.start_time,
      endTime: row.end_time,
    };
  },

  findById(id: string): PipelineExecution | undefined {
    const row = db.prepare('SELECT * FROM pipeline_executions WHERE id = ?').get(id) as any;
    if (!row) return undefined;
    return {
      id: row.id,
      pipelineId: row.pipeline_id,
      status: row.status as PipelineStatus,
      triggerType: row.trigger_type as TriggerType,
      triggerTime: row.trigger_time,
      startTime: row.start_time,
      endTime: row.end_time,
    };
  },

  updateStatus(id: string, status: PipelineStatus, startTime?: number, endTime?: number) {
    const stmt = db.prepare(`
      UPDATE pipeline_executions 
      SET status = ?, start_time = COALESCE(?, start_time), end_time = COALESCE(?, end_time)
      WHERE id = ?
    `);
    stmt.run(status, startTime ?? null, endTime ?? null, id);
  },
};

export const stageRepo = {
  create(stage: Omit<Stage, 'id'>): Stage {
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO stages (id, execution_id, name, status, "order")
      VALUES (?, ?, ?, ?, ?)
    `);
    stmt.run(id, stage.executionId, stage.name, stage.status, stage.order);
    return { id, ...stage };
  },

  findAllByExecutionId(executionId: string): Stage[] {
    const rows = db.prepare('SELECT * FROM stages WHERE execution_id = ? ORDER BY "order"').all(executionId) as any[];
    return rows.map((row: any) => ({
      id: row.id,
      executionId: row.execution_id,
      name: row.name,
      status: row.status as StageStatus,
      order: row.order,
    }));
  },

  findById(id: string): Stage | undefined {
    const row = db.prepare('SELECT * FROM stages WHERE id = ?').get(id) as any;
    if (!row) return undefined;
    return {
      id: row.id,
      executionId: row.execution_id,
      name: row.name,
      status: row.status as StageStatus,
      order: row.order,
    };
  },

  updateStatus(id: string, status: StageStatus) {
    db.prepare('UPDATE stages SET status = ? WHERE id = ?').run(status, id);
  },
};

export const taskRepo = {
  create(task: Omit<Task, 'id'>): Task {
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO tasks (id, execution_id, stage_id, name, type, environment, status, start_time, end_time, logs, "order")
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(
      id,
      task.executionId,
      task.stageId,
      task.name,
      task.type,
      task.environment ?? null,
      task.status,
      task.startTime ?? null,
      task.endTime ?? null,
      task.logs,
      task.order
    );
    return { id, ...task };
  },

  findAllByStageId(stageId: string): Task[] {
    const rows = db.prepare('SELECT * FROM tasks WHERE stage_id = ? ORDER BY "order"').all(stageId) as any[];
    return rows.map((row: any) => ({
      id: row.id,
      executionId: row.execution_id,
      stageId: row.stage_id,
      name: row.name,
      type: row.type as TaskType,
      environment: row.environment,
      status: row.status as TaskStatus | 'pending' | 'running',
      startTime: row.start_time,
      endTime: row.end_time,
      logs: row.logs,
      order: row.order,
    }));
  },

  findById(id: string): Task | undefined {
    const row = db.prepare('SELECT * FROM tasks WHERE id = ?').get(id) as any;
    if (!row) return undefined;
    return {
      id: row.id,
      executionId: row.execution_id,
      stageId: row.stage_id,
      name: row.name,
      type: row.type as TaskType,
      environment: row.environment,
      status: row.status as TaskStatus | 'pending' | 'running',
      startTime: row.start_time,
      endTime: row.end_time,
      logs: row.logs,
      order: row.order,
    };
  },

  update(task: Pick<Task, 'id' | 'status' | 'startTime' | 'endTime' | 'logs'>) {
    const stmt = db.prepare(`
      UPDATE tasks 
      SET status = ?, start_time = COALESCE(?, start_time), end_time = COALESCE(?, end_time), logs = ?
      WHERE id = ?
    `);
    stmt.run(task.status, task.startTime ?? null, task.endTime ?? null, task.logs, task.id);
  },

  findRunningDeploysByEnvironment(environment: string): Task[] {
    const rows = db.prepare("SELECT * FROM tasks WHERE type = 'deploy' AND status = 'running' AND environment = ?").all(environment) as any[];
    return rows.map((row: any) => ({
      id: row.id,
      executionId: row.execution_id,
      stageId: row.stage_id,
      name: row.name,
      type: row.type as TaskType,
      environment: row.environment,
      status: row.status as TaskStatus | 'pending' | 'running',
      startTime: row.start_time,
      endTime: row.end_time,
      logs: row.logs,
      order: row.order,
    }));
  },
};

export const notificationRepo = {
  create(notification: Omit<Notification, 'id' | 'createdAt'>): Notification {
    const id = uuidv4();
    const createdAt = Date.now();
    const stmt = db.prepare(`
      INSERT INTO notifications (id, pipeline_id, execution_id, content, created_at)
      VALUES (?, ?, ?, ?, ?)
    `);
    stmt.run(id, notification.pipelineId, notification.executionId, notification.content, createdAt);
    return { id, ...notification, createdAt };
  },
};
