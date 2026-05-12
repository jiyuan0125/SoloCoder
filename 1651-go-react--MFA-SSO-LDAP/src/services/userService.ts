import { db } from '../database';
import { config, UserRow } from '../config';
import { 
  hashPassword, verifyPassword, generateMfaSecret, generateMfaUri,
  verifyTotp, generateToken, nowSeconds, formatMinutes
} from '../utils/auth';

const loginLocks = new Map<string, Promise<void>>();

export async function withLoginLock(userId: string, fn: () => Promise<void> | void): Promise<void> {
  const prevLock = loginLocks.get(userId);
  if (prevLock) {
    await prevLock;
  }
  
  let resolveLock: (() => void) | undefined;
  const lock = new Promise<void>((resolve) => {
    resolveLock = resolve;
  });
  loginLocks.set(userId, lock);
  
  try {
    await fn();
  } finally {
    if (resolveLock) resolveLock();
    const currentLock = loginLocks.get(userId);
    if (currentLock === lock) {
      loginLocks.delete(userId);
    }
  }
}

export function getUserByUsername(username: string): UserRow | undefined {
  return db.prepare('SELECT * FROM users WHERE username = ?').get(username) as UserRow | undefined;
}

export function getUserById(id: number): UserRow | undefined {
  return db.prepare('SELECT * FROM users WHERE id = ?').get(id) as UserRow | undefined;
}

export function createUser(username: string, password: string): UserRow {
  const passwordHash = hashPassword(password);
  const result = db.prepare(`
    INSERT INTO users (username, password_hash)
    VALUES (?, ?)
  `).run(username, passwordHash);
  
  return db.prepare('SELECT * FROM users WHERE id = ?').get(result.lastInsertRowid) as UserRow;
}

export function isLoginFrozen(user: UserRow): { frozen: boolean; remainingSeconds: number } {
  const now = nowSeconds();
  if (user.login_frozen_until > now) {
    return { frozen: true, remainingSeconds: user.login_frozen_until - now };
  }
  return { frozen: false, remainingSeconds: 0 };
}

export function isMfaFrozen(user: UserRow): { frozen: boolean; remainingSeconds: number } {
  const now = nowSeconds();
  if (user.mfa_frozen_until > now) {
    return { frozen: true, remainingSeconds: user.mfa_frozen_until - now };
  }
  return { frozen: false, remainingSeconds: 0 };
}

export function recordLoginFailure(userId: number): void {
  const user = getUserById(userId);
  if (!user) return;
  
  const now = nowSeconds();
  let currentAttempts = user.login_failed_attempts;
  
  if (user.login_frozen_until > 0 && user.login_frozen_until <= now) {
    currentAttempts = 0;
    db.prepare(`
      UPDATE users 
      SET login_failed_attempts = 0, login_frozen_until = 0 
      WHERE id = ?
    `).run(userId);
  }
  
  const newAttempts = currentAttempts + 1;
  
  if (newAttempts >= config.loginMaxFailedAttempts) {
    const frozenUntil = now + (config.loginFreezeDurationMinutes * 60);
    db.prepare(`
      UPDATE users 
      SET login_failed_attempts = ?, login_frozen_until = ? 
      WHERE id = ?
    `).run(newAttempts, frozenUntil, userId);
  } else {
    db.prepare(`
      UPDATE users 
      SET login_failed_attempts = ? 
      WHERE id = ?
    `).run(newAttempts, userId);
  }
}

export function resetLoginFailure(userId: number): void {
  db.prepare(`
    UPDATE users 
    SET login_failed_attempts = 0, login_frozen_until = 0 
    WHERE id = ?
  `).run(userId);
}

export function recordMfaFailure(userId: number): void {
  const user = getUserById(userId);
  if (!user) return;
  
  const now = nowSeconds();
  let currentAttempts = user.mfa_failed_attempts;
  
  if (user.mfa_frozen_until > 0 && user.mfa_frozen_until <= now) {
    currentAttempts = 0;
    db.prepare(`
      UPDATE users 
      SET mfa_failed_attempts = 0, mfa_frozen_until = 0 
      WHERE id = ?
    `).run(userId);
  }
  
  const newAttempts = currentAttempts + 1;
  
  if (newAttempts >= config.mfaMaxFailedAttempts) {
    const frozenUntil = now + (config.mfaFreezeDurationMinutes * 60);
    db.prepare(`
      UPDATE users 
      SET mfa_failed_attempts = ?, mfa_frozen_until = ? 
      WHERE id = ?
    `).run(newAttempts, frozenUntil, userId);
  } else {
    db.prepare(`
      UPDATE users 
      SET mfa_failed_attempts = ? 
      WHERE id = ?
    `).run(newAttempts, userId);
  }
}

export function resetMfaFailure(userId: number): void {
  db.prepare(`
    UPDATE users 
    SET mfa_failed_attempts = 0, mfa_frozen_until = 0 
    WHERE id = ?
  `).run(userId);
}

export function enableMfa(userId: number): { secret: string; uri: string } {
  const user = getUserById(userId);
  if (!user) throw new Error('User not found');
  
  const secret = generateMfaSecret();
  const uri = generateMfaUri(user.username, secret);
  
  db.prepare(`
    UPDATE users 
    SET mfa_secret = ?, mfa_enabled = 1 
    WHERE id = ?
  `).run(secret, userId);
  
  return { secret, uri };
}

export function disableMfa(userId: number): void {
  db.prepare(`
    UPDATE users 
    SET mfa_secret = NULL, mfa_enabled = 0, mfa_failed_attempts = 0, mfa_frozen_until = 0 
    WHERE id = ?
  `).run(userId);
}

export function verifyLoginAndGetToken(
  username: string, 
  password: string,
  totpCode?: string
): { 
  success: true; token: string; user: UserRow 
} | { 
  success: false; 
  errorCode: 'unauthorized' | 'forbidden' | 'mfa_required';
  message: string 
} {
  const user = getUserByUsername(username);
  
  if (!user) {
    return { 
      success: false, 
      errorCode: 'unauthorized', 
      message: '用户名或密码错误' 
    };
  }
  
  const loginFrozen = isLoginFrozen(user);
  if (loginFrozen.frozen) {
    return { 
      success: false, 
      errorCode: 'forbidden', 
      message: `账号已冻结，剩余时间：${formatMinutes(loginFrozen.remainingSeconds)}` 
    };
  }
  
  const passwordValid = verifyPassword(password, user.password_hash);
  
  if (!passwordValid) {
    recordLoginFailure(user.id);
    const updatedUser = getUserById(user.id)!;
    const frozen = isLoginFrozen(updatedUser);
    if (frozen.frozen) {
      return { 
        success: false, 
        errorCode: 'forbidden', 
        message: `账号已冻结，剩余时间：${formatMinutes(frozen.remainingSeconds)}` 
      };
    }
    return { 
      success: false, 
      errorCode: 'unauthorized', 
      message: '用户名或密码错误' 
    };
  }
  
  if (user.mfa_enabled) {
    if (!totpCode) {
      return { 
        success: false, 
        errorCode: 'mfa_required', 
        message: '需要多因素认证' 
      };
    }
    
    const mfaFrozen = isMfaFrozen(user);
    if (mfaFrozen.frozen) {
      return { 
        success: false, 
        errorCode: 'forbidden', 
        message: `TOTP验证已冻结，剩余时间：${formatMinutes(mfaFrozen.remainingSeconds)}` 
      };
    }
    
    if (!user.mfa_secret) {
      return { 
        success: false, 
        errorCode: 'unauthorized', 
        message: '验证码错误' 
      };
    }
    
    const totpValid = verifyTotp(user.mfa_secret, totpCode);
    if (!totpValid) {
      recordMfaFailure(user.id);
      const updatedUser = getUserById(user.id)!;
      const frozen = isMfaFrozen(updatedUser);
      if (frozen.frozen) {
        return { 
          success: false, 
          errorCode: 'forbidden', 
          message: `TOTP验证已冻结，剩余时间：${formatMinutes(frozen.remainingSeconds)}` 
        };
      }
      return { 
        success: false, 
        errorCode: 'unauthorized', 
        message: '验证码错误' 
      };
    }
    
    resetMfaFailure(user.id);
  }
  
  resetLoginFailure(user.id);
  const token = generateToken(user.id, user.username);
  
  return { success: true, token, user };
}
