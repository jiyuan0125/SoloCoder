import { Request, Response } from 'express';
import { 
  CreateTaskRequest, 
  UpdateTaskRequest, 
  TaskStatus,
  ConcurrencyStrategy
} from '../models/task';
import { taskStorageService } from '../services/taskStorageService';
import { taskSchedulerService } from '../services/taskSchedulerService';
import { taskExecutionEngine } from '../services/taskExecutionEngine';
import { validateCron } from '../utils/cronValidator';

export class TaskController {
  async createTask(req: Request, res: Response): Promise<void> {
    try {
      const request: CreateTaskRequest = req.body;

      if (!request.name || request.name.trim() === '') {
        res.status(400).json({ error: '任务名称不能为空' });
        return;
      }

      if (!request.cronExpression) {
        res.status(400).json({ error: 'Cron 表达式不能为空' });
        return;
      }

      const cronValidation = validateCron(request.cronExpression);
      if (!cronValidation.valid) {
        res.status(400).json({
          error: cronValidation.message,
          position: cronValidation.position
        });
        return;
      }

      if (request.timeout <= 0 || !Number.isInteger(request.timeout)) {
        res.status(400).json({ error: '超时时间必须是正整数' });
        return;
      }

      if (request.retryCount < 0 || !Number.isInteger(request.retryCount)) {
        res.status(400).json({ error: '重试次数必须是非负整数' });
        return;
      }

      if (!request.params || typeof request.params !== 'object') {
        res.status(400).json({ error: '执行参数必须是对象' });
        return;
      }

      const validStrategies = Object.values(ConcurrencyStrategy);
      if (!validStrategies.includes(request.concurrencyStrategy)) {
        res.status(400).json({ 
          error: `并发策略无效，有效值: ${validStrategies.join(', ')}` 
        });
        return;
      }

      const task = taskStorageService.createTask(request);
      taskSchedulerService.scheduleTask(task);

      res.status(201).json(task);
    } catch (error: any) {
      console.error('[TaskController] 创建任务失败:', error);
      res.status(500).json({ error: '创建任务失败: ' + error.message });
    }
  }

  async getAllTasks(req: Request, res: Response): Promise<void> {
    try {
      const tasks = taskStorageService.getAllTasks();
      res.status(200).json(tasks);
    } catch (error: any) {
      console.error('[TaskController] 获取任务列表失败:', error);
      res.status(500).json({ error: '获取任务列表失败: ' + error.message });
    }
  }

  async getTaskById(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const taskId = Array.isArray(id) ? id[0] : id;
      const task = taskStorageService.getTask(taskId);

      if (!task) {
        res.status(404).json({ error: '任务不存在' });
        return;
      }

      res.status(200).json(task);
    } catch (error: any) {
      console.error('[TaskController] 获取任务详情失败:', error);
      res.status(500).json({ error: '获取任务详情失败: ' + error.message });
    }
  }

  async updateTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const taskId = Array.isArray(id) ? id[0] : id;
      const request: UpdateTaskRequest = req.body;

      const task = taskStorageService.getTask(taskId);
      if (!task) {
        res.status(404).json({ error: '任务不存在' });
        return;
      }

      if (request.cronExpression) {
        const cronValidation = validateCron(request.cronExpression);
        if (!cronValidation.valid) {
          res.status(400).json({
            error: cronValidation.message,
            position: cronValidation.position
          });
          return;
        }
      }

      if (request.timeout !== undefined && (request.timeout <= 0 || !Number.isInteger(request.timeout))) {
        res.status(400).json({ error: '超时时间必须是正整数' });
        return;
      }

      if (request.retryCount !== undefined && (request.retryCount < 0 || !Number.isInteger(request.retryCount))) {
        res.status(400).json({ error: '重试次数必须是非负整数' });
        return;
      }

      if (request.concurrencyStrategy) {
        const validStrategies = Object.values(ConcurrencyStrategy);
        if (!validStrategies.includes(request.concurrencyStrategy)) {
          res.status(400).json({ 
            error: `并发策略无效，有效值: ${validStrategies.join(', ')}` 
          });
          return;
        }
      }

      const updatedTask = taskStorageService.updateTask(taskId, request);
      if (updatedTask) {
        if (request.cronExpression) {
          taskSchedulerService.updateTask(updatedTask);
        }
        res.status(200).json(updatedTask);
      } else {
        res.status(404).json({ error: '任务不存在' });
      }
    } catch (error: any) {
      console.error('[TaskController] 更新任务失败:', error);
      res.status(500).json({ error: '更新任务失败: ' + error.message });
    }
  }

  async deleteTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const taskId = Array.isArray(id) ? id[0] : id;
      
      const task = taskStorageService.getTask(taskId);
      if (!task) {
        res.status(404).json({ error: '任务不存在' });
        return;
      }

      taskSchedulerService.deleteTask(taskId);
      const deleted = taskStorageService.deleteTask(taskId);

      if (deleted) {
        res.status(204).send();
      } else {
        res.status(404).json({ error: '任务不存在' });
      }
    } catch (error: any) {
      console.error('[TaskController] 删除任务失败:', error);
      res.status(500).json({ error: '删除任务失败: ' + error.message });
    }
  }

  async getTaskExecutions(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const taskId = Array.isArray(id) ? id[0] : id;
      
      const task = taskStorageService.getTask(taskId);
      if (!task) {
        res.status(404).json({ error: '任务不存在' });
        return;
      }

      const executions = taskStorageService.getExecutions(taskId);
      res.status(200).json(executions);
    } catch (error: any) {
      console.error('[TaskController] 获取执行记录失败:', error);
      res.status(500).json({ error: '获取执行记录失败: ' + error.message });
    }
  }

  async triggerTask(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const taskId = Array.isArray(id) ? id[0] : id;
      
      const task = taskStorageService.getTask(taskId);
      if (!task) {
        res.status(404).json({ error: '任务不存在' });
        return;
      }

      if (task.status === TaskStatus.COMPLETED) {
        res.status(400).json({ 
          error: '任务已完成，无法再触发',
          currentStatus: task.status
        });
        return;
      }

      if (taskExecutionEngine.isTaskRunning(taskId)) {
        res.status(409).json({ 
          error: '任务正在执行中',
          currentStatus: task.status,
          runningExecutionId: taskExecutionEngine.getRunningExecutionId(taskId)
        });
        return;
      }

      const triggered = taskSchedulerService.triggerTaskManually(taskId);
      
      if (triggered) {
        res.status(202).json({ 
          message: '任务已触发',
          taskId: id
        });
      } else {
        res.status(400).json({ 
          error: '任务触发失败',
          currentStatus: task.status
        });
      }
    } catch (error: any) {
      console.error('[TaskController] 触发任务失败:', error);
      res.status(500).json({ error: '触发任务失败: ' + error.message });
    }
  }
}

export const taskController = new TaskController();
