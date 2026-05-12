import db from '../database';
import { BackupTask } from '../types';
import { generateId, validateCronExpression } from '../utils';

export class BackupTaskService {
  private static instance: BackupTaskService;
  
  static getInstance(): BackupTaskService {
    if (!this.instance) {
      this.instance = new BackupTaskService();
    }
    return this.instance;
  }

  private constructor() {}

  createTask(name: string, cronExpression: string, dataRange: string): { task: BackupTask | null; error?: string } {
    const validation = validateCronExpression(cronExpression);
    if (!validation.valid) {
      return { task: null, error: validation.reason || 'Invalid cron expression' };
    }

    const id = generateId();
    const task: BackupTask = {
      id,
      name,
      cronExpression,
      enabled: true,
      dataRange
    };

    try {
      const stmt = db.prepare(`
        INSERT INTO backup_tasks (id, name, cron_expression, enabled, data_range)
        VALUES (?, ?, ?, ?, ?)
      `);
      stmt.run(task.id, task.name, task.cronExpression, 1, task.dataRange);
      
      return { task };
    } catch (error) {
      return { task: null, error: error instanceof Error ? error.message : 'Failed to create task' };
    }
  }

  getTask(id: string): BackupTask | null {
    const row = db.prepare('SELECT * FROM backup_tasks WHERE id = ?').get(id) as any;
    if (!row) return null;
    
    return {
      id: row.id,
      name: row.name,
      cronExpression: row.cron_expression,
      enabled: row.enabled === 1,
      dataRange: row.data_range
    };
  }

  getAllTasks(): BackupTask[] {
    const rows = db.prepare('SELECT * FROM backup_tasks').all() as any[];
    return rows.map(row => ({
      id: row.id,
      name: row.name,
      cronExpression: row.cron_expression,
      enabled: row.enabled === 1,
      dataRange: row.data_range
    }));
  }

  toggleTask(id: string, enabled: boolean): { success: boolean; error?: string } {
    const task = this.getTask(id);
    if (!task) {
      return { success: false, error: 'Task not found' };
    }

    const stmt = db.prepare('UPDATE backup_tasks SET enabled = ? WHERE id = ?');
    stmt.run(enabled ? 1 : 0, id);
    
    return { success: true };
  }

  updateTaskCron(id: string, cronExpression: string): { success: boolean; error?: string } {
    const validation = validateCronExpression(cronExpression);
    if (!validation.valid) {
      return { success: false, error: validation.reason || 'Invalid cron expression' };
    }

    const task = this.getTask(id);
    if (!task) {
      return { success: false, error: 'Task not found' };
    }

    const stmt = db.prepare('UPDATE backup_tasks SET cron_expression = ? WHERE id = ?');
    stmt.run(cronExpression, id);
    
    return { success: true };
  }

  deleteTask(id: string): { success: boolean; error?: string } {
    const task = this.getTask(id);
    if (!task) {
      return { success: false, error: 'Task not found' };
    }

    const stmt = db.prepare('DELETE FROM backup_tasks WHERE id = ?');
    stmt.run(id);
    
    return { success: true };
  }
}
