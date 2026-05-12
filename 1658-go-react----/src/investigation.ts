import { getDatabase } from './database';
import { getRelatedEntitiesForOrder, buildRelationshipGraph } from './graph';
import { setUserBlacklisted, addRiskTag, getUserById, decrementUserRiskScore } from './services/userService';
import { setDeviceBlacklisted } from './services/deviceService';
import { updateOrderStatus, getOrderById, updateOrderRiskLevel } from './services/orderService';
import { RelationshipGraph, RiskTag } from './types';

const db = getDatabase();

export interface InvestigationResult {
  orderId: string;
  markedAsFraud: boolean;
  usersBlacklisted: string[];
  devicesBlacklisted: string[];
  error?: string;
}

export function investigateOrder(orderId: string): RelationshipGraph | null {
  return buildRelationshipGraph(orderId);
}

export function markAsFraud(orderId: string): InvestigationResult {
  const entities = getRelatedEntitiesForOrder(orderId);
  if (!entities) {
    return {
      orderId,
      markedAsFraud: false,
      usersBlacklisted: [],
      devicesBlacklisted: [],
      error: 'Order not found'
    };
  }
  
  const order = getOrderById(orderId);
  if (!order) {
    return {
      orderId,
      markedAsFraud: false,
      usersBlacklisted: [],
      devicesBlacklisted: [],
      error: 'Order not found'
    };
  }
  
  try {
    const result = db.transaction(() => {
      const blacklistedUsers: string[] = [];
      const blacklistedDevices: string[] = [];
      
      const orderUser = getUserById(order.userId);
      if (orderUser && !orderUser.isBlacklisted) {
        setUserBlacklisted(orderUser.id, true, db);
        addRiskTag(orderUser.id, RiskTag.FRAUD);
        blacklistedUsers.push(orderUser.id);
      }
      
      for (const user of entities.users) {
        if (!user.isBlacklisted) {
          setUserBlacklisted(user.id, true, db);
          addRiskTag(user.id, RiskTag.FRAUD);
          blacklistedUsers.push(user.id);
        }
      }
      
      const orderDevice = db.prepare('SELECT * FROM devices WHERE id = ?').get(order.deviceId) as any;
      if (orderDevice && !orderDevice.isBlacklisted) {
        setDeviceBlacklisted(order.deviceId, true, db);
        blacklistedDevices.push(order.deviceId);
      }
      
      for (const device of entities.devices) {
        if (!device.isBlacklisted) {
          setDeviceBlacklisted(device.id, true, db);
          blacklistedDevices.push(device.id);
        }
      }
      
      updateOrderStatus(orderId, 'fraud');
      updateOrderRiskLevel(orderId, 'high');
      
      return {
        orderId,
        markedAsFraud: true,
        usersBlacklisted: [...new Set(blacklistedUsers)],
        devicesBlacklisted: [...new Set(blacklistedDevices)]
      };
    })();
    
    return result;
  } catch (error: any) {
    return {
      orderId,
      markedAsFraud: false,
      usersBlacklisted: [],
      devicesBlacklisted: [],
      error: error.message || 'Transaction failed'
    };
  }
}

export function dismissAsFalsePositive(orderId: string): InvestigationResult {
  const order = getOrderById(orderId);
  if (!order) {
    return {
      orderId,
      markedAsFraud: false,
      usersBlacklisted: [],
      devicesBlacklisted: [],
      error: 'Order not found'
    };
  }
  
  const user = getUserById(order.userId);
  if (user && !user.isBlacklisted) {
    decrementUserRiskScore(user.id);
  }
  
  updateOrderStatus(orderId, 'completed');
  updateOrderRiskLevel(orderId, 'low');
  
  return {
    orderId,
    markedAsFraud: false,
    usersBlacklisted: [],
    devicesBlacklisted: []
  };
}
