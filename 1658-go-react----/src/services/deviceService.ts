import { v4 as uuidv4 } from 'uuid';
import { getDatabase } from '../database';
import { Device, DeviceUserAssociation, User } from '../types';
import { getUserById, updateUserRiskScore } from './userService';

const db = getDatabase();

export function getDeviceById(id: string): Device | null {
  const row = db.prepare('SELECT * FROM devices WHERE id = ?').get(id) as any;
  if (!row) return null;
  return {
    ...row,
    isBlacklisted: Boolean(row.isBlacklisted)
  };
}

export function getDeviceByFingerprint(fingerprintId: string): Device | null {
  const row = db.prepare('SELECT * FROM devices WHERE fingerprintId = ?').get(fingerprintId) as any;
  if (!row) return null;
  return {
    ...row,
    isBlacklisted: Boolean(row.isBlacklisted)
  };
}

export function createDevice(fingerprintId: string): Device {
  const existing = getDeviceByFingerprint(fingerprintId);
  if (existing) return existing;
  
  const id = uuidv4();
  const now = Date.now();
  db.prepare(`
    INSERT INTO devices (id, fingerprintId, isBlacklisted, createdAt)
    VALUES (?, ?, 0, ?)
  `).run(id, fingerprintId, now);
  
  return getDeviceById(id)!;
}

export function associateDeviceWithUser(deviceId: string, userId: string): DeviceUserAssociation | null {
  const device = getDeviceById(deviceId);
  const user = getUserById(userId);
  
  if (!device || !user) return null;
  
  const existing = db.prepare(`
    SELECT * FROM device_user_associations WHERE deviceId = ? AND userId = ?
  `).get(deviceId, userId);
  
  if (existing) return existing as DeviceUserAssociation;
  
  const now = Date.now();
  db.prepare(`
    INSERT INTO device_user_associations (deviceId, userId, createdAt)
    VALUES (?, ?, ?)
  `).run(deviceId, userId, now);
  
  return { deviceId, userId, createdAt: now };
}

export function getUsersForDevice(deviceId: string): User[] {
  const rows = db.prepare(`
    SELECT u.* FROM users u
    INNER JOIN device_user_associations dua ON u.id = dua.userId
    WHERE dua.deviceId = ?
  `).all(deviceId) as any[];
  
  return rows.map(row => ({
    ...row,
    riskTags: JSON.parse(row.riskTags),
    isBlacklisted: Boolean(row.isBlacklisted)
  }));
}

export function getDevicesForUser(userId: string): Device[] {
  const rows = db.prepare(`
    SELECT d.* FROM devices d
    INNER JOIN device_user_associations dua ON d.id = dua.deviceId
    WHERE dua.userId = ?
  `).all(userId) as any[];
  
  return rows.map(row => ({
    ...row,
    isBlacklisted: Boolean(row.isBlacklisted)
  }));
}

export function listDevices(): Device[] {
  const rows = db.prepare('SELECT * FROM devices ORDER BY createdAt DESC').all() as any[];
  return rows.map(row => ({
    ...row,
    isBlacklisted: Boolean(row.isBlacklisted)
  }));
}

export function setDeviceBlacklisted(deviceId: string, blacklisted: boolean, transaction?: any): Device | null {
  const database = transaction || db;
  const device = getDeviceById(deviceId);
  if (!device) return null;
  
  database.prepare('UPDATE devices SET isBlacklisted = ? WHERE id = ?').run(
    blacklisted ? 1 : 0,
    deviceId
  );
  
  const associatedUsers = getUsersForDevice(deviceId);
  for (const user of associatedUsers) {
    updateUserRiskScore(user.id, database);
  }
  
  return getDeviceById(deviceId);
}
