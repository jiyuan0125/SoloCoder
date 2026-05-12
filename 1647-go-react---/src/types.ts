export enum NodeType {
  APPROVAL = 'approval',
  CONDITION = 'condition',
  PARALLEL = 'parallel',
  NOTIFICATION = 'notification'
}

export enum NodeStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  COMPLETED = 'completed',
  SKIPPED = 'skipped'
}

export enum InstanceStatus {
  RUNNING = 'running',
  COMPLETED = 'completed',
  TERMINATED = 'terminated'
}

export interface ApprovalNodeConfig {
  approvers: string[];
  onReject: 'terminate' | { jumpTo: string };
}

export interface ConditionBranch {
  condition: string;
  nextNodeId: string;
}

export interface ConditionNodeConfig {
  variable: string;
  branches: ConditionBranch[];
  defaultNextNodeId: string;
}

export interface ParallelBranch {
  startNodeId: string;
}

export interface ParallelNodeConfig {
  branches: ParallelBranch[];
  nextNodeId: string;
}

export interface NotificationNodeConfig {
  message: string;
  recipients?: string[];
  nextNodeId: string;
}

export type NodeConfig =
  | ApprovalNodeConfig
  | ConditionNodeConfig
  | ParallelNodeConfig
  | NotificationNodeConfig;

export interface WorkflowNode {
  id: string;
  type: NodeType;
  config: NodeConfig;
}

export interface WorkflowDefinition {
  id: string;
  name: string;
  startNodeId: string;
  nodes: WorkflowNode[];
  createdAt: number;
}

export interface NodeInstance {
  id: string;
  instanceId: string;
  nodeId: string;
  status: NodeStatus;
  approver?: string;
  completedAt?: number;
  parallelBranchIndex?: number;
}

export interface WorkflowInstance {
  id: string;
  workflowId: string;
  status: InstanceStatus;
  currentNodeIds: string[];
  createdAt: number;
}

export interface WorkflowVariable {
  instanceId: string;
  key: string;
  value: string;
}

export interface HistoryRecord {
  id: string;
  instanceId: string;
  nodeId: string;
  action: string;
  operator: string;
  timestamp: number;
  details?: string;
}

export interface CreateWorkflowRequest {
  name: string;
  startNodeId: string;
  nodes: WorkflowNode[];
}

export interface StartInstanceRequest {
  initialVariables?: Record<string, string>;
}

export interface ApproveRejectRequest {
  operator: string;
  comment?: string;
  variables?: Record<string, string>;
}
