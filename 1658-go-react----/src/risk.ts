import { Database } from 'better-sqlite3';
import { getDatabase } from './database';
import { getDevicesForUser, getUsersForDevice } from './services/deviceService';

export function calculateUserRiskScore(userId: string, database?: Database): number {
  const db = database || getDatabase();
  let score = 0;
  
  const userRow = db.prepare('SELECT isBlacklisted FROM users WHERE id = ?').get(userId) as any;
  if (!userRow) return 0;
  
  if (userRow.isBlacklisted) {
    return 100;
  }
  
  const userDevices = getDevicesForUser(userId);
  
  for (const device of userDevices) {
    if (device.isBlacklisted) {
      score += 30;
    }
    
    const deviceUsers = getUsersForDevice(device.id);
    for (const du of deviceUsers) {
      if (du.id !== userId && du.isBlacklisted) {
        score += 15;
      }
    }
  }
  
  const blacklistedDevicesCount = userDevices.filter(d => d.isBlacklisted).length;
  if (blacklistedDevicesCount > 0) {
    score += blacklistedDevicesCount * 10;
  }
  
  return Math.min(score, 100);
}

export function calculateDeviceRiskScore(deviceId: string, database?: Database): number {
  const db = database || getDatabase();
  let score = 0;
  
  const deviceRow = db.prepare('SELECT isBlacklisted FROM devices WHERE id = ?').get(deviceId) as any;
  if (!deviceRow) return 0;
  
  if (deviceRow.isBlacklisted) {
    return 100;
  }
  
  const deviceUsers = getUsersForDevice(deviceId);
  for (const user of deviceUsers) {
    if (user.isBlacklisted) {
      score += 25;
    } else if (user.riskScore >= 50) {
      score += 10;
    }
  }
  
  return Math.min(score, 100);
}

export function isHighRisk(userId: string, deviceId: string): boolean {
  const userScore = calculateUserRiskScore(userId);
  const deviceScore = calculateDeviceRiskScore(deviceId);
  
  return userScore >= 50 || deviceScore >= 30;
}

export function getRiskLevel(score: number): 'low' | 'medium' | 'high' {
  if (score >= 50) return 'high';
  if (score >= 20) return 'medium';
  return 'low';
}
