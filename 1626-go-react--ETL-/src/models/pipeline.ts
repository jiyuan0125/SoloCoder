export interface Pipeline {
  id: string;
  name: string;
  description?: string;
  isScheduled: boolean;
  scheduleCron?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePipelineRequest {
  name: string;
  description?: string;
  isScheduled?: boolean;
  scheduleCron?: string;
}

export interface UpdatePipelineRequest {
  name?: string;
  description?: string;
  isScheduled?: boolean;
  scheduleCron?: string;
}
