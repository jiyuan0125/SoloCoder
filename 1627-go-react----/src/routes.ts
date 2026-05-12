import { Router, Request, Response, NextFunction } from 'express';
import { QueueManager, QueueAlreadyExistsError, QueueNotFoundError, QueueFullError, InvalidStateTransitionError, NotOwnerError, MessageNotFoundError } from './queueManager';
import { CreateQueueRequest, ConsumeMessagesRequest, SendMessageRequest } from './types';
import { v4 as uuidv4 } from 'uuid';

export function createRoutes(queueManager: QueueManager): Router {
  const router = Router();

  router.use(expressJson());

  router.get('/queues', (req: Request, res: Response) => {
    const queues = queueManager.getAllQueuesInfo();
    res.json(queues);
  });

  router.post('/queues/:name', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name } = req.params;
      const body = req.body as CreateQueueRequest;
      const queue = queueManager.createQueue(
        name,
        body.maxSize,
        body.retentionTime
      );
      res.status(201).json({
        name: queue.name,
        maxSize: queue.maxSize,
        retentionTime: queue.retentionTime
      });
    } catch (error) {
      if (error instanceof QueueAlreadyExistsError) {
        res.status(409).json({ error: error.message });
      } else {
        next(error);
      }
    }
  });

  router.get('/queues/:name', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name } = req.params;
      const info = queueManager.getQueueInfo(name);
      res.json(info);
    } catch (error) {
      if (error instanceof QueueNotFoundError) {
        res.status(404).json({ error: error.message });
      } else {
        next(error);
      }
    }
  });

  router.delete('/queues/:name', (req: Request, res: Response) => {
    const { name } = req.params;
    queueManager.deleteQueue(name);
    res.status(204).send();
  });

  router.post('/queues/:name/messages', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name } = req.params;
      const body = req.body as SendMessageRequest;
      
      if (body.content === undefined) {
        return res.status(400).json({ error: 'Missing message content' });
      }

      const message = queueManager.sendMessage(name, body.content);
      res.status(201).json(message);
    } catch (error) {
      if (error instanceof QueueNotFoundError) {
        res.status(404).json({ error: error.message });
      } else if (error instanceof QueueFullError) {
        res.status(503).json({
          error: error.message,
          currentSize: error.currentSize,
          maxSize: error.maxSize
        });
      } else {
        next(error);
      }
    }
  });

  router.post('/queues/:name/messages/consume', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name } = req.params;
      const body = req.body as ConsumeMessagesRequest;
      const consumerId = body.consumerId || uuidv4();
      const batchSize = body.batchSize || 1;

      const messages = queueManager.consumeMessages(name, consumerId, batchSize);
      res.json({
        consumerId,
        messages
      });
    } catch (error) {
      if (error instanceof QueueNotFoundError) {
        res.status(404).json({ error: error.message });
      } else {
        next(error);
      }
    }
  });

  router.post('/queues/:name/messages/:msgId/ack', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name, msgId } = req.params;
      const body = req.body as { consumerId?: string };
      const consumerId = body.consumerId;

      if (!consumerId) {
        return res.status(400).json({ error: 'Missing consumerId' });
      }

      const message = queueManager.acknowledgeMessage(name, msgId, consumerId);
      res.json({
        status: 'acknowledged',
        messageId: message.id
      });
    } catch (error) {
      if (error instanceof QueueNotFoundError) {
        res.status(404).json({ error: error.message });
      } else if (error instanceof MessageNotFoundError) {
        res.status(404).json({ error: error.message });
      } else if (error instanceof NotOwnerError) {
        res.status(403).json({ error: error.message });
      } else if (error instanceof InvalidStateTransitionError) {
        res.status(400).json({ error: error.message });
      } else {
        next(error);
      }
    }
  });

  router.get('/queues/:name/dead-letter', (req: Request, res: Response, next: NextFunction) => {
    try {
      const { name } = req.params;
      const messages = queueManager.getDeadLetterMessages(name);
      res.json(messages);
    } catch (error) {
      if (error instanceof QueueNotFoundError) {
        res.status(404).json({ error: error.message });
      } else {
        next(error);
      }
    }
  });

  router.use(errorHandler);

  return router;
}

function expressJson() {
  return (req: Request, res: Response, next: NextFunction) => {
    if (req.method === 'POST' || req.method === 'PUT' || req.method === 'PATCH') {
      let data = '';
      req.on('data', chunk => {
        data += chunk;
      });
      req.on('end', () => {
        try {
          if (data) {
            (req as any).body = JSON.parse(data);
          } else {
            (req as any).body = {};
          }
          next();
        } catch (err) {
          res.status(400).json({ error: 'Invalid JSON' });
        }
      });
    } else {
      (req as any).body = {};
      next();
    }
  };
}

function errorHandler(err: any, req: Request, res: Response, next: NextFunction) {
  console.error(err);
  res.status(500).json({ error: 'Internal server error' });
}
