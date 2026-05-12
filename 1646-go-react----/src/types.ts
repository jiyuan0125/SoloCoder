export type TaskStatus = 'CREATED' | 'WAITING_DEPENDENCIES' | 'PENDING' | 'RUNNING' | 'COMPLETED' | 'TIMEOUT' | 'SKIPPED';

export type ExecutionStatus = 'RUNNING' | 'COMPLETED' | 'TIMEOUT' | 'SKIPPED';

export interface TaskGroup {
  id: string;
  name: string;
  paused: boolean;
  createdAt: Date;
}

export interface Task {
  id: string;
  name: string;
  groupId: string | null;
  cronExpression: string | null;
  intervalSeconds: number | null;
  timeoutSeconds: number;
  executionParams: string;
  status: TaskStatus;
  createdAt: Date;
  updatedAt: Date;
}

export interface TaskDependency {
  taskId: string;
  dependsOnTaskId: string;
}

export interface TaskExecution {
  id: string;
  taskId: string;
  startTime: Date;
  endTime: Date | null;
  status: ExecutionStatus;
  output: string;
}

export interface CreateTaskRequest {
  name: string;
  groupId?: string;
  cronExpression?: string;
  intervalSeconds?: number;
  timeoutSeconds: number;
  executionParams: string;
}

export interface CreateTaskGroupRequest {
  name: string;
}

export interface SetTaskDependenciesRequest {
  dependencyIds: string[];
}

export interface GetExecutionsQuery {
  startTime?: string;
  endTime?: string;
}
