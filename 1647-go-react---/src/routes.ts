import { Router, Request, Response, NextFunction } from 'express';
import { workflowService, instanceService, WorkflowError } from './services';
import {
  CreateWorkflowRequest,
  StartInstanceRequest,
  ApproveRejectRequest
} from './types';

const router = Router();

function handleAsync(fn: (req: Request, res: Response, next: NextFunction) => Promise<void> | void) {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      const result = fn(req, res, next);
      if (result instanceof Promise) {
        result.catch(next);
      }
    } catch (err) {
      next(err);
    }
  };
}

function handleError(err: any, _req: Request, res: Response, _next: NextFunction) {
  if (err instanceof WorkflowError) {
    res.status(err.statusCode).json({
      error: err.message
    });
    return;
  }
  console.error(err);
  res.status(500).json({
    error: '服务器内部错误'
  });
}

router.get('/workflows', handleAsync((_req, res) => {
  const workflows = workflowService.getAllWorkflows();
  res.json(workflows);
}));

router.post('/workflows', handleAsync((req, res) => {
  const body = req.body as CreateWorkflowRequest;
  if (!body.name || !body.startNodeId || !Array.isArray(body.nodes)) {
    throw new WorkflowError('缺少必要字段: name, startNodeId, nodes');
  }
  const workflow = workflowService.createWorkflow(body);
  res.status(201).json(workflow);
}));

router.get('/workflows/:id', handleAsync((req, res) => {
  const workflow = workflowService.getWorkflow(req.params.id);
  res.json(workflow);
}));

router.post('/workflows/:id/instances', handleAsync((req, res) => {
  const body = req.body as StartInstanceRequest;
  const instance = instanceService.startInstance(req.params.id, body || {});
  const nodeInsts = instanceService.getNodeInstances(instance.id);
  res.status(201).json({
    instance,
    nodeInstances: nodeInsts
  });
}));

router.get('/instances/:id', handleAsync((req, res) => {
  const instance = instanceService.getInstance(req.params.id);
  const nodeInsts = instanceService.getNodeInstances(req.params.id);
  res.json({
    instance,
    nodeInstances: nodeInsts
  });
}));

router.post('/instances/:instanceId/nodes/:nodeId/approve', handleAsync((req, res) => {
  const body = req.body as ApproveRejectRequest;
  if (!body.operator) {
    throw new WorkflowError('缺少操作人 operator');
  }
  instanceService.approve(req.params.instanceId, req.params.nodeId, body);
  res.json({ success: true });
}));

router.post('/instances/:instanceId/nodes/:nodeId/reject', handleAsync((req, res) => {
  const body = req.body as ApproveRejectRequest;
  if (!body.operator) {
    throw new WorkflowError('缺少操作人 operator');
  }
  instanceService.reject(req.params.instanceId, req.params.nodeId, body);
  res.json({ success: true });
}));

router.get('/instances/:id/variables', handleAsync((req, res) => {
  const variables = instanceService.getVariables(req.params.id);
  res.json(variables);
}));

router.get('/instances/:id/history', handleAsync((req, res) => {
  const history = instanceService.getHistory(req.params.id);
  res.json(history);
}));

router.use(handleError);

export default router;
