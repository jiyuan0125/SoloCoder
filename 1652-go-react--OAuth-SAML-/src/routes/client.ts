import { Router, Request, Response } from 'express';
import { store } from '../store/memoryStore';
import { Client } from '../types';
import { generateClientId, generateClientSecret, generateRandomString } from '../utils/random';

const router = Router();

router.post('/register', (req: Request, res: Response) => {
  const { name, redirectUri, scopes } = req.body;

  if (!name || !redirectUri || !scopes || !Array.isArray(scopes)) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  if (typeof name !== 'string' || typeof redirectUri !== 'string') {
    return res.status(400).json({ error: '字段类型错误' });
  }

  if (store.getClientByName(name)) {
    return res.status(409).json({ error: '应用名称已存在' });
  }

  const clientId = generateClientId();
  const clientSecret = generateClientSecret();

  const client: Client = {
    id: generateRandomString(16),
    name,
    clientId,
    clientSecret,
    redirectUri,
    scopes,
    createdAt: Date.now()
  };

  store.addClient(client);

  return res.status(201).json({
    clientId: client.clientId,
    clientSecret: client.clientSecret
  });
});

export { router as clientRouter };
