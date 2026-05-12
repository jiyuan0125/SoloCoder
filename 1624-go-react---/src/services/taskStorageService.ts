import { v4 as uuidv4 } from 'uuid';
import { 
  Task, 
  Execution, 
  CreateTaskRequest, 
  TaskStatus, 
  ExecutionStatus,
  ConcurrencyStrategy
} from '../models/task';

export class TaskStorageService {
  private tasks: Map<string, Task> = new Map();
  private executions: Map<string, Execution[]> = new Map();

  createTask(request: CreateTaskRequest): Task {
    const now = new Date();
    const task: Task = {
      id: uuidv4(),
      name: request.name,
      cronExpression: request.cronExpression,
      timeout: request.timeout,
      retryCount: request.retryCount,
      params: request.params,
      concurrencyStrategy: request.concurrencyStrategy,
      status: TaskStatus.CREATED,
      createdAt: now,
      updatedAt: now
    };
    
    this.tasks.set(task.id, task);
    this.executions.set(task.id, []);
    
    return task;
  }

  getTask(id: string): Task | undefined {
    return this.tasks.get(id);
  }

  getAllTasks(): Task[] {
    return Array.from(this.tasks.values());
  }

  updateTask(id: string, updates: Partial<Task>): Task | undefined {
    const task = this.tasks.get(id);
    if (!task) return undefined;
    
    const updatedTask: Task = {
      ...task,
      ...updates,
      updatedAt: new Date()
    };
    
    this.tasks.set(id, updatedTask);
    return updatedTask;
  }

  deleteTask(id: string): boolean {
    const deleted = this.tasks.delete(id);
    if (deleted) {
      this.executions.delete(id);
    }
    return deleted;
  }

  createExecution(taskId: string, isManual: boolean = false): Execution {
    const now = new Date();
    const execution: Execution = {
      id: uuidv4(),
      taskId,
      status: ExecutionStatus.RUNNING,
      isManual,
      startTime: now,
      retryCount: 0
    };
    
    const taskExecutions = this.executions.get(taskId) || [];
    taskExecutions.push(execution);
    this.executions.set(taskId, taskExecutions);
    
    return execution;
  }

  getExecutions(taskId: string): Execution[] {
    return this.executions.get(taskId) || [];
  }

  getRunningExecution(taskId: string): Execution | undefined {
    const executions = this.executions.get(taskId) || [];
    return executions.find(e => e.status === ExecutionStatus.RUNNING);
  }

  updateExecution(taskId: string, executionId: string, updates: Partial<Execution>): Execution | undefined {
    const executions = this.executions.get(taskId);
    if (!executions) return undefined;
    
    const index = executions.findIndex(e => e.id === executionId);
    if (index === -1) return undefined;
    
    const updatedExecution: Execution = {
      ...executions[index],
      ...updates
    };
    
    executions[index] = updatedExecution;
    this.executions.set(taskId, executions);
    
    return updatedExecution;
  }
}

export const taskStorageService = new TaskStorageService();
