export interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  riskTags: string[];
  riskScore: number;
  isBlacklisted: boolean;
  createdAt: number;
}

export interface Order {
  id: string;
  userId: string;
  deviceId: string;
  productInfo: string;
  amount: number;
  shippingAddress: string;
  paymentMethod: string;
  status: 'pending' | 'completed' | 'cancelled' | 'fraud';
  riskLevel: 'low' | 'medium' | 'high';
  createdAt: number;
}

export interface Device {
  id: string;
  fingerprintId: string;
  isBlacklisted: boolean;
  createdAt: number;
}

export interface DeviceUserAssociation {
  deviceId: string;
  userId: string;
  createdAt: number;
}

export type GraphNodeType = 'order' | 'user' | 'device';

export interface GraphNode {
  id: string;
  type: GraphNodeType;
  level: number;
  data: Partial<User> | Partial<Order> | Partial<Device>;
  isBlacklisted?: boolean;
}

export interface GraphEdge {
  from: string;
  to: string;
  relation: string;
}

export interface RelationshipGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export enum RiskTag {
  FRAUD = 'fraud',
  SUSPICIOUS = 'suspicious',
  HIGH_RISK = 'high_risk',
  LOW_RISK = 'low_risk'
}
