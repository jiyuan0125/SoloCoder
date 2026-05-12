export enum TaskStatus {
  CREATED = 'created',
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  TIMEOUT = 'timeout',
  FAILED = 'failed'
}

export enum ConcurrencyStrategy {
  SKIP = 'skip',
  WAIT = 'wait',
  FORCE_TERMINATE = 'force_terminate'
}

export interface TaskParams {
  [key: string]: any;
}

export interface Task {
  id: string;
  name: string;
  cronExpression: string;
  timeout: number;
  retryCount: number;
  params: TaskParams;
  concurrencyStrategy: ConcurrencyStrategy;
  status: TaskStatus;
  createdAt: Date;
  updatedAt: Date;
}

export enum ExecutionStatus {
  RUNNING = 'running',
  COMPLETED = 'completed',
  TIMEOUT = 'timeout',
  FAILED = 'failed'
}

export interface Execution {
  id: string;
  taskId: string;
  status: ExecutionStatus;
  isManual: boolean;
  startTime: Date;
  endTime?: Date;
  result?: any;
  error?: string;
  retryCount: number;
}

export interface CreateTaskRequest {
  name: string;
  cronExpression: string;
  timeout: number;
  retryCount: number;
  params: TaskParams;
  concurrencyStrategy: ConcurrencyStrategy;
}

export interface UpdateTaskRequest {
  name?: string;
  cronExpression?: string;
  timeout?: number;
  retryCount?: number;
  params?: TaskParams;
  concurrencyStrategy?: ConcurrencyStrategy;
}
