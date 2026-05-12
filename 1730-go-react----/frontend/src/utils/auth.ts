import { User, Role } from '../types';

export const getStoredUser = (): User | null => {
  const userStr = localStorage.getItem('user');
  if (userStr) {
    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  }
  return null;
};

export const getStoredToken = (): string | null => {
  return localStorage.getItem('token');
};

export const setAuth = (user: User, token: string): void => {
  localStorage.setItem('user', JSON.stringify(user));
  localStorage.setItem('token', token);
};

export const clearAuth = (): void => {
  localStorage.removeItem('user');
  localStorage.removeItem('token');
};

export const hasRole = (requiredRoles: Role | Role[]): boolean => {
  const user = getStoredUser();
  if (!user) return false;
  
  if (Array.isArray(requiredRoles)) {
    return requiredRoles.includes(user.role);
  }
  return user.role === requiredRoles;
};

export const isAdmin = (): boolean => hasRole('admin');
export const isVerifier = (): boolean => hasRole(['admin', 'verifier']);
export const isViewer = (): boolean => hasRole(['admin', 'verifier', 'viewer']);
