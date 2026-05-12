import { Router, Request, Response } from 'express';
import { channelService } from '../services/ChannelService';
import { ChannelType } from '../types';

const router = Router();

const VALID_TYPES: ChannelType[] = ['email', 'sms', 'inapp', 'dingtalk'];

router.post('/', (req: Request, res: Response) => {
  const { name, type, primary, backup, rateLimits } = req.body;

  if (!name || !type || !primary) {
    return res.status(400).json({ error: 'name, type, and primary are required' });
  }

  if (!VALID_TYPES.includes(type)) {
    return res.status(400).json({ error: 'Invalid channel type. Must be one of: email, sms, inapp, dingtalk' });
  }

  if (rateLimits) {
    const { perMinute, perHour, perDay } = rateLimits;
    if (perMinute <= 0 || perHour <= 0 || perDay <= 0) {
      return res.status(400).json({ error: 'Rate limits must be positive numbers' });
    }
  }

  const channel = channelService.create({
    name,
    type,
    primary,
    backup,
    rateLimits,
  });
  return res.status(201).json(channel);
});

router.get('/', (_req: Request, res: Response) => {
  const channels = channelService.list();
  return res.json(channels);
});

router.get('/:id', (req: Request, res: Response) => {
  const channel = channelService.get(req.params.id);
  if (!channel) {
    return res.status(404).json({ error: 'Channel not found' });
  }
  return res.json(channel);
});

router.put('/:id/config', (req: Request, res: Response) => {
  const { name, primary, backup, rateLimits } = req.body;

  if (name === undefined && primary === undefined && backup === undefined && rateLimits === undefined) {
    return res.status(400).json({ error: 'At least one field is required for update' });
  }

  const existing = channelService.get(req.params.id);
  if (!existing) {
    return res.status(404).json({ error: 'Channel not found' });
  }

  const updated = channelService.updateConfig(req.params.id, {
    ...(name !== undefined && { name }),
    ...(primary !== undefined && { primary }),
    ...(backup !== undefined && { backup }),
    ...(rateLimits !== undefined && { rateLimits }),
  });

  if (!updated) {
    return res.status(400).json({ error: 'Rate limits must be positive numbers' });
  }

  return res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const existing = channelService.get(req.params.id);
  if (!existing) {
    return res.status(404).json({ error: 'Channel not found' });
  }

  if (channelService.hasPendingNotifications(req.params.id)) {
    return res.status(409).json({ error: 'Cannot delete channel with pending notifications' });
  }

  channelService.delete(req.params.id);
  return res.status(204).send();
});

export { router as channelsRouter };
