import { Router, type Request, type Response } from 'express';
import {
  getAllServices,
  getServiceById,
  createService,
  updateService,
  deleteService
} from '../services/serviceStore';
import {
  createClient,
  getAllClients,
  getClientById,
  deleteClient,
  createApiKeyForClient,
  getKeysByClientId,
  getKeyStats,
  revokeApiKey
} from '../middlewares/auth';

const router = Router();

function handleServiceError(res: Response, error: Error & { statusCode?: number }): void {
  const statusCode = error.statusCode || 500;
  res.status(statusCode).json({ error: statusCode === 409 ? 'Conflict' : 'Error', message: error.message });
}

router.get('/services', (_req: Request, res: Response) => {
  const services = getAllServices();
  res.json({ services });
});

router.post('/services', (req: Request, res: Response) => {
  try {
    const { name, upstream_url, route_prefix } = req.body;
    
    if (!name || !upstream_url || !route_prefix) {
      res.status(400).json({
        error: 'Bad Request',
        message: 'name, upstream_url, and route_prefix are required'
      });
      return;
    }

    const service = createService({ name, upstream_url, route_prefix });
    res.status(201).json({ service });
  } catch (error) {
    handleServiceError(res, error as Error & { statusCode?: number });
  }
});

router.get('/services/:id', (req: Request, res: Response) => {
  const service = getServiceById(req.params.id);
  if (!service) {
    res.status(404).json({ error: 'Not Found', message: 'Service not found' });
    return;
  }
  res.json({ service });
});

router.put('/services/:id', (req: Request, res: Response) => {
  try {
    const { name, upstream_url, route_prefix } = req.body;
    const updated = updateService(req.params.id, { name, upstream_url, route_prefix });
    
    if (!updated) {
      res.status(404).json({ error: 'Not Found', message: 'Service not found' });
      return;
    }
    res.json({ service: updated });
  } catch (error) {
    handleServiceError(res, error as Error & { statusCode?: number });
  }
});

router.delete('/services/:id', (req: Request, res: Response) => {
  const deleted = deleteService(req.params.id);
  if (!deleted) {
    res.status(404).json({ error: 'Not Found', message: 'Service not found' });
    return;
  }
  res.status(204).send();
});

router.get('/clients', (_req: Request, res: Response) => {
  const clients = getAllClients();
  res.json({ clients });
});

router.post('/clients', (req: Request, res: Response) => {
  const { name, description } = req.body;
  
  if (!name) {
    res.status(400).json({
      error: 'Bad Request',
      message: 'name is required'
    });
    return;
  }

  const client = createClient(name, description);
  res.status(201).json({ client });
});

router.get('/clients/:clientId', (req: Request, res: Response) => {
  const client = getClientById(req.params.clientId);
  if (!client) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }
  res.json({ client });
});

router.delete('/clients/:clientId', (req: Request, res: Response) => {
  const deleted = deleteClient(req.params.clientId);
  if (!deleted) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }
  res.status(204).send();
});

router.get('/clients/:clientId/keys', (req: Request, res: Response) => {
  const client = getClientById(req.params.clientId);
  if (!client) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }
  
  const keys = getKeysByClientId(req.params.clientId);
  res.json({ keys });
});

router.post('/clients/:clientId/keys', (req: Request, res: Response) => {
  const client = getClientById(req.params.clientId);
  if (!client) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }

  const { expires_at, permissions } = req.body;
  
  let expiresAt: number | undefined;
  if (expires_at !== undefined && expires_at !== null) {
    if (typeof expires_at === 'string') {
      expiresAt = new Date(expires_at).getTime();
    } else if (typeof expires_at === 'number') {
      expiresAt = expires_at;
    }
  }

  const key = createApiKeyForClient(
    req.params.clientId,
    expiresAt,
    Array.isArray(permissions) ? permissions : []
  );
  res.status(201).json({ key });
});

router.delete('/clients/:clientId/keys/:keyId', (req: Request, res: Response) => {
  const client = getClientById(req.params.clientId);
  if (!client) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }

  const revoked = revokeApiKey(req.params.keyId);
  if (!revoked) {
    res.status(404).json({ error: 'Not Found', message: 'Key not found' });
    return;
  }
  res.status(204).send();
});

router.get('/clients/:clientId/keys/:keyId/stats', (req: Request, res: Response) => {
  const client = getClientById(req.params.clientId);
  if (!client) {
    res.status(404).json({ error: 'Not Found', message: 'Client not found' });
    return;
  }

  const keys = getKeysByClientId(req.params.clientId);
  const keyExists = keys.some(k => k.id === req.params.keyId);
  
  if (!keyExists) {
    res.status(404).json({ error: 'Not Found', message: 'Key not found' });
    return;
  }

  const stats = getKeyStats(req.params.keyId) || {
    key_id: req.params.keyId,
    total_requests: 0,
    successful_requests: 0,
    failed_requests: 0
  };
  res.json({ stats });
});

export default router;
