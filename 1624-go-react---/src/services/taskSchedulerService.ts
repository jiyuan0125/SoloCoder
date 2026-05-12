import { schedule, ScheduledTask } from 'node-cron';
import { Task, TaskStatus } from '../models/task';
import { taskStorageService } from './taskStorageService';
import { taskExecutionEngine } from './taskExecutionEngine';

export class TaskSchedulerService {
  private scheduledTasks: Map<string, ScheduledTask> = new Map();

  start(): void {
    console.log('[Scheduler] 启动任务调度器');
    
    const tasks = taskStorageService.getAllTasks();
    tasks.forEach(task => {
      this.scheduleTask(task);
    });
  }

  scheduleTask(task: Task): void {
    this.unscheduleTask(task.id);
    
    if (task.status === TaskStatus.COMPLETED) {
      console.log(`[Scheduler] 任务 ${task.id} 已完成，跳过调度`);
      return;
    }

    try {
      taskStorageService.updateTask(task.id, { status: TaskStatus.PENDING });
      
      const scheduledTask = schedule(task.cronExpression, async () => {
        console.log(`[Scheduler] 触发任务执行: ${task.id} (${task.name})`);
        await taskExecutionEngine.scheduleTaskExecution(task.id, false);
      });

      this.scheduledTasks.set(task.id, scheduledTask);
      console.log(`[Scheduler] 任务 ${task.id} 调度成功，cron: ${task.cronExpression}`);
    } catch (error: any) {
      console.error(`[Scheduler] 任务 ${task.id} 调度失败: ${error.message}`);
    }
  }

  unscheduleTask(taskId: string): void {
    const scheduledTask = this.scheduledTasks.get(taskId);
    if (scheduledTask) {
      scheduledTask.stop();
      this.scheduledTasks.delete(taskId);
      console.log(`[Scheduler] 取消任务调度: ${taskId}`);
    }
  }

  updateTask(task: Task): void {
    console.log(`[Scheduler] 更新任务调度: ${task.id}`);
    this.scheduleTask(task);
  }

  deleteTask(taskId: string): void {
    this.unscheduleTask(taskId);
  }

  triggerTaskManually(taskId: string): boolean {
    const task = taskStorageService.getTask(taskId);
    if (!task) {
      return false;
    }

    if (task.status === TaskStatus.COMPLETED) {
      console.log(`[Scheduler] 任务 ${taskId} 已完成，无法手动触发`);
      return false;
    }

    if (taskExecutionEngine.isTaskRunning(taskId)) {
      console.log(`[Scheduler] 任务 ${taskId} 正在执行中，无法手动触发`);
      return false;
    }

    taskExecutionEngine.scheduleTaskExecution(taskId, true);
    return true;
  }
}

export const taskSchedulerService = new TaskSchedulerService();
