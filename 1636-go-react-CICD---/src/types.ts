export type TriggerType = 'manual' | 'schedule' | 'webhook';

export type TaskType = 'code_pull' | 'build' | 'test' | 'deploy' | 'notify';

export type TaskStatus = 'success' | 'failure' | 'skipped';

export type StageStatus = 'pending' | 'running' | 'success' | 'failure' | 'skipped';

export type PipelineStatus = 'pending' | 'running' | 'success' | 'failure' | 'cancelled';

export interface PipelineDefinition {
  name: string;
  stages: StageDefinition[];
  allowedWebhookSources?: string[];
}

export interface StageDefinition {
  name: string;
  tasks: TaskDefinition[];
}

export interface TaskDefinition {
  name: string;
  type: TaskType;
  environment?: string;
  config?: Record<string, any>;
}

export interface Task {
  id: string;
  executionId: string;
  stageId: string;
  name: string;
  type: TaskType;
  environment?: string;
  status: TaskStatus | 'pending' | 'running';
  startTime?: number;
  endTime?: number;
  logs: string;
  order: number;
}

export interface Stage {
  id: string;
  executionId: string;
  name: string;
  status: StageStatus;
  order: number;
}

export interface PipelineExecution {
  id: string;
  pipelineId: string;
  status: PipelineStatus;
  triggerType: TriggerType;
  triggerTime: number;
  startTime?: number;
  endTime?: number;
}

export interface Pipeline {
  id: string;
  name: string;
  definition: string;
  allowedWebhookSources?: string;
  createdAt: number;
}

export interface Notification {
  id: string;
  pipelineId: string;
  executionId: string;
  content: string;
  createdAt: number;
}
