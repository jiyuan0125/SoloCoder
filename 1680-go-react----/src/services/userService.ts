import { db } from '../database';
import { User, UserRole } from '../types';

export const getUserById = (id: number): User | undefined => {
  const row = db.prepare(`
    SELECT 
      id, username, name, role, 
      is_registered as isRegistered, 
      is_verified as isVerified, 
      house_area as houseArea, 
      created_at as createdAt
    FROM users WHERE id = ?
  `).get(id) as any;
  return row;
};

export const getUserByUsername = (username: string): User | undefined => {
  const row = db.prepare(`
    SELECT 
      id, username, name, role, 
      is_registered as isRegistered, 
      is_verified as isVerified, 
      house_area as houseArea, 
      created_at as createdAt
    FROM users WHERE username = ?
  `).get(username) as any;
  return row;
};

export const isRegistered = (user: User): boolean => user.isRegistered === 1;

export const isVerified = (user: User): boolean => user.isVerified === 1;

export const isCommittee = (user: User): boolean => user.role === UserRole.COMMITTEE;

export const isExecutor = (user: User): boolean => user.role === UserRole.EXECUTOR;

export const getValidHouseArea = (user: User): number => {
  return user.houseArea;
};

export const listUsers = (): User[] => {
  const rows = db.prepare(`
    SELECT 
      id, username, name, role, 
      is_registered as isRegistered, 
      is_verified as isVerified, 
      house_area as houseArea, 
      created_at as createdAt
    FROM users
  `).all() as any[];
  return rows;
};
