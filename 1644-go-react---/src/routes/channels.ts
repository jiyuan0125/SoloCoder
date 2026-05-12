import { Router, Request, Response } from 'express';
import { ChannelType, ChannelConfig } from '../types';
import { getChannelConfig, updateChannelConfig } from '../services/channelService';

const router = Router();

const validChannelTypes: ChannelType[] = ['email', 'sms', 'webhook'];

router.put('/:type/config', async (req: Request, res: Response) => {
  try {
    const { type } = req.params as { type: ChannelType };

    if (!validChannelTypes.includes(type)) {
      return res.status(400).json({
        error: 'Invalid channel type',
        message: 'Channel type must be one of: email, sms, webhook'
      });
    }

    const {
      enabled,
      rateLimitPerMinute,
      emailRecipients,
      smsNumbers,
      webhookUrl
    } = req.body;

    if (enabled === undefined || enabled === null) {
      return res.status(400).json({
        error: 'Missing required fields',
        message: 'enabled is required'
      });
    }

    if (rateLimitPerMinute === undefined || rateLimitPerMinute === null) {
      return res.status(400).json({
        error: 'Missing required fields',
        message: 'rateLimitPerMinute is required'
      });
    }

    if (typeof rateLimitPerMinute !== 'number' || rateLimitPerMinute < 0) {
      return res.status(400).json({
        error: 'Invalid rate limit',
        message: 'rateLimitPerMinute must be a non-negative number'
      });
    }

    const config: ChannelConfig = {
      type,
      enabled: Boolean(enabled),
      rateLimitPerMinute: rateLimitPerMinute,
      emailRecipients,
      smsNumbers,
      webhookUrl
    };

    const updatedConfig = await updateChannelConfig(config);
    res.json(updatedConfig);
  } catch (err) {
    console.error('Failed to update channel config:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:type/config', async (req: Request, res: Response) => {
  try {
    const { type } = req.params as { type: ChannelType };

    if (!validChannelTypes.includes(type)) {
      return res.status(400).json({
        error: 'Invalid channel type',
        message: 'Channel type must be one of: email, sms, webhook'
      });
    }

    const config = await getChannelConfig(type);
    if (!config) {
      return res.status(404).json({ error: 'Channel config not found' });
    }

    res.json(config);
  } catch (err) {
    console.error('Failed to get channel config:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
