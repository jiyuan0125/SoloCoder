import { Router, Request, Response, NextFunction } from 'express';
import * as communityService from './services/community';
import * as userTagService from './services/userTag';
import * as broadcastService from './services/broadcast';
import { UserTier } from './types';

const router = Router();

router.use((err: any, _req: Request, res: Response, _next: NextFunction) => {
  const status = err.status || 500;
  res.status(status).json({ error: err.message });
});

// Communities
router.post('/communities', (req, res) => {
  try {
    const community = communityService.createCommunity(req.body);
    res.status(201).json(community);
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.get('/communities', (_req, res) => {
  res.json(communityService.getAllCommunities());
});

router.get('/communities/:id', (req, res) => {
  const community = communityService.getCommunityById(req.params.id);
  if (!community) {
    res.status(404).json({ error: 'Community not found' });
    return;
  }
  res.json(community);
});

router.delete('/communities/:id', (req, res) => {
  communityService.deleteCommunity(req.params.id);
  res.status(204).send();
});

router.post('/communities/:id/join', (req, res) => {
  try {
    const { userId } = req.body;
    if (!userId) {
      res.status(400).json({ error: 'userId is required' });
      return;
    }
    communityService.joinCommunity(req.params.id, userId);
    res.json({ success: true });
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.post('/communities/:id/leave', (req, res) => {
  try {
    const { userId } = req.body;
    if (!userId) {
      res.status(400).json({ error: 'userId is required' });
      return;
    }
    communityService.leaveCommunity(req.params.id, userId);
    res.json({ success: true });
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.get('/communities/:id/members', (req, res) => {
  res.json({
    userIds: communityService.getCommunityMembers(req.params.id)
  });
});

// Tags
router.post('/tags', (req, res) => {
  try {
    const tag = userTagService.createTag(req.body);
    res.status(201).json(tag);
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.get('/tags', (_req, res) => {
  res.json(userTagService.getAllTags());
});

// User tags
router.post('/users/:userId/tags/:tagId', (req, res) => {
  try {
    userTagService.assignTagToUser(req.params.userId, req.params.tagId);
    res.json({ success: true });
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.delete('/users/:userId/tags/:tagId', (req, res) => {
  userTagService.removeTagFromUser(req.params.userId, req.params.tagId);
  res.status(204).send();
});

router.get('/users/:userId/tags', (req, res) => {
  res.json(userTagService.getUserTags(req.params.userId));
});

router.get('/users/:userId/tier', (req, res) => {
  res.json({ tier: userTagService.getUserTier(req.params.userId) });
});

router.post('/users/:userId/activity', (req, res) => {
  userTagService.updateUserActivity(req.params.userId);
  res.json({ success: true });
});

router.get('/users/filter-by-tags', (req, res) => {
  const { tagIds, mode } = req.query;
  const ids = Array.isArray(tagIds) ? tagIds as string[] : typeof tagIds === 'string' ? [tagIds] : [];
  const filterMode = mode === 'AND' ? 'AND' : 'OR';
  res.json(userTagService.getUsersByTags(ids, filterMode));
});

router.get('/users/filter-by-tier', (req, res) => {
  const { tier } = req.query;
  if (!tier) {
    res.status(400).json({ error: 'tier is required' });
    return;
  }
  const validTiers = [UserTier.ACTIVE, UserTier.SILENT, UserTier.CHURNED];
  if (!validTiers.includes(tier as UserTier)) {
    res.status(400).json({ error: 'Invalid tier' });
    return;
  }
  res.json(userTagService.getUsersByTier(tier as UserTier));
});

// Broadcast tasks
router.post('/broadcast-tasks', (req, res) => {
  try {
    const task = broadcastService.createTask(req.body);
    res.status(201).json(task);
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

router.get('/broadcast-tasks', (_req, res) => {
  res.json(broadcastService.getAllTasks());
});

router.get('/broadcast-tasks/:id', (req, res) => {
  const detail = broadcastService.getTaskDetail(req.params.id);
  if (!detail) {
    res.status(404).json({ error: 'Task not found' });
    return;
  }
  res.json(detail);
});

router.post('/broadcast-tasks/:id/cancel', (req, res) => {
  try {
    broadcastService.cancelTask(req.params.id);
    res.json({ success: true });
  } catch (e: any) {
    const status = e.status || 500;
    res.status(status).json({ error: e.message });
  }
});

export default router;
