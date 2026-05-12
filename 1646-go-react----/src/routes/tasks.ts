import express, { Request, Response } from 'express';
import { dbQueries } from '../database';
import { CreateTaskRequest, SetTaskDependenciesRequest, GetExecutionsQuery, TaskStatus } from '../types';
import { validateCronExpression } from '../utils';
import { canTransition } from '../stateMachine';

let scheduler: typeof import('../scheduler').scheduleTask | null = null;

export function setScheduler(scheduleFn: typeof import('../scheduler').scheduleTask) {
  scheduler = scheduleFn;
}

const router = express.Router();

router.post('/', async (req: Request, res: Response) => {
  try {
    const body: CreateTaskRequest = req.body;
    
    if (!body.name || typeof body.name !== 'string') {
      return res.status(400).json({ error: 'Task name is required' });
    }

    if (typeof body.timeoutSeconds !== 'number' || body.timeoutSeconds <= 0) {
      return res.status(400).json({ error: 'Timeout seconds must be a positive number' });
    }

    if (!body.executionParams || typeof body.executionParams !== 'string') {
      return res.status(400).json({ error: 'Execution params are required' });
    }

    const hasCron = !!body.cronExpression;
    const hasInterval = typeof body.intervalSeconds === 'number' && body.intervalSeconds > 0;
    
    if (!hasCron && !hasInterval) {
      return res.status(400).json({ error: 'Either cronExpression or intervalSeconds must be provided' });
    }

    if (hasCron && hasInterval) {
      return res.status(400).json({ error: 'Cannot provide both cronExpression and intervalSeconds' });
    }

    if (hasCron && !validateCronExpression(body.cronExpression!)) {
      return res.status(400).json({ error: 'Invalid cron expression format' });
    }

    if (body.groupId) {
      const group = await dbQueries.getTaskGroup(body.groupId);
      if (!group) {
        return res.status(404).json({ error: 'Task group not found' });
      }
    }

    const task = await dbQueries.createTask({
      name: body.name,
      groupId: body.groupId || null,
      cronExpression: body.cronExpression || null,
      intervalSeconds: body.intervalSeconds || null,
      timeoutSeconds: body.timeoutSeconds,
      executionParams: body.executionParams,
      status: 'CREATED'
    });

    if (scheduler) {
      const dependencies = await dbQueries.getTaskDependencies(task.id);
      const targetStatus = dependencies.length > 0 ? 'WAITING_DEPENDENCIES' : 'PENDING';
      if (canTransition(task.status, targetStatus)) {
        await dbQueries.updateTaskStatus(task.id, targetStatus);
        task.status = targetStatus;
      }
      scheduler(task);
    }

    res.status(201).json(task);
  } catch (error) {
    console.error('Error creating task:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/:id/dependencies', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const body: SetTaskDependenciesRequest = req.body;

    if (!body.dependencyIds || !Array.isArray(body.dependencyIds)) {
      return res.status(400).json({ error: 'dependencyIds must be an array' });
    }

    const task = await dbQueries.getTask(id);
    if (!task) {
      return res.status(404).json({ error: 'Task not found' });
    }

    if (body.dependencyIds.includes(id)) {
      return res.status(400).json({ error: 'Task cannot depend on itself' });
    }

    for (const depId of body.dependencyIds) {
      const depTask = await dbQueries.getTask(depId);
      if (!depTask) {
        return res.status(404).json({ error: `Dependency task ${depId} not found` });
      }
    }

    await dbQueries.setTaskDependencies(id, body.dependencyIds);

    if (body.dependencyIds.length > 0) {
      if (canTransition(task.status, 'WAITING_DEPENDENCIES')) {
        await dbQueries.updateTaskStatus(id, 'WAITING_DEPENDENCIES');
      }
    }

    const dependencies = await dbQueries.getTaskDependencies(id);
    res.status(200).json({ taskId: id, dependencies });
  } catch (error) {
    console.error('Error setting task dependencies:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id/status', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const { status } = req.body;

    const validStatuses: TaskStatus[] = ['CREATED', 'WAITING_DEPENDENCIES', 'PENDING', 'RUNNING', 'COMPLETED', 'TIMEOUT', 'SKIPPED'];
    if (!status || !validStatuses.includes(status)) {
      return res.status(400).json({ error: 'Invalid status' });
    }

    const task = await dbQueries.getTask(id);
    if (!task) {
      return res.status(404).json({ error: 'Task not found' });
    }

    if (!canTransition(task.status, status)) {
      return res.status(400).json({ error: `Cannot transition from ${task.status} to ${status}` });
    }

    await dbQueries.updateTaskStatus(id, status);
    res.status(200).json({ taskId: id, status });
  } catch (error) {
    console.error('Error updating task status:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:id/executions', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const query: GetExecutionsQuery = req.query as any;

    const task = await dbQueries.getTask(id);
    if (!task) {
      return res.status(404).json({ error: 'Task not found' });
    }

    if (query.startTime) {
      const date = new Date(query.startTime);
      if (isNaN(date.getTime())) {
        return res.status(400).json({ error: 'Invalid startTime format' });
      }
    }

    if (query.endTime) {
      const date = new Date(query.endTime);
      if (isNaN(date.getTime())) {
        return res.status(400).json({ error: 'Invalid endTime format' });
      }
    }

    const executions = await dbQueries.getTaskExecutions(id, query.startTime, query.endTime);
    res.status(200).json(executions);
  } catch (error) {
    console.error('Error getting task executions:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
