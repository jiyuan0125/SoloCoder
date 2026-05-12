import { Router, Request, Response } from 'express';
import { notificationService } from '../services/NotificationService';

const router = Router();

router.post('/send', async (req: Request, res: Response) => {
  const { channelId, templateId, userId, variables } = req.body;

  if (!channelId || !templateId || !userId) {
    return res.status(400).json({ error: 'channelId, templateId, and userId are required' });
  }

  const result = await notificationService.send({
    channelId,
    templateId,
    userId,
    variables: variables || {},
  });

  if (result.missingVariables && result.missingVariables.length > 0) {
    return res.status(400).json({
      error: 'Missing required variables',
      missingVariables: result.missingVariables,
    });
  }

  if (result.waitSeconds !== undefined) {
    return res.status(429).json({
      error: 'Rate limit exceeded',
      waitSeconds: result.waitSeconds,
      notificationId: result.notificationId,
    });
  }

  if (!result.success) {
    return res.status(500).json({
      error: 'Failed to send notification',
      notificationId: result.notificationId,
    });
  }

  return res.status(200).json({
    success: true,
    notificationId: result.notificationId,
    renderedContent: result.renderedContent,
  });
});

router.get('/', (_req: Request, res: Response) => {
  const notifications = notificationService.list();
  return res.json(notifications);
});

router.get('/:id', (req: Request, res: Response) => {
  const notification = notificationService.get(req.params.id);
  if (!notification) {
    return res.status(404).json({ error: 'Notification not found' });
  }
  return res.json(notification);
});

export { router as notificationsRouter };
