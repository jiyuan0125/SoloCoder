import sqlite3 from 'sqlite3';
import { v4 as uuidv4 } from 'uuid';
import { Task, TaskGroup, TaskExecution, TaskStatus, ExecutionStatus } from './types';

const db = new sqlite3.Database('./tasks.db');

db.serialize(() => {
  db.run(`
    CREATE TABLE IF NOT EXISTS task_groups (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      paused INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS tasks (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      group_id TEXT,
      cron_expression TEXT,
      interval_seconds INTEGER,
      timeout_seconds INTEGER NOT NULL,
      execution_params TEXT NOT NULL,
      status TEXT NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (group_id) REFERENCES task_groups(id)
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS task_dependencies (
      task_id TEXT NOT NULL,
      depends_on_task_id TEXT NOT NULL,
      PRIMARY KEY (task_id, depends_on_task_id),
      FOREIGN KEY (task_id) REFERENCES tasks(id),
      FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id)
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS task_executions (
      id TEXT PRIMARY KEY,
      task_id TEXT NOT NULL,
      start_time TEXT NOT NULL,
      end_time TEXT,
      status TEXT NOT NULL,
      output TEXT NOT NULL,
      FOREIGN KEY (task_id) REFERENCES tasks(id)
    )
  `);
});

export const dbQueries = {
  // Task Group operations
  createTaskGroup: (name: string): Promise<TaskGroup> => {
    return new Promise((resolve, reject) => {
      const id = uuidv4();
      const createdAt = new Date().toISOString();
      db.run(
        'INSERT INTO task_groups (id, name, paused, created_at) VALUES (?, ?, 0, ?)',
        [id, name, createdAt],
        (err) => {
          if (err) reject(err);
          else resolve({ id, name, paused: false, createdAt: new Date(createdAt) });
        }
      );
    });
  },

  getTaskGroup: (id: string): Promise<TaskGroup | null> => {
    return new Promise((resolve, reject) => {
      db.get(
        'SELECT * FROM task_groups WHERE id = ?',
        [id],
        (err, row: any) => {
          if (err) reject(err);
          else if (row) resolve({
            id: row.id,
            name: row.name,
            paused: row.paused === 1,
            createdAt: new Date(row.created_at)
          });
          else resolve(null);
        }
      );
    });
  },

  updateTaskGroupPaused: (id: string, paused: boolean): Promise<void> => {
    return new Promise((resolve, reject) => {
      db.run(
        'UPDATE task_groups SET paused = ? WHERE id = ?',
        [paused ? 1 : 0, id],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  },

  // Task operations
  createTask: (task: Omit<Task, 'id' | 'createdAt' | 'updatedAt'>): Promise<Task> => {
    return new Promise((resolve, reject) => {
      const id = uuidv4();
      const now = new Date().toISOString();
      db.run(
        `INSERT INTO tasks (id, name, group_id, cron_expression, interval_seconds, timeout_seconds, execution_params, status, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          id,
          task.name,
          task.groupId,
          task.cronExpression,
          task.intervalSeconds,
          task.timeoutSeconds,
          task.executionParams,
          task.status,
          now,
          now
        ],
        (err) => {
          if (err) reject(err);
          else resolve({
            ...task,
            id,
            createdAt: new Date(now),
            updatedAt: new Date(now)
          });
        }
      );
    });
  },

  getTask: (id: string): Promise<Task | null> => {
    return new Promise((resolve, reject) => {
      db.get(
        'SELECT * FROM tasks WHERE id = ?',
        [id],
        (err, row: any) => {
          if (err) reject(err);
          else if (row) resolve({
            id: row.id,
            name: row.name,
            groupId: row.group_id,
            cronExpression: row.cron_expression,
            intervalSeconds: row.interval_seconds,
            timeoutSeconds: row.timeout_seconds,
            executionParams: row.execution_params,
            status: row.status as TaskStatus,
            createdAt: new Date(row.created_at),
            updatedAt: new Date(row.updated_at)
          });
          else resolve(null);
        }
      );
    });
  },

  updateTaskStatus: (id: string, status: TaskStatus): Promise<void> => {
    return new Promise((resolve, reject) => {
      const now = new Date().toISOString();
      db.run(
        'UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?',
        [status, now, id],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  },

  getAllTasks: (): Promise<Task[]> => {
    return new Promise((resolve, reject) => {
      db.all(
        'SELECT * FROM tasks',
        [],
        (err, rows: any[]) => {
          if (err) reject(err);
          else resolve(rows.map(row => ({
            id: row.id,
            name: row.name,
            groupId: row.group_id,
            cronExpression: row.cron_expression,
            intervalSeconds: row.interval_seconds,
            timeoutSeconds: row.timeout_seconds,
            executionParams: row.execution_params,
            status: row.status as TaskStatus,
            createdAt: new Date(row.created_at),
            updatedAt: new Date(row.updated_at)
          })));
        }
      );
    });
  },

  // Task Dependency operations
  setTaskDependencies: (taskId: string, dependencyIds: string[]): Promise<void> => {
    return new Promise((resolve, reject) => {
      db.run('DELETE FROM task_dependencies WHERE task_id = ?', [taskId], (err) => {
        if (err) reject(err);
        else {
          const promises = dependencyIds.map(depId => 
            new Promise<void>((resolveDep, rejectDep) => {
              db.run(
                'INSERT INTO task_dependencies (task_id, depends_on_task_id) VALUES (?, ?)',
                [taskId, depId],
                (err) => {
                  if (err) rejectDep(err);
                  else resolveDep();
                }
              );
            })
          );
          Promise.all(promises).then(() => resolve()).catch(reject);
        }
      });
    });
  },

  getTaskDependencies: (taskId: string): Promise<string[]> => {
    return new Promise((resolve, reject) => {
      db.all(
        'SELECT depends_on_task_id FROM task_dependencies WHERE task_id = ?',
        [taskId],
        (err, rows: any[]) => {
          if (err) reject(err);
          else resolve(rows.map(row => row.depends_on_task_id));
        }
      );
    });
  },

  getTasksDependingOn: (taskId: string): Promise<string[]> => {
    return new Promise((resolve, reject) => {
      db.all(
        'SELECT task_id FROM task_dependencies WHERE depends_on_task_id = ?',
        [taskId],
        (err, rows: any[]) => {
          if (err) reject(err);
          else resolve(rows.map(row => row.task_id));
        }
      );
    });
  },

  // Task Execution operations
  createExecution: (taskId: string, status: ExecutionStatus = 'RUNNING'): Promise<TaskExecution> => {
    return new Promise((resolve, reject) => {
      const id = uuidv4();
      const startTime = new Date().toISOString();
      db.run(
        'INSERT INTO task_executions (id, task_id, start_time, status, output) VALUES (?, ?, ?, ?, ?)',
        [id, taskId, startTime, status, ''],
        (err) => {
          if (err) reject(err);
          else resolve({
            id,
            taskId,
            startTime: new Date(startTime),
            endTime: null,
            status,
            output: ''
          });
        }
      );
    });
  },

  updateExecution: (id: string, status: ExecutionStatus, output: string, endTime: string): Promise<void> => {
    return new Promise((resolve, reject) => {
      db.run(
        'UPDATE task_executions SET end_time = ?, status = ?, output = ? WHERE id = ?',
        [endTime, status, output, id],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  },

  getTaskExecutions: (taskId: string, startTime?: string, endTime?: string): Promise<TaskExecution[]> => {
    return new Promise((resolve, reject) => {
      let query = 'SELECT * FROM task_executions WHERE task_id = ?';
      const params: any[] = [taskId];
      
      if (startTime) {
        query += ' AND start_time >= ?';
        params.push(startTime);
      }
      if (endTime) {
        query += ' AND start_time <= ?';
        params.push(endTime);
      }
      query += ' ORDER BY start_time DESC';
      
      db.all(query, params, (err, rows: any[]) => {
        if (err) reject(err);
        else resolve(rows.map(row => ({
          id: row.id,
          taskId: row.task_id,
          startTime: new Date(row.start_time),
          endTime: row.end_time ? new Date(row.end_time) : null,
          status: row.status as ExecutionStatus,
          output: row.output
        })));
      });
    });
  },

  getLatestCompletedExecution: (taskId: string): Promise<TaskExecution | null> => {
    return new Promise((resolve, reject) => {
      db.get(
        `SELECT * FROM task_executions 
         WHERE task_id = ? AND status IN ('COMPLETED', 'TIMEOUT', 'SKIPPED')
         ORDER BY end_time DESC
         LIMIT 1`,
        [taskId],
        (err, row: any) => {
          if (err) reject(err);
          else if (row) resolve({
            id: row.id,
            taskId: row.task_id,
            startTime: new Date(row.start_time),
            endTime: row.end_time ? new Date(row.end_time) : null,
            status: row.status as ExecutionStatus,
            output: row.output
          });
          else resolve(null);
        }
      );
    });
  },

  hasRunningExecution: (taskId: string): Promise<boolean> => {
    return new Promise((resolve, reject) => {
      db.get(
        'SELECT COUNT(*) as count FROM task_executions WHERE task_id = ? AND status = ?',
        [taskId, 'RUNNING'],
        (err, row: any) => {
          if (err) reject(err);
          else resolve(row.count > 0);
        }
      );
    });
  },

  getTaskGroupPaused: (taskId: string): Promise<boolean> => {
    return new Promise((resolve, reject) => {
      db.get(
        `SELECT tg.paused 
         FROM tasks t
         LEFT JOIN task_groups tg ON t.group_id = tg.id
         WHERE t.id = ?`,
        [taskId],
        (err, row: any) => {
          if (err) reject(err);
          else resolve(row && row.paused === 1);
        }
      );
    });
  }
};

export default db;
