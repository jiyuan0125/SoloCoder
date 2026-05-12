export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed';
export type StepExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';

export interface PipelineExecution {
  id: string;
  pipelineId: string;
  status: ExecutionStatus;
  startTime: string;
  endTime?: string;
  failedStepNumber?: number;
  errorMessage?: string;
  createdAt: string;
}

export interface StepExecution {
  id: string;
  executionId: string;
  pipelineId: string;
  stepNumber: number;
  status: StepExecutionStatus;
  startTime: string;
  endTime?: string;
  recordsProcessed: number;
  errorMessage?: string;
  dataConsumed: boolean;
}

export interface RetryExecutionRequest {
  stepNumber?: number;
}
