import { v4 as uuidv4 } from 'uuid';
import { db, rowToInstance, ServiceInstanceRow } from '../database';
import { ServiceInstance, RegisterRequest, InstanceStatus } from '../types';

export const HEARTBEAT_INTERVAL_MS = 10000;
export const MAX_FAILED_HEARTBEATS = 3;

const nextServerIndex: Record<string, number> = {};

interface NormalizedRegisterRequest {
  serviceName: string;
  address: string;
  port: number;
  weight: number;
  metadata: Record<string, string>;
}

function validateAddress(address: string): boolean {
  if (!address || typeof address !== 'string') return false;
  const trimmed = address.trim();
  if (trimmed.length === 0) return false;
  
  const hostnameRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/;
  const ipv4Regex = /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  const ipv6Regex = /^\[([0-9a-fA-F:]+)\]$/;
  
  return hostnameRegex.test(trimmed) || ipv4Regex.test(trimmed) || ipv6Regex.test(trimmed);
}

function validatePort(port: number): boolean {
  return typeof port === 'number' && Number.isInteger(port) && port >= 1 && port <= 65535;
}

function validateWeight(weight: number): boolean {
  return typeof weight === 'number' && Number.isInteger(weight) && weight >= 0;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
}

function validateRegisterRequest(request: RegisterRequest): ValidationResult {
  const errors: string[] = [];
  
  if (!request.serviceName || typeof request.serviceName !== 'string' || request.serviceName.trim().length === 0) {
    errors.push('Service name is required and must be a non-empty string');
  }
  
  if (!validateAddress(request.address)) {
    errors.push('Address is required and must be a valid hostname, IPv4 or IPv6 address');
  }
  
  if (!validatePort(request.port)) {
    errors.push('Port is required and must be a valid integer between 1 and 65535');
  }
  
  if (request.weight !== undefined && !validateWeight(request.weight)) {
    errors.push('Weight must be a non-negative integer');
  }
  
  return { valid: errors.length === 0, errors };
}

export function validateAndNormalizeRequest(request: RegisterRequest): { request: NormalizedRegisterRequest; errors: string[] } {
  const normalized: NormalizedRegisterRequest = {
    serviceName: request.serviceName,
    address: typeof request.address === 'string' ? request.address.trim() : request.address,
    port: request.port,
    weight: request.weight === undefined ? 1 : request.weight,
    metadata: request.metadata || {},
  };
  
  const result = validateRegisterRequest(normalized);
  return { request: normalized, errors: result.errors };
}

export function registerOrUpdateInstance(request: RegisterRequest): ServiceInstance {
  const { request: normalized, errors } = validateAndNormalizeRequest(request);
  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
  
  const now = Date.now();
  const existing = findInstanceByAddress(normalized.address, normalized.port);
  
  if (existing) {
    const stmt = db.prepare(`
      UPDATE service_instances
      SET service_name = ?, address = ?, port = ?, weight = ?, metadata = ?, status = ?, last_heartbeat = ?, failed_heartbeats = 0
      WHERE id = ?
    `);
    
    const newStatus: InstanceStatus = existing.status === 'deregistered' ? 'deregistered' : 'healthy';
    
    stmt.run(
      normalized.serviceName,
      normalized.address,
      normalized.port,
      normalized.weight,
      JSON.stringify(normalized.metadata),
      newStatus,
      now,
      existing.id
    );
    
    addHeartbeatRecord(existing.id, now, true);
    
    return {
      ...existing,
      serviceName: normalized.serviceName,
      weight: normalized.weight,
      metadata: normalized.metadata,
      status: newStatus,
      lastHeartbeat: now,
      failedHeartbeats: 0,
    };
  }
  
  const newId = uuidv4();
  const instance: ServiceInstance = {
    id: newId,
    serviceName: normalized.serviceName,
    address: normalized.address,
    port: normalized.port,
    weight: normalized.weight,
    metadata: normalized.metadata,
    status: 'healthy',
    registeredAt: now,
    lastHeartbeat: now,
    failedHeartbeats: 0,
  };
  
  const stmt = db.prepare(`
    INSERT INTO service_instances (id, service_name, address, port, weight, metadata, status, registered_at, last_heartbeat, failed_heartbeats)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  
  stmt.run(
    instance.id,
    instance.serviceName,
    instance.address,
    instance.port,
    instance.weight,
    JSON.stringify(instance.metadata),
    instance.status,
    instance.registeredAt,
    instance.lastHeartbeat,
    instance.failedHeartbeats
  );
  
  addHeartbeatRecord(instance.id, now, true);
  return instance;
}

export function findInstanceByAddress(address: string, port: number): ServiceInstance | null {
  const stmt = db.prepare<[string, number], ServiceInstanceRow>(`
    SELECT * FROM service_instances WHERE address = ? AND port = ?
  `);
  const row = stmt.get(address, port);
  return row ? rowToInstance(row) : null;
}

export function findInstanceById(id: string): ServiceInstance | null {
  const stmt = db.prepare<[string], ServiceInstanceRow>(`
    SELECT * FROM service_instances WHERE id = ?
  `);
  const row = stmt.get(id);
  return row ? rowToInstance(row) : null;
}

export function getAllServices(): string[] {
  const rows = db.prepare<[], { service_name: string }>(`
    SELECT DISTINCT service_name FROM service_instances ORDER BY service_name
  `).all();
  return rows.map(row => row.service_name);
}

export function getInstancesByServiceName(serviceName: string): ServiceInstance[] {
  const rows = db.prepare<[string], ServiceInstanceRow>(`
    SELECT * FROM service_instances WHERE service_name = ? ORDER BY registered_at
  `).all(serviceName);
  return rows.map(rowToInstance);
}

export function getHealthyInstancesByServiceName(serviceName: string): ServiceInstance[] {
  return getInstancesByServiceName(serviceName).filter(
    instance => instance.status === 'healthy'
  );
}

export function getInstanceByIdAndServiceName(id: string, serviceName: string): ServiceInstance | null {
  const stmt = db.prepare<[string, string], ServiceInstanceRow>(`
    SELECT * FROM service_instances WHERE id = ? AND service_name = ?
  `);
  const row = stmt.get(id, serviceName);
  return row ? rowToInstance(row) : null;
}

export function processHeartbeat(instanceId: string): boolean {
  const instance = findInstanceById(instanceId);
  if (!instance) {
    return false;
  }
  
  const now = Date.now();
  const newStatus: InstanceStatus = instance.status === 'deregistered' ? 'deregistered' : 'healthy';
  
  const stmt = db.prepare(`
    UPDATE service_instances
    SET status = ?, last_heartbeat = ?, failed_heartbeats = 0
    WHERE id = ?
  `);
  
  stmt.run(newStatus, now, instanceId);
  addHeartbeatRecord(instanceId, now, true);
  return true;
}

function addHeartbeatRecord(instanceId: string, timestamp: number, success: boolean): void {
  const stmt = db.prepare(`
    INSERT INTO heartbeat_records (instance_id, timestamp, success)
    VALUES (?, ?, ?)
  `);
  stmt.run(instanceId, timestamp, success ? 1 : 0);
}

export function checkHeartbeats(): void {
  const now = Date.now();
  const cutoff = now - HEARTBEAT_INTERVAL_MS;
  
  const rows = db.prepare<[number], ServiceInstanceRow>(`
    SELECT * FROM service_instances 
    WHERE status IN ('healthy', 'unhealthy')
      AND last_heartbeat < ?
  `).all(cutoff);
  
  for (const row of rows) {
    const newFailed = row.failed_heartbeats + 1;
    let newStatus: InstanceStatus = row.status;
    
    if (row.status === 'healthy' && newFailed >= MAX_FAILED_HEARTBEATS) {
      newStatus = 'unhealthy';
    }
    
    db.prepare(`
      UPDATE service_instances
      SET status = ?, failed_heartbeats = ?
      WHERE id = ?
    `).run(newStatus, newFailed, row.id);
    
    addHeartbeatRecord(row.id, now, false);
  }
}

export function weightedRoundRobin(serviceName: string): ServiceInstance | null {
  const healthy = getHealthyInstancesByServiceName(serviceName).filter(
    instance => instance.weight > 0
  );
  
  if (healthy.length === 0) {
    return null;
  }
  
  const totalWeight = healthy.reduce((sum, instance) => sum + instance.weight, 0);
  if (totalWeight === 0) {
    return null;
  }
  
  if (!(serviceName in nextServerIndex)) {
    nextServerIndex[serviceName] = 0;
  }
  
  let currentIndex = nextServerIndex[serviceName];
  let acc = 0;
  
  for (let i = 0; i < totalWeight; i++) {
    for (let j = 0; j < healthy.length; j++) {
      const instance = healthy[(currentIndex + j) % healthy.length];
      if (acc < instance.weight) {
        const selected = instance;
        nextServerIndex[serviceName] = (currentIndex + j + 1) % healthy.length;
        return selected;
      }
      acc -= instance.weight;
    }
  }
  
  return healthy[0];
}

export function deregisterInstance(instanceId: string): boolean {
  const instance = findInstanceById(instanceId);
  if (!instance) {
    return false;
  }
  
  db.prepare(`
    UPDATE service_instances
    SET status = 'deregistered'
    WHERE id = ?
  `).run(instanceId);
  
  return true;
}

export function deregisterService(serviceName: string): number {
  const result = db.prepare(`
    UPDATE service_instances
    SET status = 'deregistered'
    WHERE service_name = ? AND status IN ('healthy', 'unhealthy')
  `).run(serviceName);
  
  return result.changes;
}

export function deleteInstance(instanceId: string): boolean {
  const instance = findInstanceById(instanceId);
  if (!instance) {
    return false;
  }
  
  db.prepare(`
    DELETE FROM service_instances WHERE id = ?
  `).run(instanceId);
  
  return true;
}

export function updateInstance(instanceId: string, updates: Partial<ServiceInstance>): ServiceInstance | null {
  const existing = findInstanceById(instanceId);
  if (!existing) {
    return null;
  }
  
  const merged: ServiceInstance = { ...existing, ...updates };
  
  if (updates.weight !== undefined && !validateWeight(updates.weight)) {
    throw new Error('Weight must be a non-negative integer');
  }
  
  const stmt = db.prepare(`
    UPDATE service_instances
    SET service_name = ?, address = ?, port = ?, weight = ?, metadata = ?, status = ?
    WHERE id = ?
  `);
  
  stmt.run(
    merged.serviceName,
    merged.address,
    merged.port,
    merged.weight,
    JSON.stringify(merged.metadata),
    merged.status,
    instanceId
  );
  
  return findInstanceById(instanceId);
}
