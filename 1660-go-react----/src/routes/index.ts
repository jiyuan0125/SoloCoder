import { Router, Request, Response } from 'express';
import {
  submitContent,
  getPost,
  updatePostStatus,
  getAuditQueue,
  getAuditQueueCount,
  getUser,
  getBlacklistEntry,
  setWhitelist,
  markAsBlacklist,
  removeFromBlacklist,
} from '../services/antiSpam';
import type { ContentType, PostStatus } from '../types';

const router = Router();

router.post('/content', (req: Request, res: Response) => {
  const { user_id, content, content_type } = req.body as {
    user_id: string;
    content: string;
    content_type: ContentType;
  };

  if (!user_id || !content || !content_type) {
    return res.status(400).json({ error: 'missing_required_fields' });
  }

  if (content_type !== 'post' && content_type !== 'comment') {
    return res.status(400).json({ error: 'invalid_content_type' });
  }

  const result = submitContent(user_id, content, content_type);

  const responseBody: Record<string, unknown> = {
    success: result.success,
    message: result.message,
  };

  if (result.post) {
    responseBody.post = result.post;
  }

  if (result.exceededDimension) {
    responseBody.exceeded_dimension = result.exceededDimension;
  }

  if (result.reasons) {
    responseBody.reasons = result.reasons;
  }

  res.status(result.code).json(responseBody);
});

router.get('/posts/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const post = getPost(id);
  if (!post) {
    return res.status(404).json({ error: 'content_not_found' });
  }
  res.json(post);
});

router.patch('/posts/:id/status', (req: Request, res: Response) => {
  const { id } = req.params;
  const { status } = req.body as { status: PostStatus };

  const existingPost = getPost(id);
  if (!existingPost) {
    return res.status(404).json({ error: 'content_not_found' });
  }

  const validStatuses: PostStatus[] = [
    'published',
    'pending_review',
    'suspicious',
    'blocked',
    'publisher_blocked',
    'approved',
    'rejected',
  ];

  if (!validStatuses.includes(status)) {
    return res.status(400).json({ error: 'invalid_status' });
  }

  updatePostStatus(id, status);
  const updated = getPost(id);
  res.json(updated);
});

router.get('/audit-queue', (req: Request, res: Response) => {
  const limit = parseInt(req.query.limit as string) || 100;
  const offset = parseInt(req.query.offset as string) || 0;
  const items = getAuditQueue(limit, offset);
  const total = getAuditQueueCount();
  res.json({ total, limit, offset, items });
});

router.get('/users/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const user = getUser(id);
  if (!user) {
    return res.status(404).json({ error: 'user_not_found' });
  }
  const blacklistEntry = getBlacklistEntry(id);
  res.json({
    ...user,
    blacklist_entry: blacklistEntry || null,
    is_blacklisted: !!blacklistEntry,
  });
});

router.post('/users/:id/whitelist', (req: Request, res: Response) => {
  const { id } = req.params;
  const { is_whitelist } = req.body as { is_whitelist: boolean };

  const user = getUser(id);
  if (!user) {
    return res.status(404).json({ error: 'user_not_found' });
  }

  setWhitelist(id, !!is_whitelist);
  const updated = getUser(id);
  res.json(updated);
});

router.post('/users/:id/blacklist', (req: Request, res: Response) => {
  const { id } = req.params;
  const { duration_seconds, reason } = req.body as {
    duration_seconds?: number;
    reason?: string;
  };

  const user = getUser(id);
  if (!user) {
    return res.status(404).json({ error: 'user_not_found' });
  }

  const durationMs =
    duration_seconds !== undefined && duration_seconds !== null
      ? duration_seconds * 1000
      : null;

  const result = markAsBlacklist(id, durationMs, reason || null);
  const entry = getBlacklistEntry(id);

  if (result.success) {
    res.status(200).json({
      blacklist_entry: entry,
      cleanup: {
        total: result.totalCleaned,
        success: result.successCleaned,
        failed: result.failedCleaned,
      },
    });
  } else {
    res.status(207).json({
      blacklist_entry: entry,
      cleanup: {
        total: result.totalCleaned,
        success: result.successCleaned,
        failed: result.failedCleaned,
      },
      message: 'blacklist_applied_but_partial_cleanup_failure',
    });
  }
});

router.delete('/users/:id/blacklist', (req: Request, res: Response) => {
  const { id } = req.params;
  const user = getUser(id);
  if (!user) {
    return res.status(404).json({ error: 'user_not_found' });
  }
  const entry = getBlacklistEntry(id);
  if (!entry) {
    return res.status(404).json({ error: 'user_not_in_blacklist' });
  }
  removeFromBlacklist(id);
  res.json({ message: 'removed_from_blacklist' });
});

export default router;