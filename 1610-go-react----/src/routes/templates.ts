import { Router, Request, Response } from 'express';
import { templateService } from '../services/TemplateService';
import { ChannelType } from '../types';

const router = Router();

const VALID_TYPES: ChannelType[] = ['email', 'sms', 'inapp', 'dingtalk'];

router.post('/', (req: Request, res: Response) => {
  const { name, type, content } = req.body;

  if (!name || !type || !content) {
    return res.status(400).json({ error: 'name, type, and content are required' });
  }

  if (!VALID_TYPES.includes(type)) {
    return res.status(400).json({ error: 'Invalid channel type. Must be one of: email, sms, inapp, dingtalk' });
  }

  const template = templateService.create(name, type, content);
  return res.status(201).json(template);
});

router.get('/', (_req: Request, res: Response) => {
  const templates = templateService.list();
  return res.json(templates);
});

router.get('/:id', (req: Request, res: Response) => {
  const template = templateService.get(req.params.id);
  if (!template) {
    return res.status(404).json({ error: 'Template not found' });
  }
  return res.json(template);
});

router.put('/:id', (req: Request, res: Response) => {
  const { name, content } = req.body;

  if (name === undefined && content === undefined) {
    return res.status(400).json({ error: 'name or content is required' });
  }

  const updated = templateService.update(req.params.id, {
    ...(name !== undefined && { name }),
    ...(content !== undefined && { content }),
  });

  if (!updated) {
    return res.status(404).json({ error: 'Template not found' });
  }

  return res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const deleted = templateService.delete(req.params.id);
  if (!deleted) {
    return res.status(404).json({ error: 'Template not found' });
  }
  return res.status(204).send();
});

export { router as templatesRouter };
