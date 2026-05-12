export interface PasswordPolicy {
  id: string;
  name: string;
  minLength: number;
  requireUppercase: boolean;
  requireLowercase: boolean;
  requireNumber: boolean;
  requireSpecial: boolean;
  forbidUsername: boolean;
  historyCount: number;
  createdAt: number;
  updatedAt: number;
}

export interface User {
  id: string;
  username: string;
  passwordHash: string;
  policyId: string;
  createdAt: number;
  updatedAt: number;
}

export interface PasswordHistory {
  id: string;
  userId: string;
  passwordHash: string;
  createdAt: number;
}

export interface ValidationResult {
  passed: boolean;
  reasons: string[];
  score: number;
}

export interface ValidationRequest {
  password: string;
  userId?: string;
  policyId?: string;
}

export interface ChangePasswordRequest {
  userId: string;
  oldPassword?: string;
  newPassword: string;
}

export interface CreatePolicyRequest {
  name: string;
  minLength?: number;
  requireUppercase?: boolean;
  requireLowercase?: boolean;
  requireNumber?: boolean;
  requireSpecial?: boolean;
  forbidUsername?: boolean;
  historyCount?: number;
}
