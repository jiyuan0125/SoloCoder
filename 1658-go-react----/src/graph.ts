import { getDatabase } from './database';
import { Order, User, Device, RelationshipGraph, GraphNode, GraphEdge } from './types';
import { getOrderById, getOrdersByUserId } from './services/orderService';
import { getUserById } from './services/userService';
import { getDeviceById, getUsersForDevice, getDevicesForUser } from './services/deviceService';

interface VisitContext {
  visited: Set<string>;
  nodes: GraphNode[];
  edges: GraphEdge[];
}

function createNode(
  id: string,
  type: 'order' | 'user' | 'device',
  level: number,
  data: Partial<Order> | Partial<User> | Partial<Device>,
  isBlacklisted?: boolean
): GraphNode {
  return { id, type, level, data, isBlacklisted };
}

function processOrder(order: Order, level: number, context: VisitContext): void {
  if (context.visited.has(`order-${order.id}`)) return;
  context.visited.add(`order-${order.id}`);
  
  context.nodes.push(createNode(order.id, 'order', level, {
    amount: order.amount,
    productInfo: order.productInfo,
    status: order.status,
    riskLevel: order.riskLevel
  }));
  
  const user = getUserById(order.userId);
  if (user) {
    context.edges.push({ from: order.id, to: user.id, relation: 'created_by' });
    processUser(user, level + 1, context);
  }
  
  const device = getDeviceById(order.deviceId);
  if (device) {
    context.edges.push({ from: order.id, to: device.id, relation: 'placed_on' });
    processDevice(device, level + 1, context);
  }
}

function processUser(user: User, level: number, context: VisitContext): void {
  if (context.visited.has(`user-${user.id}`)) return;
  context.visited.add(`user-${user.id}`);
  
  context.nodes.push(createNode(user.id, 'user', level, {
    name: user.name,
    email: user.email,
    riskScore: user.riskScore,
    riskTags: user.riskTags
  }, user.isBlacklisted));
  
  if (level < 2) {
    const userOrders = getOrdersByUserId(user.id);
    for (const order of userOrders) {
      if (!context.visited.has(`order-${order.id}`)) {
        context.edges.push({ from: user.id, to: order.id, relation: 'placed' });
        processOrder(order, level + 1, context);
      }
    }
  }
  
  if (level < 2) {
    const userDevices = getDevicesForUser(user.id);
    for (const device of userDevices) {
      if (!context.visited.has(`device-${device.id}`)) {
        context.edges.push({ from: user.id, to: device.id, relation: 'used' });
        processDevice(device, level + 1, context);
      }
    }
  }
}

function processDevice(device: Device, level: number, context: VisitContext): void {
  if (context.visited.has(`device-${device.id}`)) return;
  context.visited.add(`device-${device.id}`);
  
  context.nodes.push(createNode(device.id, 'device', level, {
    fingerprintId: device.fingerprintId
  }, device.isBlacklisted));
  
  if (level < 2) {
    const deviceUsers = getUsersForDevice(device.id);
    for (const user of deviceUsers) {
      if (!context.visited.has(`user-${user.id}`)) {
        context.edges.push({ from: device.id, to: user.id, relation: 'belongs_to' });
        processUser(user, level + 1, context);
      }
    }
  }
}

export function buildRelationshipGraph(orderId: string): RelationshipGraph | null {
  const order = getOrderById(orderId);
  if (!order) return null;
  
  const context: VisitContext = {
    visited: new Set(),
    nodes: [],
    edges: []
  };
  
  processOrder(order, 0, context);
  
  return {
    nodes: context.nodes,
    edges: context.edges
  };
}

export function getRelatedEntitiesForOrder(orderId: string): {
  users: User[];
  devices: Device[];
} | null {
  const order = getOrderById(orderId);
  if (!order) return null;
  
  const context: VisitContext = {
    visited: new Set(),
    nodes: [],
    edges: []
  };
  
  processOrder(order, 0, context);
  
  const userIds = new Set<string>();
  const deviceIds = new Set<string>();
  
  for (const node of context.nodes) {
    if (node.type === 'user' && node.level > 0) {
      userIds.add(node.id);
    }
    if (node.type === 'device' && node.level > 0) {
      deviceIds.add(node.id);
    }
  }
  
  const users: User[] = [];
  for (const id of userIds) {
    const user = getUserById(id);
    if (user) users.push(user);
  }
  
  const devices: Device[] = [];
  for (const id of deviceIds) {
    const device = getDeviceById(id);
    if (device) devices.push(device);
  }
  
  return { users, devices };
}
