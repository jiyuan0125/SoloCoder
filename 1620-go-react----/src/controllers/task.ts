import { Request, Response } from 'express';
import {
  getAllTasks,
  getTaskById,
  createTask,
  updateTask,
  deleteTask,
  updateTaskStatus,
} from '../models/task';
import { getReportsByTaskId, getLatestReport } from '../models/report';
import { getManualInterventionItems } from '../models/retry';
import { executeSync } from '../services/sync';
import { TaskStatus, ApiResponse, CreateTaskRequest, UpdateTaskRequest } from '../types';
import { SyncMode, ConflictStrategy } from '../types';

function isValidSyncMode(mode: string): mode is SyncMode {
  return Object.values(SyncMode).includes(mode as SyncMode);
}

function isValidConflictStrategy(strategy: string): strategy is ConflictStrategy {
  return Object.values(ConflictStrategy).includes(strategy as ConflictStrategy);
}

function isValidTaskStatus(status: string): status is TaskStatus {
  return Object.values(TaskStatus).includes(status as TaskStatus);
}

export const taskController = {
  list: (_req: Request, res: Response): void => {
    const tasks = getAllTasks();
    const response: ApiResponse = {
      success: true,
      data: tasks,
    };
    res.json(response);
  },

  get: (req: Request, res: Response): void => {
    const { id } = req.params;
    const task = getTaskById(id);

    if (!task) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    res.json({
      success: true,
      data: task,
    });
  },

  create: (req: Request, res: Response): void => {
    const body = req.body as Partial<CreateTaskRequest>;

    if (!body.name || !body.mode || !body.sourceSystem || !body.targetSystem || !body.conflictStrategy) {
      res.status(400).json({
        success: false,
        error: 'Missing required fields: name, mode, sourceSystem, targetSystem, conflictStrategy',
      });
      return;
    }

    if (!isValidSyncMode(body.mode)) {
      res.status(400).json({
        success: false,
        error: `Invalid mode. Must be one of: ${Object.values(SyncMode).join(', ')}`,
      });
      return;
    }

    if (!isValidConflictStrategy(body.conflictStrategy)) {
      res.status(400).json({
        success: false,
        error: `Invalid conflict strategy. Must be one of: ${Object.values(ConflictStrategy).join(', ')}`,
      });
      return;
    }

    const task = createTask(body as CreateTaskRequest);

    res.status(201).json({
      success: true,
      data: task,
    });
  },

  update: (req: Request, res: Response): void => {
    const { id } = req.params;
    const body = req.body as UpdateTaskRequest;

    const existing = getTaskById(id);
    if (!existing) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    if (body.conflictStrategy !== undefined && !isValidConflictStrategy(body.conflictStrategy)) {
      res.status(400).json({
        success: false,
        error: `Invalid conflict strategy. Must be one of: ${Object.values(ConflictStrategy).join(', ')}`,
      });
      return;
    }

    const updated = updateTask(id, body);

    res.json({
      success: true,
      data: updated,
    });
  },

  remove: (req: Request, res: Response): void => {
    const { id } = req.params;
    const deleted = deleteTask(id);

    if (!deleted) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    res.json({
      success: true,
      data: null,
    });
  },

  start: async (req: Request, res: Response): Promise<void> => {
    const { id } = req.params;
    const task = getTaskById(id);

    if (!task) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    const result = await executeSync(task);

    if (!result.success) {
      if (result.error && result.error.includes('Invalid status transition')) {
        res.status(400).json({
          success: false,
          error: result.error,
        });
        return;
      }

      res.status(500).json({
        success: false,
        error: result.error,
      });
      return;
    }

    if (result.manualInterventionItems && result.manualInterventionItems.length > 0) {
      res.json({
        success: true,
        data: {
          report: result.report,
          manualInterventionItems: result.manualInterventionItems,
          message: `Task completed with items requiring manual intervention: ${result.manualInterventionItems.join(', ')}`,
        },
      });
      return;
    }

    res.json({
      success: true,
      data: {
        report: result.report,
        message: 'Sync completed successfully',
      },
    });
  },

  updateStatus: (req: Request, res: Response): void => {
    const { id } = req.params;
    const { status } = req.body;

    if (!status || !isValidTaskStatus(status)) {
      res.status(400).json({
        success: false,
        error: `Invalid status. Must be one of: ${Object.values(TaskStatus).join(', ')}`,
      });
      return;
    }

    const result = updateTaskStatus(id, status);

    if (!result.success) {
      if (result.reason === 'Task not found') {
        res.status(404).json({
          success: false,
          error: result.reason,
        });
      } else {
        res.status(400).json({
          success: false,
          error: result.reason,
        });
      }
      return;
    }

    res.json({
      success: true,
      data: result.task,
    });
  },

  listReports: (req: Request, res: Response): void => {
    const { id } = req.params;
    const task = getTaskById(id);

    if (!task) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    const reports = getReportsByTaskId(id);

    res.json({
      success: true,
      data: reports,
    });
  },

  getLatestReport: (req: Request, res: Response): void => {
    const { id } = req.params;
    const task = getTaskById(id);

    if (!task) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    const report = getLatestReport(id);

    res.json({
      success: true,
      data: report,
    });
  },

  getManualItems: (req: Request, res: Response): void => {
    const { id } = req.params;
    const task = getTaskById(id);

    if (!task) {
      res.status(404).json({
        success: false,
        error: 'Task not found',
      });
      return;
    }

    const items = getManualInterventionItems(id);

    res.json({
      success: true,
      data: items,
    });
  },
};
