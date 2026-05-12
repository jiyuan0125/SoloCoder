import { Request, Response } from 'express';
import { BackupTaskService } from '../services/BackupTaskService';
import { BackupService } from '../services/BackupService';
import { validateCronExpression } from '../utils';

export class BackupTaskController {
  static async createTask(req: Request, res: Response): Promise<void> {
    try {
      const { name, cronExpression, dataRange } = req.body;
      
      if (!name || !cronExpression || !dataRange) {
        res.status(400).json({ error: 'Missing required fields: name, cronExpression, dataRange' });
        return;
      }

      const validation = validateCronExpression(cronExpression);
      if (!validation.valid) {
        res.status(400).json({ 
          error: 'Invalid cron expression',
          reason: validation.reason 
        });
        return;
      }

      const result = BackupTaskService.getInstance().createTask(name, cronExpression, dataRange);
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      if (result.task) {
        const backupService = BackupService.getInstance();
        backupService.onTaskEnabled(result.task.id, result.task.cronExpression);
      }

      res.status(201).json(result.task);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async getTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const task = BackupTaskService.getInstance().getTask(id);
      
      if (!task) {
        res.status(404).json({ error: 'Task not found' });
        return;
      }

      res.json(task);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async getAllTasks(req: Request, res: Response): Promise<void> {
    try {
      const tasks = BackupTaskService.getInstance().getAllTasks();
      res.json(tasks);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async toggleTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const { enabled } = req.body;
      
      if (typeof enabled !== 'boolean') {
        res.status(400).json({ error: 'enabled must be a boolean' });
        return;
      }

      const task = BackupTaskService.getInstance().getTask(id);
      if (!task) {
        res.status(404).json({ error: 'Task not found' });
        return;
      }

      const result = BackupTaskService.getInstance().toggleTask(id, enabled);
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      const backupService = BackupService.getInstance();
      if (enabled) {
        backupService.onTaskEnabled(id, task.cronExpression);
      } else {
        backupService.onTaskDisabled(id);
      }

      const updatedTask = BackupTaskService.getInstance().getTask(id);
      res.json(updatedTask);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async updateCron(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const { cronExpression } = req.body;
      
      if (!cronExpression) {
        res.status(400).json({ error: 'cronExpression is required' });
        return;
      }

      const validation = validateCronExpression(cronExpression);
      if (!validation.valid) {
        res.status(400).json({ 
          error: 'Invalid cron expression',
          reason: validation.reason 
        });
        return;
      }

      const result = BackupTaskService.getInstance().updateTaskCron(id, cronExpression);
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      const backupService = BackupService.getInstance();
      backupService.onTaskCronUpdated(id, cronExpression);

      const updatedTask = BackupTaskService.getInstance().getTask(id);
      res.json(updatedTask);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async deleteTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      
      const backupService = BackupService.getInstance();
      backupService.onTaskDisabled(id);

      const result = BackupTaskService.getInstance().deleteTask(id);
      
      if (result.error) {
        res.status(404).json({ error: result.error });
        return;
      }

      res.status(204).send();
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async triggerManualBackup(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      
      const task = BackupTaskService.getInstance().getTask(id);
      if (!task) {
        res.status(404).json({ error: 'Task not found' });
        return;
      }

      if (!task.enabled) {
        res.status(400).json({ error: '任务已禁用，无法触发手动备份' });
        return;
      }

      const result = await BackupService.getInstance().triggerManualBackup(id);
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      res.status(201).json(result.snapshot);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }
}
