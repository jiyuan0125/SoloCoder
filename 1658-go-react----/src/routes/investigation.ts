import { Router, Request, Response } from 'express';
import { investigateOrder, markAsFraud, dismissAsFalsePositive } from '../investigation';
import { getOrderById } from '../services/orderService';

const router = Router();

router.get('/graph/:orderId', (req: Request, res: Response) => {
  const order = getOrderById(req.params.orderId);
  if (!order) {
    return res.status(404).json({ error: 'Order not found' });
  }
  
  const graph = investigateOrder(req.params.orderId);
  if (!graph) {
    return res.status(404).json({ error: 'Order not found' });
  }
  
  res.json(graph);
});

router.post('/fraud/:orderId', (req: Request, res: Response) => {
  const order = getOrderById(req.params.orderId);
  if (!order) {
    return res.status(404).json({ error: 'Order not found' });
  }
  
  const result = markAsFraud(req.params.orderId);
  
  if (!result.markedAsFraud) {
    if (result.error === 'Order not found') {
      return res.status(404).json({ error: result.error });
    }
    return res.status(500).json({ error: result.error });
  }
  
  res.json(result);
});

router.post('/dismiss/:orderId', (req: Request, res: Response) => {
  const order = getOrderById(req.params.orderId);
  if (!order) {
    return res.status(404).json({ error: 'Order not found' });
  }
  
  const result = dismissAsFalsePositive(req.params.orderId);
  res.json(result);
});

export default router;
