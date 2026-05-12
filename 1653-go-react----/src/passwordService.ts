import bcrypt from 'bcrypt';
import { PasswordPolicy, User, ValidationResult } from './types';
import {
  getPolicyById,
  getPolicyByName,
  getUserById,
  getUserPasswordHistory,
} from './database';

const MAX_CALL_DEPTH = 2;
const SALT_ROUNDS = 10;

export const callDepthStore = {
  current: 0,
};

export function calculateStrength(password: string): number {
  let score = 0;

  if (/[A-Z]/.test(password)) score += 20;
  if (/[0-9]/.test(password)) score += 20;
  if (/[^A-Za-z0-9]/.test(password)) score += 20;
  if (password.length > 12) score += 20;
  if (password.length > 16) score += 20;

  return Math.min(score, 100);
}

export function validateAgainstPolicy(
  password: string,
  policy: PasswordPolicy,
  username?: string
): ValidationResult {
  const reasons: string[] = [];

  if (password.length < policy.minLength) {
    reasons.push(`密码长度至少为 ${policy.minLength} 位`);
  }

  if (policy.requireUppercase && !/[A-Z]/.test(password)) {
    reasons.push('密码必须包含大写字母');
  }

  if (policy.requireLowercase && !/[a-z]/.test(password)) {
    reasons.push('密码必须包含小写字母');
  }

  if (policy.requireNumber && !/[0-9]/.test(password)) {
    reasons.push('密码必须包含数字');
  }

  if (policy.requireSpecial && !/[^A-Za-z0-9]/.test(password)) {
    reasons.push('密码必须包含特殊字符');
  }

  if (policy.forbidUsername && username) {
    if (password.toLowerCase().includes(username.toLowerCase())) {
      reasons.push('密码不能包含用户名');
    }
  }

  return {
    passed: reasons.length === 0,
    reasons,
    score: calculateStrength(password),
  };
}

export async function checkPasswordHistory(
  password: string,
  userId: string,
  historyCount: number
): Promise<boolean> {
  const history = getUserPasswordHistory(userId);
  const recentHistory = history.slice(0, historyCount);

  for (const record of recentHistory) {
    const match = await bcrypt.compare(password, record.passwordHash);
    if (match) {
      return true;
    }
  }

  return false;
}

export async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, SALT_ROUNDS);
}

export async function verifyPassword(
  password: string,
  hash: string
): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

export interface ValidateOptions {
  password: string;
  userId?: string;
  policyId?: string;
}

export async function validatePassword(
  options: ValidateOptions
): Promise<{
  result?: ValidationResult;
  error?: { status: number; message: string };
}> {
  if (callDepthStore.current > MAX_CALL_DEPTH) {
    return { error: { status: 500, message: '调用链异常' } };
  }

  callDepthStore.current++;

  try {
    const { password, userId, policyId } = options;

    if (!password || password.length === 0 || password.length > 128) {
      return { error: { status: 400, message: '密码为空或超过128位' } };
    }

    let policy: PasswordPolicy | undefined;

    if (policyId) {
      policy = getPolicyById(policyId);
      if (!policy) {
        return { error: { status: 404, message: '策略ID不存在' } };
      }
    } else if (userId) {
      const user = getUserById(userId);
      if (!user) {
        return { error: { status: 404, message: '用户ID不存在' } };
      }
      policy = getPolicyById(user.policyId);
    } else {
      policy = getPolicyByName('default');
    }

    if (!policy) {
      return { error: { status: 500, message: '无法获取密码策略' } };
    }

    const username = userId
      ? getUserById(userId)?.username
      : undefined;

    const result = validateAgainstPolicy(password, policy, username);

    if (result.passed && userId) {
      const inHistory = await checkPasswordHistory(
        password,
        userId,
        policy.historyCount
      );
      if (inHistory) {
        result.passed = false;
        result.reasons.push('密码不能使用最近使用过的密码');
      }
    }

    return { result };
  } finally {
    callDepthStore.current--;
  }
}

export { MAX_CALL_DEPTH };
