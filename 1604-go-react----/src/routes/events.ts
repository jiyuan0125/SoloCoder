import { Router, Request, Response } from 'express';
import { memoryStore } from '../storage/memoryStore';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { userId, eventName, occurredAt, pagePath, properties } = req.body;

  if (!eventName || !occurredAt) {
    return res.status(400).json({
      error: '事件名称和发生时间不能为空',
    });
  }

  try {
    const occurredDate = new Date(occurredAt);
    if (isNaN(occurredDate.getTime())) {
      return res.status(400).json({
        error: '发生时间格式无效',
      });
    }

    const event = memoryStore.addEvent({
      userId: String(userId),
      eventName: String(eventName),
      occurredAt: occurredDate,
      pagePath: String(pagePath || ''),
      properties: properties || {},
    });

    res.status(201).json(event);
  } catch (error) {
    res.status(500).json({ error: '创建事件失败' });
  }
});

router.get('/', (req: Request, res: Response) => {
  const { userId, eventName, startTime, endTime } = req.query;

  try {
    const events = memoryStore.getEvents({
      userId: userId ? String(userId) : undefined,
      eventName: eventName ? String(eventName) : undefined,
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
    });

    res.json(events);
  } catch (error) {
    res.status(500).json({ error: '查询事件失败' });
  }
});

router.get('/types', (req: Request, res: Response) => {
  res.json(memoryStore.getEventTypes());
});

router.post('/types', (req: Request, res: Response) => {
  const { eventName } = req.body;
  if (!eventName) {
    return res.status(400).json({ error: '事件名称不能为空' });
  }
  memoryStore.addEventType(String(eventName));
  res.json({ success: true });
});

router.delete('/types/:eventName', (req: Request, res: Response) => {
  const { eventName } = req.params;
  
  if (memoryStore.isEventTypeUsedInFunnel(eventName)) {
    return res.status(400).json({
      error: '该事件类型被漏斗步骤引用，无法删除',
    });
  }

  const deleted = memoryStore.deleteEventType(eventName);
  if (deleted) {
    res.json({ success: true });
  } else {
    res.status(404).json({ error: '事件类型不存在' });
  }
});

export default router;
