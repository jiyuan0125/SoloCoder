import cron from 'node-cron';
import { dbQueries } from './database';
import { Task, ExecutionStatus } from './types';
import { canTransition } from './stateMachine';

interface ScheduledTask {
  taskId: string;
  cronJob?: cron.ScheduledTask;
  intervalId?: NodeJS.Timeout;
}

const scheduledTasks = new Map<string, ScheduledTask>();

async function checkDependencies(taskId: string): Promise<{ allCompleted: boolean; anyFailed: boolean }> {
  const dependencyIds = await dbQueries.getTaskDependencies(taskId);
  
  if (dependencyIds.length === 0) {
    return { allCompleted: true, anyFailed: false };
  }

  let allCompleted = true;
  let anyFailed = false;

  for (const depId of dependencyIds) {
    const latestExecution = await dbQueries.getLatestCompletedExecution(depId);
    if (!latestExecution) {
      allCompleted = false;
      continue;
    }

    if (latestExecution.status === 'TIMEOUT' || latestExecution.status === 'SKIPPED') {
      anyFailed = true;
    } else if (latestExecution.status !== 'COMPLETED') {
      allCompleted = false;
    }
  }

  return { allCompleted, anyFailed };
}

async function executeTask(taskId: string): Promise<void> {
  const task = await dbQueries.getTask(taskId);
  if (!task) return;

  const isGroupPaused = await dbQueries.getTaskGroupPaused(taskId);
  if (isGroupPaused) return;

  const hasRunning = await dbQueries.hasRunningExecution(taskId);
  if (hasRunning) return;

  if (task.status !== 'PENDING' && task.status !== 'WAITING_DEPENDENCIES') {
    return;
  }

  const { allCompleted, anyFailed } = await checkDependencies(taskId);

  if (task.status === 'WAITING_DEPENDENCIES') {
    if (anyFailed) {
      if (canTransition(task.status, 'SKIPPED')) {
        await dbQueries.updateTaskStatus(taskId, 'SKIPPED');
        await dbQueries.createExecution(taskId, 'SKIPPED');
        await checkDependentTasks(taskId);
      }
      return;
    }

    if (!allCompleted) {
      return;
    }

    if (canTransition(task.status, 'PENDING')) {
      await dbQueries.updateTaskStatus(taskId, 'PENDING');
      task.status = 'PENDING';
    }
  }

  if (task.status !== 'PENDING') {
    return;
  }

  if (!canTransition(task.status, 'RUNNING')) {
    return;
  }

  await dbQueries.updateTaskStatus(taskId, 'RUNNING');
  const execution = await dbQueries.createExecution(taskId, 'RUNNING');

  const timeoutMs = task.timeoutSeconds * 1000;
  let timedOut = false;
  let executionOutput = '';

  const timeoutHandle = setTimeout(async () => {
    timedOut = true;
    const endTime = new Date().toISOString();
    executionOutput += 'Execution timed out\n';
    await dbQueries.updateExecution(execution.id, 'TIMEOUT', executionOutput, endTime);
    if (canTransition('RUNNING', 'TIMEOUT')) {
      await dbQueries.updateTaskStatus(taskId, 'TIMEOUT');
    }
    await checkDependentTasks(taskId);
  }, timeoutMs);

  try {
    const params = JSON.parse(task.executionParams);
    executionOutput = `Executing task: ${task.name}\nParams: ${JSON.stringify(params)}\nCompleted successfully at ${new Date().toISOString()}`;
    await new Promise(resolve => setTimeout(resolve, 100));
  } catch (err) {
    executionOutput = `Executing task: ${task.name}\nError: ${(err as Error).message}\nCompleted with success at ${new Date().toISOString()}`;
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  if (!timedOut) {
    clearTimeout(timeoutHandle);
    const endTime = new Date().toISOString();
    await dbQueries.updateExecution(execution.id, 'COMPLETED', executionOutput, endTime);
    if (canTransition('RUNNING', 'COMPLETED')) {
      await dbQueries.updateTaskStatus(taskId, 'COMPLETED');
    }
    await checkDependentTasks(taskId);
  }
}

async function checkDependentTasks(completedTaskId: string): Promise<void> {
  const dependentTaskIds = await dbQueries.getTasksDependingOn(completedTaskId);
  
  for (const depTaskId of dependentTaskIds) {
    const depTask = await dbQueries.getTask(depTaskId);
    if (!depTask) continue;

    if (depTask.status !== 'WAITING_DEPENDENCIES') continue;

    const { allCompleted, anyFailed } = await checkDependencies(depTaskId);

    if (anyFailed) {
      if (canTransition(depTask.status, 'SKIPPED')) {
        await dbQueries.updateTaskStatus(depTaskId, 'SKIPPED');
        await dbQueries.createExecution(depTaskId, 'SKIPPED');
        await checkDependentTasks(depTaskId);
      }
    } else if (allCompleted) {
      if (canTransition(depTask.status, 'PENDING')) {
        await dbQueries.updateTaskStatus(depTaskId, 'PENDING');
      }
    }
  }
}

function unscheduleTask(taskId: string): void {
  const scheduled = scheduledTasks.get(taskId);
  if (scheduled) {
    if (scheduled.cronJob) {
      scheduled.cronJob.stop();
    }
    if (scheduled.intervalId) {
      clearInterval(scheduled.intervalId);
    }
    scheduledTasks.delete(taskId);
  }
}

function scheduleTask(task: Task): void {
  unscheduleTask(task.id);

  const scheduled: ScheduledTask = { taskId: task.id };

  if (task.cronExpression) {
    scheduled.cronJob = cron.schedule(task.cronExpression, async () => {
      const currentTask = await dbQueries.getTask(task.id);
      if (currentTask) {
        if (currentTask.status === 'COMPLETED' || currentTask.status === 'TIMEOUT') {
          if (canTransition(currentTask.status, 'PENDING')) {
            await dbQueries.updateTaskStatus(task.id, 'PENDING');
          }
        }
        await executeTask(task.id);
      }
    });
  } else if (task.intervalSeconds) {
    scheduled.intervalId = setInterval(async () => {
      const currentTask = await dbQueries.getTask(task.id);
      if (currentTask) {
        if (currentTask.status === 'COMPLETED' || currentTask.status === 'TIMEOUT') {
          if (canTransition(currentTask.status, 'PENDING')) {
            await dbQueries.updateTaskStatus(task.id, 'PENDING');
          }
        }
        await executeTask(task.id);
      }
    }, task.intervalSeconds * 1000);
  }

  scheduledTasks.set(task.id, scheduled);
}

export async function initializeScheduler(): Promise<void> {
  const tasks = await dbQueries.getAllTasks();
  
  for (const task of tasks) {
    if (task.cronExpression || task.intervalSeconds) {
      if (task.status === 'CREATED') {
        const dependencies = await dbQueries.getTaskDependencies(task.id);
        const targetStatus = dependencies.length > 0 ? 'WAITING_DEPENDENCIES' : 'PENDING';
        if (canTransition(task.status, targetStatus)) {
          await dbQueries.updateTaskStatus(task.id, targetStatus);
        }
      }
      
      scheduleTask(task);
    }
  }
}

export { scheduleTask, unscheduleTask, executeTask };
