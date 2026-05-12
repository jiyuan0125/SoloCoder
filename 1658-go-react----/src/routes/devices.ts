import { Router, Request, Response } from 'express';
import {
  createDevice,
  getDeviceById,
  getDeviceByFingerprint,
  associateDeviceWithUser,
  listDevices,
  getUsersForDevice
} from '../services/deviceService';
import { getUserById } from '../services/userService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { fingerprintId } = req.body;
  if (!fingerprintId) {
    return res.status(400).json({ error: 'Fingerprint ID is required' });
  }
  const device = createDevice(fingerprintId);
  res.status(201).json(device);
});

router.get('/', (_req: Request, res: Response) => {
  const devices = listDevices();
  res.json(devices);
});

router.get('/:id', (req: Request, res: Response) => {
  const device = getDeviceById(req.params.id);
  if (!device) {
    return res.status(404).json({ error: 'Device not found' });
  }
  res.json(device);
});

router.post('/fingerprint', (req: Request, res: Response) => {
  const { fingerprintId } = req.body;
  if (!fingerprintId) {
    return res.status(400).json({ error: 'Fingerprint ID is required' });
  }
  const device = getDeviceByFingerprint(fingerprintId);
  if (!device) {
    return res.status(404).json({ error: 'Device not found' });
  }
  res.json(device);
});

router.post('/:id/associate/:userId', (req: Request, res: Response) => {
  const { id, userId } = req.params;
  
  const device = getDeviceById(id);
  const user = getUserById(userId);
  
  if (!device || !user) {
    return res.status(400).json({ error: 'Device or user not found' });
  }
  
  const association = associateDeviceWithUser(id, userId);
  if (!association) {
    return res.status(400).json({ error: 'Failed to associate' });
  }
  res.json(association);
});

router.get('/:id/users', (req: Request, res: Response) => {
  const device = getDeviceById(req.params.id);
  if (!device) {
    return res.status(404).json({ error: 'Device not found' });
  }
  const users = getUsersForDevice(req.params.id);
  res.json(users);
});

export default router;
