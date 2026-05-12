import Database from 'better-sqlite3';
import { ServiceInstance, InstanceStatus } from './types';

const db = new Database('./service-registry.db');

db.pragma('journal_mode = WAL');

export interface ServiceInstanceRow {
  id: string;
  service_name: string;
  address: string;
  port: number;
  weight: number;
  metadata: string;
  status: InstanceStatus;
  registered_at: number;
  last_heartbeat: number;
  failed_heartbeats: number;
}

export interface HeartbeatRecordRow {
  id: number;
  instance_id: string;
  timestamp: number;
  success: number;
}

function createTables(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS service_instances (
      id TEXT PRIMARY KEY,
      service_name TEXT NOT NULL,
      address TEXT NOT NULL,
      port INTEGER NOT NULL,
      weight INTEGER NOT NULL DEFAULT 1,
      metadata TEXT,
      status TEXT NOT NULL DEFAULT 'healthy',
      registered_at INTEGER NOT NULL,
      last_heartbeat INTEGER NOT NULL,
      failed_heartbeats INTEGER NOT NULL DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS heartbeat_records (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      instance_id TEXT NOT NULL,
      timestamp INTEGER NOT NULL,
      success INTEGER NOT NULL,
      FOREIGN KEY (instance_id) REFERENCES service_instances(id)
    );

    CREATE INDEX IF NOT EXISTS idx_service_name ON service_instances(service_name);
    CREATE INDEX IF NOT EXISTS idx_address_port ON service_instances(address, port);
    CREATE INDEX IF NOT EXISTS idx_heartbeat_instance ON heartbeat_records(instance_id);
  `);
}

function rowToInstance(row: ServiceInstanceRow): ServiceInstance {
  return {
    id: row.id,
    serviceName: row.service_name,
    address: row.address,
    port: row.port,
    weight: row.weight,
    metadata: JSON.parse(row.metadata || '{}'),
    status: row.status,
    registeredAt: row.registered_at,
    lastHeartbeat: row.last_heartbeat,
    failedHeartbeats: row.failed_heartbeats,
  };
}

function instanceToRow(instance: ServiceInstance): ServiceInstanceRow {
  return {
    id: instance.id,
    service_name: instance.serviceName,
    address: instance.address,
    port: instance.port,
    weight: instance.weight,
    metadata: JSON.stringify(instance.metadata),
    status: instance.status,
    registered_at: instance.registeredAt,
    last_heartbeat: instance.lastHeartbeat,
    failed_heartbeats: instance.failedHeartbeats,
  };
}

createTables();

export { db, rowToInstance, instanceToRow };
