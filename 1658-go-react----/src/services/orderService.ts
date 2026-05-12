import { v4 as uuidv4 } from 'uuid';
import { getDatabase } from '../database';
import { Order, User, Device } from '../types';
import { getUserById } from './userService';
import { getDeviceById, associateDeviceWithUser } from './deviceService';

const db = getDatabase();

export function getOrderById(id: string): Order | null {
  const row = db.prepare('SELECT * FROM orders WHERE id = ?').get(id) as any;
  if (!row) return null;
  return row as Order;
}

export function getOrdersByUserId(userId: string): Order[] {
  const rows = db.prepare('SELECT * FROM orders WHERE userId = ? ORDER BY createdAt DESC').all(userId) as any[];
  return rows as Order[];
}

export function getOrdersByDeviceId(deviceId: string): Order[] {
  const rows = db.prepare('SELECT * FROM orders WHERE deviceId = ? ORDER BY createdAt DESC').all(deviceId) as any[];
  return rows as Order[];
}

export function createOrder(data: {
  userId: string;
  deviceId: string;
  productInfo: string;
  amount: number;
  shippingAddress: string;
  paymentMethod: string;
}): Order {
  const user = getUserById(data.userId);
  const device = getDeviceById(data.deviceId);
  
  if (!user || !device) {
    throw new Error('User or device not found');
  }
  
  const id = uuidv4();
  const now = Date.now();
  
  associateDeviceWithUser(data.deviceId, data.userId);
  
  const riskLevel = evaluateOrderRisk(user, device);
  
  db.prepare(`
    INSERT INTO orders (
      id, userId, deviceId, productInfo, amount, 
      shippingAddress, paymentMethod, status, riskLevel, createdAt
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    data.userId,
    data.deviceId,
    data.productInfo,
    data.amount,
    data.shippingAddress,
    data.paymentMethod,
    'pending',
    riskLevel,
    now
  );
  
  return getOrderById(id)!;
}

function evaluateOrderRisk(user: User, device: Device): 'low' | 'medium' | 'high' {
  if (user.isBlacklisted || device.isBlacklisted) {
    return 'high';
  }
  
  if (user.riskScore >= 50) {
    return 'high';
  }
  
  if (user.riskScore >= 20) {
    return 'medium';
  }
  
  return 'low';
}

export function updateOrderStatus(orderId: string, status: Order['status']): Order | null {
  const order = getOrderById(orderId);
  if (!order) return null;
  
  db.prepare('UPDATE orders SET status = ? WHERE id = ?').run(status, orderId);
  return getOrderById(orderId);
}

export function updateOrderRiskLevel(orderId: string, riskLevel: Order['riskLevel']): Order | null {
  const order = getOrderById(orderId);
  if (!order) return null;
  
  db.prepare('UPDATE orders SET riskLevel = ? WHERE id = ?').run(riskLevel, orderId);
  return getOrderById(orderId);
}

export function listOrders(): Order[] {
  const rows = db.prepare('SELECT * FROM orders ORDER BY createdAt DESC').all() as any[];
  return rows as Order[];
}

export function getSuspiciousOrders(): Order[] {
  const rows = db.prepare(`
    SELECT * FROM orders 
    WHERE riskLevel IN ('medium', 'high') OR status = 'pending'
    ORDER BY createdAt DESC
  `).all() as any[];
  return rows as Order[];
}
