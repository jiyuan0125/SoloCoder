import { v4 as uuidv4 } from 'uuid';
import {
  WorkflowDefinition,
  WorkflowNode,
  NodeType,
  NodeStatus,
  InstanceStatus,
  WorkflowInstance,
  NodeInstance,
  ApprovalNodeConfig,
  ConditionNodeConfig,
  ParallelNodeConfig,
  NotificationNodeConfig,
  CreateWorkflowRequest,
  StartInstanceRequest,
  ApproveRejectRequest
} from './types';
import {
  workflowRepo,
  instanceRepo,
  nodeInstanceRepo,
  variableRepo,
  historyRepo
} from './database';

export class WorkflowError extends Error {
  constructor(message: string, public statusCode: number = 400) {
    super(message);
    this.name = 'WorkflowError';
  }
}

function validateWorkflow(nodes: WorkflowNode[]): void {
  const nodeIds = new Set<string>();
  for (const node of nodes) {
    if (!node.id || !node.type) {
      throw new WorkflowError('节点必须包含 id 和 type');
    }
    if (nodeIds.has(node.id)) {
      throw new WorkflowError(`节点 ID 重复: ${node.id}`);
    }
    nodeIds.add(node.id);

    if (node.type === NodeType.APPROVAL) {
      const config = node.config as ApprovalNodeConfig;
      if (!config.approvers || config.approvers.length === 0) {
        throw new WorkflowError(`审批节点 ${node.id} 必须指定至少一个审批人`);
      }
    }
  }
}

function findNode(workflow: WorkflowDefinition, nodeId: string): WorkflowNode {
  const node = workflow.nodes.find(n => n.id === nodeId);
  if (!node) {
    throw new WorkflowError(`节点不存在: ${nodeId}`);
  }
  return node;
}

export const workflowService = {
  createWorkflow(req: CreateWorkflowRequest): WorkflowDefinition {
    validateWorkflow(req.nodes);

    const workflow: WorkflowDefinition = workflowRepo.create({
      id: uuidv4(),
      name: req.name,
      startNodeId: req.startNodeId,
      nodes: req.nodes
    });

    return workflow;
  },

  getWorkflow(id: string): WorkflowDefinition {
    const workflow = workflowRepo.getById(id);
    if (!workflow) {
      throw new WorkflowError('工作流不存在', 404);
    }
    return workflow;
  },

  getAllWorkflows(): WorkflowDefinition[] {
    return workflowRepo.getAll();
  }
};

function processNode(
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  nodeId: string,
  parallelBranchIndex?: number
): void {
  const node = findNode(workflow, nodeId);
  const existing = nodeInstanceRepo.getByInstanceAndNode(instance.id, nodeId, parallelBranchIndex);

  if (existing && existing.status !== NodeStatus.PENDING) {
    return;
  }

  let nodeInst: NodeInstance;
  if (existing) {
    nodeInst = existing;
    nodeInst.status = NodeStatus.RUNNING;
    nodeInstanceRepo.update(nodeInst);
  } else {
    nodeInst = nodeInstanceRepo.create({
      instanceId: instance.id,
      nodeId: nodeId,
      status: NodeStatus.RUNNING,
      parallelBranchIndex
    });
  }

  historyRepo.create({
    instanceId: instance.id,
    nodeId: nodeId,
    action: 'START',
    operator: 'SYSTEM'
  });

  if (node.type === NodeType.APPROVAL) {
    if (!instance.currentNodeIds.includes(nodeId)) {
      instance.currentNodeIds.push(nodeId);
      instanceRepo.update(instance);
    }
    return;
  }

  if (node.type === NodeType.NOTIFICATION) {
    const config = node.config as NotificationNodeConfig;
    nodeInst.status = NodeStatus.COMPLETED;
    nodeInst.completedAt = Date.now();
    nodeInstanceRepo.update(nodeInst);

    historyRepo.create({
      instanceId: instance.id,
      nodeId: nodeId,
      action: 'NOTIFY',
      operator: 'SYSTEM',
      details: config.message
    });

    if (config.nextNodeId) {
      processNode(workflow, instance, config.nextNodeId);
    } else {
      checkInstanceComplete(workflow, instance);
    }
    return;
  }

  if (node.type === NodeType.CONDITION) {
    const config = node.config as ConditionNodeConfig;
    const variables = variableRepo.getAll(instance.id);
    const value = variables[config.variable];

    if (value === undefined) {
      throw new WorkflowError(`缺少变量: ${config.variable}`);
    }

    let nextNodeId = config.defaultNextNodeId;
    for (const branch of config.branches) {
      if (value === branch.condition) {
        nextNodeId = branch.nextNodeId;
        break;
      }
    }

    nodeInst.status = NodeStatus.COMPLETED;
    nodeInst.completedAt = Date.now();
    nodeInstanceRepo.update(nodeInst);

    historyRepo.create({
      instanceId: instance.id,
      nodeId: nodeId,
      action: 'CONDITION_MATCH',
      operator: 'SYSTEM',
      details: `${config.variable}=${value} -> ${nextNodeId}`
    });

    if (nextNodeId) {
      processNode(workflow, instance, nextNodeId);
    } else {
      checkInstanceComplete(workflow, instance);
    }
    return;
  }

  if (node.type === NodeType.PARALLEL) {
    const config = node.config as ParallelNodeConfig;
    nodeInst.status = NodeStatus.RUNNING;
    nodeInstanceRepo.update(nodeInst);

    historyRepo.create({
      instanceId: instance.id,
      nodeId: nodeId,
      action: 'PARALLEL_START',
      operator: 'SYSTEM',
      details: `${config.branches.length} 个分支`
    });

    if (!instance.currentNodeIds.includes(nodeId)) {
      instance.currentNodeIds.push(nodeId);
      instanceRepo.update(instance);
    }

    for (let i = 0; i < config.branches.length; i++) {
      const branch = config.branches[i];
      processNode(workflow, instance, branch.startNodeId, i);
    }
    return;
  }
}

function checkParallelComplete(
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  parallelNodeId: string
): boolean {
  const parallelNode = findNode(workflow, parallelNodeId);
  if (parallelNode.type !== NodeType.PARALLEL) return false;

  const config = parallelNode.config as ParallelNodeConfig;
  for (let i = 0; i < config.branches.length; i++) {
    const startNodeId = config.branches[i].startNodeId;
    const completed = isBranchComplete(workflow, instance, startNodeId, i);
    if (!completed) {
      return false;
    }
  }
  return true;
}

function isBranchComplete(
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  startNodeId: string,
  branchIndex: number
): boolean {
  const node = findNode(workflow, startNodeId);
  const nodeInst = nodeInstanceRepo.getByInstanceAndNode(instance.id, startNodeId, branchIndex);

  if (!nodeInst) return false;

  const terminalStatuses = [
    NodeStatus.COMPLETED,
    NodeStatus.APPROVED,
    NodeStatus.REJECTED,
    NodeStatus.SKIPPED
  ];

  if (!terminalStatuses.includes(nodeInst.status)) {
    return false;
  }

  if (node.type === NodeType.NOTIFICATION) {
    const config = node.config as NotificationNodeConfig;
    if (config.nextNodeId) {
      return isBranchComplete(workflow, instance, config.nextNodeId, branchIndex);
    }
  }

  if (node.type === NodeType.CONDITION) {
    const config = node.config as ConditionNodeConfig;
    const variables = variableRepo.getAll(instance.id);
    const value = variables[config.variable];
    let next = config.defaultNextNodeId;
    for (const branch of config.branches) {
      if (value === branch.condition) {
        next = branch.nextNodeId;
        break;
      }
    }
    if (next) {
      return isBranchComplete(workflow, instance, next, branchIndex);
    }
  }

  return true;
}

function checkInstanceComplete(workflow: WorkflowDefinition, instance: WorkflowInstance): void {
  if (instance.currentNodeIds.length === 0) {
    instance.status = InstanceStatus.COMPLETED;
    instanceRepo.update(instance);

    historyRepo.create({
      instanceId: instance.id,
      nodeId: 'END',
      action: 'INSTANCE_COMPLETED',
      operator: 'SYSTEM'
    });
  }
}

function removeCurrentNode(instance: WorkflowInstance, nodeId: string): void {
  const idx = instance.currentNodeIds.indexOf(nodeId);
  if (idx !== -1) {
    instance.currentNodeIds.splice(idx, 1);
  }
}

export const instanceService = {
  startInstance(workflowId: string, req: StartInstanceRequest): WorkflowInstance {
    const workflow = workflowService.getWorkflow(workflowId);

    const instance: WorkflowInstance = instanceRepo.create({
      id: uuidv4(),
      workflowId,
      status: InstanceStatus.RUNNING,
      currentNodeIds: []
    });

    if (req.initialVariables && Object.keys(req.initialVariables).length > 0) {
      variableRepo.setMany(instance.id, req.initialVariables);
    }

    historyRepo.create({
      instanceId: instance.id,
      nodeId: 'START',
      action: 'INSTANCE_START',
      operator: 'SYSTEM'
    });

    processNode(workflow, instance, workflow.startNodeId);

    return instance;
  },

  getInstance(id: string): WorkflowInstance {
    const instance = instanceRepo.getById(id);
    if (!instance) {
      throw new WorkflowError('实例不存在', 404);
    }
    return instance;
  },

  approve(instanceId: string, nodeId: string, req: ApproveRejectRequest): void {
    const instance = this.getInstance(instanceId);
    if (instance.status !== InstanceStatus.RUNNING) {
      throw new WorkflowError('实例已终止或已完成');
    }

    const workflow = workflowService.getWorkflow(instance.workflowId);
    const node = findNode(workflow, nodeId);

    if (node.type !== NodeType.APPROVAL) {
      throw new WorkflowError('只能对审批节点执行此操作');
    }

    if (!instance.currentNodeIds.includes(nodeId)) {
      throw new WorkflowError('节点不可操作');
    }

    const config = node.config as ApprovalNodeConfig;
    if (!config.approvers.includes(req.operator)) {
      throw new WorkflowError('操作人不是该节点的审批人');
    }

    if (req.variables && Object.keys(req.variables).length > 0) {
      variableRepo.setMany(instanceId, req.variables);
    }

    const nodeInst = nodeInstanceRepo.getByInstanceAndNode(instanceId, nodeId);
    if (!nodeInst) {
      throw new WorkflowError('节点实例不存在');
    }

    nodeInst.status = NodeStatus.APPROVED;
    nodeInst.approver = req.operator;
    nodeInst.completedAt = Date.now();
    nodeInstanceRepo.update(nodeInst);

    historyRepo.create({
      instanceId,
      nodeId,
      action: 'APPROVE',
      operator: req.operator,
      details: req.comment
    });

    removeCurrentNode(instance, nodeId);

    let nextNodeId: string | undefined;

    if (node.type === NodeType.APPROVAL) {
      const approvalConfig = node.config as ApprovalNodeConfig;
      if (typeof approvalConfig.onReject !== 'string' && approvalConfig.onReject.jumpTo) {
        // 审批通过正常推进，需要找到下一个节点
        // 这里简化处理：假设审批节点通过后通过并行节点判断或流程结束
        // 实际场景可能需要更复杂的 next 配置
      }
    }

    const inParallel = checkParallelParent(workflow, instance, nodeId);
    if (inParallel) {
      const parallelNode = findNode(workflow, inParallel);
      const parallelConfig = parallelNode.config as ParallelNodeConfig;
      const branchIndex = findBranchIndex(parallelConfig, workflow, instance, nodeId);

      if (branchIndex !== undefined) {
        let hasMore = false;
        const parallelNodeInst = nodeInstanceRepo.getByInstanceAndNode(instanceId, inParallel);
        if (parallelNodeInst) {
          for (let i = branchIndex + 1; i < parallelConfig.branches.length; i++) {
            const branchStartId = parallelConfig.branches[i].startNodeId;
            const branchNodeInst = nodeInstanceRepo.getByInstanceAndNode(instanceId, branchStartId, i);
            if (branchNodeInst && ![NodeStatus.COMPLETED, NodeStatus.APPROVED, NodeStatus.REJECTED, NodeStatus.SKIPPED].includes(branchNodeInst.status)) {
              hasMore = true;
              break;
            }
          }
        }

        if (!hasMore) {
          if (checkParallelComplete(workflow, instance, inParallel)) {
            const pni = nodeInstanceRepo.getByInstanceAndNode(instanceId, inParallel);
            if (pni) {
              pni.status = NodeStatus.COMPLETED;
              pni.completedAt = Date.now();
              nodeInstanceRepo.update(pni);
            }
            removeCurrentNode(instance, inParallel);

            historyRepo.create({
              instanceId,
              nodeId: inParallel,
              action: 'PARALLEL_COMPLETE',
              operator: 'SYSTEM'
            });

            const pConfig = parallelNode.config as ParallelNodeConfig;
            if (pConfig.nextNodeId) {
              processNode(workflow, instance, pConfig.nextNodeId);
            }
          }
        }
      }
    } else {
      // 审批通过后，尝试判断流程是否结束
      instanceRepo.update(instance);
      checkInstanceComplete(workflow, instance);
    }

    instanceRepo.update(instance);
  },

  reject(instanceId: string, nodeId: string, req: ApproveRejectRequest): void {
    const instance = this.getInstance(instanceId);
    if (instance.status !== InstanceStatus.RUNNING) {
      throw new WorkflowError('实例已终止或已完成');
    }

    const workflow = workflowService.getWorkflow(instance.workflowId);
    const node = findNode(workflow, nodeId);

    if (node.type !== NodeType.APPROVAL) {
      throw new WorkflowError('只能对审批节点执行此操作');
    }

    if (!instance.currentNodeIds.includes(nodeId)) {
      throw new WorkflowError('节点不可操作');
    }

    const config = node.config as ApprovalNodeConfig;
    if (!config.approvers.includes(req.operator)) {
      throw new WorkflowError('操作人不是该节点的审批人');
    }

    if (req.variables && Object.keys(req.variables).length > 0) {
      variableRepo.setMany(instanceId, req.variables);
    }

    const nodeInst = nodeInstanceRepo.getByInstanceAndNode(instanceId, nodeId);
    if (!nodeInst) {
      throw new WorkflowError('节点实例不存在');
    }

    nodeInst.status = NodeStatus.REJECTED;
    nodeInst.approver = req.operator;
    nodeInst.completedAt = Date.now();
    nodeInstanceRepo.update(nodeInst);

    historyRepo.create({
      instanceId,
      nodeId,
      action: 'REJECT',
      operator: req.operator,
      details: req.comment
    });

    if (config.onReject === 'terminate') {
      instance.status = InstanceStatus.TERMINATED;
      instance.currentNodeIds = [];
      instanceRepo.update(instance);

      historyRepo.create({
        instanceId,
        nodeId: 'END',
        action: 'INSTANCE_TERMINATED',
        operator: 'SYSTEM'
      });
      return;
    }

    const jumpTo = (config.onReject as { jumpTo: string }).jumpTo;
    const targetNode = findNode(workflow, jumpTo);

    if (!nodeInst.completedAt) {
      nodeInst.completedAt = Date.now();
    }
    nodeInstanceRepo.resetNodesAfter(instanceId, targetNode.id);

    instance.currentNodeIds = [jumpTo];
    instanceRepo.update(instance);

    historyRepo.create({
      instanceId,
      nodeId: jumpTo,
      action: 'JUMP_TO',
      operator: 'SYSTEM',
      details: `从 ${nodeId} 跳回`
    });

    const allNodeInsts = nodeInstanceRepo.getByInstance(instanceId);
    for (const ni of allNodeInsts) {
      if (ni.nodeId === jumpTo) {
        ni.status = NodeStatus.PENDING;
        ni.approver = undefined;
        ni.completedAt = undefined;
        nodeInstanceRepo.update(ni);
      }
    }

    processNode(workflow, instance, jumpTo);
  },

  getVariables(instanceId: string): Record<string, string> {
    this.getInstance(instanceId);
    return variableRepo.getAll(instanceId);
  },

  getHistory(instanceId: string) {
    this.getInstance(instanceId);
    return historyRepo.getByInstance(instanceId);
  },

  getNodeInstances(instanceId: string): NodeInstance[] {
    this.getInstance(instanceId);
    return nodeInstanceRepo.getByInstance(instanceId);
  }
};

function checkParallelParent(
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  nodeId: string
): string | undefined {
  for (const node of workflow.nodes) {
    if (node.type === NodeType.PARALLEL) {
      const config = node.config as ParallelNodeConfig;
      for (const branch of config.branches) {
        if (isNodeInBranch(workflow, instance, branch.startNodeId, nodeId)) {
          return node.id;
        }
      }
    }
  }
  return undefined;
}

function isNodeInBranch(
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  startNodeId: string,
  targetNodeId: string
): boolean {
  const visited = new Set<string>();
  const queue = [startNodeId];

  while (queue.length > 0) {
    const current = queue.shift()!;
    if (current === targetNodeId) return true;
    if (visited.has(current)) continue;
    visited.add(current);

    try {
      const node = findNode(workflow, current);
      if (node.type === NodeType.NOTIFICATION) {
        const config = node.config as NotificationNodeConfig;
        if (config.nextNodeId) queue.push(config.nextNodeId);
      }
      if (node.type === NodeType.CONDITION) {
        const config = node.config as ConditionNodeConfig;
        for (const branch of config.branches) {
          queue.push(branch.nextNodeId);
        }
        if (config.defaultNextNodeId) queue.push(config.defaultNextNodeId);
      }
    } catch {
      continue;
    }
  }
  return false;
}

function findBranchIndex(
  parallelConfig: ParallelNodeConfig,
  workflow: WorkflowDefinition,
  instance: WorkflowInstance,
  nodeId: string
): number | undefined {
  for (let i = 0; i < parallelConfig.branches.length; i++) {
    const branch = parallelConfig.branches[i];
    if (isNodeInBranch(workflow, instance, branch.startNodeId, nodeId)) {
      return i;
    }
  }
  return undefined;
}
