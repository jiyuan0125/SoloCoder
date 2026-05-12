import { Router, Request, Response } from 'express';
import {
  getAllServices,
  getInstancesByServiceName,
  getHealthyInstancesByServiceName,
  getInstanceByIdAndServiceName,
  registerOrUpdateInstance,
  processHeartbeat,
  weightedRoundRobin,
  deregisterInstance,
  deregisterService,
  validateAndNormalizeRequest,
} from '../services/registryService';
import { RegisterRequest } from '../types';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const services = getAllServices();
  res.json({ services });
});

router.get('/:serviceName', (req: Request, res: Response) => {
  const { serviceName } = req.params;
  const instances = getInstancesByServiceName(serviceName);
  
  if (instances.length === 0) {
    return res.status(404).json({ error: 'Service not found' });
  }
  
  res.json({
    serviceName,
    instances,
  });
});

router.post('/register', (req: Request, res: Response) => {
  const body = req.body as Partial<RegisterRequest>;
  
  const request: RegisterRequest = {
    serviceName: body.serviceName as string,
    address: body.address as string,
    port: body.port as number,
    weight: body.weight,
    metadata: body.metadata,
  };
  
  const { errors } = validateAndNormalizeRequest(request);
  if (errors.length > 0) {
    return res.status(400).json({ error: errors.join(', ') });
  }
  
  try {
    const instance = registerOrUpdateInstance(request);
    res.status(200).json(instance);
  } catch (error) {
    if (error instanceof Error) {
      res.status(400).json({ error: error.message });
    } else {
      res.status(500).json({ error: 'Internal server error' });
    }
  }
});

router.post('/:serviceName/discover', (req: Request, res: Response) => {
  const { serviceName } = req.params;
  const allInstances = getInstancesByServiceName(serviceName);
  
  if (allInstances.length === 0) {
    return res.status(404).json({ error: 'Service not found' });
  }
  
  const healthyInstances = getHealthyInstancesByServiceName(serviceName);
  const selected = weightedRoundRobin(serviceName);
  
  res.json({
    serviceName,
    instances: healthyInstances,
    selected,
  });
});

router.get('/:serviceName/instances', (req: Request, res: Response) => {
  const { serviceName } = req.params;
  const instances = getInstancesByServiceName(serviceName);
  
  res.json({
    serviceName,
    instances,
  });
});

router.get('/:serviceName/instances/:instanceId', (req: Request, res: Response) => {
  const { serviceName, instanceId } = req.params;
  const instance = getInstanceByIdAndServiceName(instanceId, serviceName);
  
  if (!instance) {
    return res.status(404).json({ error: 'Instance not found' });
  }
  
  res.json(instance);
});

router.post('/:serviceName/instances/:instanceId/heartbeat', (req: Request, res: Response) => {
  const { instanceId } = req.params;
  const success = processHeartbeat(instanceId);
  
  if (!success) {
    return res.status(404).json({ error: 'Instance not found' });
  }
  
  res.json({ success: true });
});

router.delete('/:serviceName/deregister', (req: Request, res: Response) => {
  const { serviceName } = req.params;
  const count = deregisterService(serviceName);
  
  if (count === 0) {
    return res.status(404).json({ error: 'No healthy/unhealthy instances found for service' });
  }
  
  res.json({ success: true, deregisteredCount: count });
});

router.delete('/:serviceName/instances/:instanceId', (req: Request, res: Response) => {
  const { serviceName, instanceId } = req.params;
  const instance = getInstanceByIdAndServiceName(instanceId, serviceName);
  
  if (!instance) {
    return res.status(404).json({ error: 'Instance not found' });
  }
  
  const success = deregisterInstance(instanceId);
  res.json({ success });
});

export default router;
