import { Task, Execution, TaskStatus, ExecutionStatus, ConcurrencyStrategy } from '../models/task';
import { taskStorageService } from './taskStorageService';
import { executeTask } from '../utils/taskExecutor';

const RETRY_INTERVALS = [60000, 300000, 1800000];

interface ExecutionState {
  taskId: string;
  executionId: string;
  abortController: AbortController;
  timeoutId: NodeJS.Timeout;
}

export class TaskExecutionEngine {
  private runningExecutions: Map<string, ExecutionState> = new Map();
  private waitingExecutions: Map<string, string[]> = new Map();
  private lastManualTrigger: Map<string, number> = new Map();
  private manualTriggerCooldown = 1000;

  async scheduleTaskExecution(taskId: string, isManual: boolean = false): Promise<Execution | null> {
    const task = taskStorageService.getTask(taskId);
    if (!task) {
      return null;
    }

    if (isManual) {
      const now = Date.now();
      const lastTrigger = this.lastManualTrigger.get(taskId) || 0;
      if (now - lastTrigger < this.manualTriggerCooldown) {
        console.log(`[ExecutionEngine] 任务 ${taskId} 手动触发过于频繁，跳过`);
        return null;
      }
      this.lastManualTrigger.set(taskId, now);
    }

    if (task.status === TaskStatus.COMPLETED) {
      console.log(`[ExecutionEngine] 任务 ${taskId} 已完成，无法触发`);
      return null;
    }

    if (!isManual && task.status !== TaskStatus.CREATED && task.status !== TaskStatus.PENDING) {
      console.log(`[ExecutionEngine] 任务 ${taskId} 状态为 ${task.status}，无法从待调度转执行中`);
      return null;
    }

    const runningExecution = taskStorageService.getRunningExecution(taskId);
    if (runningExecution) {
      if (isManual) {
        console.log(`[ExecutionEngine] 任务 ${taskId} 正在执行中，手动触发返回 409`);
        return null;
      }

      return await this.handleConcurrencyStrategy(task, runningExecution);
    }

    if (!isManual) {
      taskStorageService.updateTask(taskId, { status: TaskStatus.RUNNING });
    }

    return await this.executeTaskWithRetry(task, isManual);
  }

  private async handleConcurrencyStrategy(
    task: Task, 
    runningExecution: Execution
  ): Promise<Execution | null> {
    switch (task.concurrencyStrategy) {
      case ConcurrencyStrategy.SKIP:
        console.log(`[ExecutionEngine] 任务 ${task.id} 前一次未执行完，跳过本次`);
        return null;

      case ConcurrencyStrategy.WAIT:
        console.log(`[ExecutionEngine] 任务 ${task.id} 前一次未执行完，等待完成后执行`);
        if (!this.waitingExecutions.has(task.id)) {
          this.waitingExecutions.set(task.id, []);
        }
        this.waitingExecutions.get(task.id)!.push('queued');
        return null;

      case ConcurrencyStrategy.FORCE_TERMINATE:
        console.log(`[ExecutionEngine] 任务 ${task.id} 强制终止前一次执行`);
        this.terminateExecution(task.id, runningExecution.id);
        return await this.executeTaskWithRetry(task, false);

      default:
        return null;
    }
  }

  private async executeTaskWithRetry(task: Task, isManual: boolean): Promise<Execution> {
    const execution = taskStorageService.createExecution(task.id, isManual);
    console.log(`[ExecutionEngine] 开始执行任务 ${task.id}，执行 ID: ${execution.id}`);

    const abortController = new AbortController();
    const timeoutId = setTimeout(() => {
      this.handleTimeout(task.id, execution.id);
    }, task.timeout * 1000);

    this.runningExecutions.set(execution.id, {
      taskId: task.id,
      executionId: execution.id,
      abortController,
      timeoutId
    });

    await this.executeWithRetry(task, execution, 0);

    return execution;
  }

  private async executeWithRetry(
    task: Task, 
    execution: Execution, 
    attempt: number
  ): Promise<void> {
    try {
      console.log(`[ExecutionEngine] 执行任务 ${task.id} 第 ${attempt + 1} 次尝试`);
      
      const result = await executeTask(task.params);
      
      clearTimeout(this.runningExecutions.get(execution.id)?.timeoutId);
      
      const currentExecution = taskStorageService.getExecutions(task.id)
        .find(e => e.id === execution.id);
      
      if (currentExecution?.status === ExecutionStatus.RUNNING) {
        taskStorageService.updateExecution(task.id, execution.id, {
          status: ExecutionStatus.COMPLETED,
          endTime: new Date(),
          result
        });

        taskStorageService.updateTask(task.id, {
          status: TaskStatus.COMPLETED
        });

        console.log(`[ExecutionEngine] 任务 ${task.id} 执行成功`);
        this.processWaitingExecutions(task.id);
      } else {
        console.log(`[ExecutionEngine] 任务 ${task.id} 已超时或已终止，结果不覆盖状态`);
      }
    } catch (error: any) {
      console.error(`[ExecutionEngine] 任务 ${task.id} 执行失败: ${error.message}`);
      
      if (attempt < task.retryCount) {
        const retryIndex = Math.min(attempt, RETRY_INTERVALS.length - 1);
        const retryInterval = RETRY_INTERVALS[retryIndex];
        
        console.log(`[ExecutionEngine] 任务 ${task.id} 将在 ${retryInterval / 1000} 秒后重试`);
        
        setTimeout(() => {
          const currentExecution = taskStorageService.getExecutions(task.id)
            .find(e => e.id === execution.id);
          
          if (currentExecution?.status === ExecutionStatus.RUNNING) {
            taskStorageService.updateExecution(task.id, execution.id, {
              retryCount: attempt + 1
            });
            this.executeWithRetry(task, execution, attempt + 1);
          }
        }, retryInterval);
      } else {
        console.log(`[ExecutionEngine] 任务 ${task.id} 重试耗尽，标记为失败`);
        
        clearTimeout(this.runningExecutions.get(execution.id)?.timeoutId);
        
        taskStorageService.updateExecution(task.id, execution.id, {
          status: ExecutionStatus.FAILED,
          endTime: new Date(),
          error: error.message
        });

        taskStorageService.updateTask(task.id, {
          status: TaskStatus.FAILED
        });

        this.processWaitingExecutions(task.id);
      }
    } finally {
      this.runningExecutions.delete(execution.id);
    }
  }

  private handleTimeout(taskId: string, executionId: string): void {
    console.log(`[ExecutionEngine] 任务 ${taskId} 执行超时`);
    
    const executionState = this.runningExecutions.get(executionId);
    if (executionState) {
      executionState.abortController.abort();
    }
    
    const currentExecution = taskStorageService.getExecutions(taskId)
      .find(e => e.id === executionId);
    
    if (currentExecution?.status === ExecutionStatus.RUNNING) {
      taskStorageService.updateExecution(taskId, executionId, {
        status: ExecutionStatus.TIMEOUT,
        endTime: new Date(),
        error: '任务执行超时'
      });

      taskStorageService.updateTask(taskId, {
        status: TaskStatus.TIMEOUT
      });

      this.processWaitingExecutions(taskId);
    }
    
    this.runningExecutions.delete(executionId);
  }

  terminateExecution(taskId: string, executionId: string): void {
    const executionState = this.runningExecutions.get(executionId);
    if (executionState) {
      executionState.abortController.abort();
      clearTimeout(executionState.timeoutId);
      
      taskStorageService.updateExecution(taskId, executionId, {
        status: ExecutionStatus.FAILED,
        endTime: new Date(),
        error: '任务被强制终止'
      });

      taskStorageService.updateTask(taskId, {
        status: TaskStatus.FAILED
      });
      
      this.runningExecutions.delete(executionId);
    }
  }

  private async processWaitingExecutions(taskId: string): Promise<void> {
    const waitingQueue = this.waitingExecutions.get(taskId);
    if (waitingQueue && waitingQueue.length > 0) {
      waitingQueue.shift();
      this.waitingExecutions.set(taskId, waitingQueue);
      
      const task = taskStorageService.getTask(taskId);
      if (task) {
        console.log(`[ExecutionEngine] 处理等待队列中的任务 ${taskId}`);
        await this.executeTaskWithRetry(task, false);
      }
    }
  }

  isTaskRunning(taskId: string): boolean {
    return !!taskStorageService.getRunningExecution(taskId);
  }

  getRunningExecutionId(taskId: string): string | undefined {
    const running = taskStorageService.getRunningExecution(taskId);
    return running?.id;
  }
}

export const taskExecutionEngine = new TaskExecutionEngine();
