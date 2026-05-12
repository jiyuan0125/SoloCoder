import { Router, Request, Response } from 'express';
import {
  createOrder,
  getOrderById,
  getOrdersByUserId,
  getOrdersByDeviceId,
  listOrders,
  getSuspiciousOrders
} from '../services/orderService';
import { getUserById } from '../services/userService';
import { getDeviceById } from '../services/deviceService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const { userId, deviceId, productInfo, amount, shippingAddress, paymentMethod } = req.body;
    
    if (!userId || !deviceId || !productInfo || amount === undefined || !shippingAddress || !paymentMethod) {
      return res.status(400).json({ error: 'All fields are required' });
    }
    
    const user = getUserById(userId);
    const device = getDeviceById(deviceId);
    
    if (!user || !device) {
      return res.status(400).json({ error: 'User or device not found' });
    }
    
    const order = createOrder({
      userId,
      deviceId,
      productInfo,
      amount,
      shippingAddress,
      paymentMethod
    });
    
    res.status(201).json(order);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/', (req: Request, res: Response) => {
  const { suspicious, userId, deviceId } = req.query;
  
  if (suspicious === 'true') {
    const orders = getSuspiciousOrders();
    return res.json(orders);
  }
  
  if (userId) {
    const user = getUserById(userId as string);
    if (!user) {
      return res.status(400).json({ error: 'User not found' });
    }
    const orders = getOrdersByUserId(userId as string);
    return res.json(orders);
  }
  
  if (deviceId) {
    const device = getDeviceById(deviceId as string);
    if (!device) {
      return res.status(400).json({ error: 'Device not found' });
    }
    const orders = getOrdersByDeviceId(deviceId as string);
    return res.json(orders);
  }
  
  const orders = listOrders();
  res.json(orders);
});

router.get('/:id', (req: Request, res: Response) => {
  const order = getOrderById(req.params.id);
  if (!order) {
    return res.status(404).json({ error: 'Order not found' });
  }
  res.json(order);
});

export default router;
