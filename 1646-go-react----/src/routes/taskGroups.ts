import express, { Request, Response } from 'express';
import { dbQueries } from '../database';
import { CreateTaskGroupRequest } from '../types';

const router = express.Router();

router.post('/', async (req: Request, res: Response) => {
  try {
    const body: CreateTaskGroupRequest = req.body;
    
    if (!body.name || typeof body.name !== 'string') {
      return res.status(400).json({ error: 'Group name is required' });
    }

    const group = await dbQueries.createTaskGroup(body.name);
    res.status(201).json(group);
  } catch (error) {
    console.error('Error creating task group:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id/pause', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const group = await dbQueries.getTaskGroup(id);
    if (!group) {
      return res.status(404).json({ error: 'Task group not found' });
    }

    await dbQueries.updateTaskGroupPaused(id, true);
    res.status(200).json({ message: 'Task group paused' });
  } catch (error) {
    console.error('Error pausing task group:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id/resume', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const group = await dbQueries.getTaskGroup(id);
    if (!group) {
      return res.status(404).json({ error: 'Task group not found' });
    }

    await dbQueries.updateTaskGroupPaused(id, false);
    res.status(200).json({ message: 'Task group resumed' });
  } catch (error) {
    console.error('Error resuming task group:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
