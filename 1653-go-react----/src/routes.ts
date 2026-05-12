import express, { Request, Response } from 'express';
import {
  createPolicy,
  getPolicyById,
  getPolicyByName,
  updatePolicy,
  getUserById,
  createUser,
  changePasswordAtomically,
} from './database';
import {
  validatePassword,
  hashPassword,
  verifyPassword,
} from './passwordService';
import { checkRateLimit } from './rateLimiter';
import { CreatePolicyRequest, ChangePasswordRequest, ValidationRequest } from './types';

const router = express.Router();

router.post('/validate', async (req: Request, res: Response) => {
  try {
    const body: ValidationRequest = req.body;
    const { password, userId, policyId } = body;

    if (userId) {
      const rateCheck = checkRateLimit(userId);
      if (!rateCheck.allowed) {
        return res.status(429).json({
          error: '请求过于频繁',
          remaining: rateCheck.remaining,
          resetTime: rateCheck.resetTime,
        });
      }
    }

    const validation = await validatePassword({
      password,
      userId,
      policyId,
    });

    if (validation.error) {
      return res.status(validation.error.status).json({
        error: validation.error.message,
      });
    }

    return res.json(validation.result);
  } catch (err) {
    console.error('Validation error:', err);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/policies', async (req: Request, res: Response) => {
  try {
    const body: CreatePolicyRequest = req.body;

    if (!body.name) {
      return res.status(400).json({ error: '策略名称必填' });
    }

    const policy = createPolicy({
      name: body.name,
      minLength: body.minLength ?? 8,
      requireUppercase: body.requireUppercase ?? false,
      requireLowercase: body.requireLowercase ?? false,
      requireNumber: body.requireNumber ?? false,
      requireSpecial: body.requireSpecial ?? false,
      forbidUsername: body.forbidUsername ?? false,
      historyCount: body.historyCount ?? 3,
    });

    return res.status(201).json(policy);
  } catch (err: any) {
    console.error('Create policy error:', err);
    if (err.message?.includes('UNIQUE constraint')) {
      return res.status(409).json({ error: '策略名称已存在' });
    }
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/policies/:id', (req: Request, res: Response) => {
  try {
    const policy = getPolicyById(req.params.id);
    if (!policy) {
      return res.status(404).json({ error: '策略不存在' });
    }
    return res.json(policy);
  } catch (err) {
    console.error('Get policy error:', err);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/policies/:id', async (req: Request, res: Response) => {
  try {
    const existing = getPolicyById(req.params.id);
    if (!existing) {
      return res.status(404).json({ error: '策略不存在' });
    }

    updatePolicy(req.params.id, req.body);
    const updated = getPolicyById(req.params.id);
    return res.json(updated);
  } catch (err) {
    console.error('Update policy error:', err);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/users', async (req: Request, res: Response) => {
  try {
    const { username, password, policyId } = req.body;

    if (!username || !password) {
      return res.status(400).json({ error: '用户名和密码必填' });
    }

    const policy = policyId
      ? getPolicyById(policyId)
      : undefined;

    if (policyId && !policy) {
      return res.status(404).json({ error: '策略不存在' });
    }

    const defaultPolicy = getPolicyByName('default');
    const effectivePolicyId = policy?.id ?? defaultPolicy?.id;
    if (!effectivePolicyId) {
      return res.status(500).json({ error: '无法获取默认策略' });
    }

    const validation = await validatePassword({
      password,
      policyId: effectivePolicyId,
    });

    if (validation.error || !validation.result?.passed) {
      const status = validation.error?.status ?? 400;
      return res.status(status).json({
        error: validation.error?.message ?? '密码不符合策略要求',
        reasons: validation.result?.reasons,
      });
    }

    const passwordHash = await hashPassword(password);

    const user = createUser({
      username,
      passwordHash,
      policyId: effectivePolicyId,
    });

    return res.status(201).json({
      id: user.id,
      username: user.username,
      policyId: user.policyId,
    });
  } catch (err: any) {
    console.error('Create user error:', err);
    if (err.message?.includes('UNIQUE constraint')) {
      return res.status(409).json({ error: '用户名已存在' });
    }
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/users/change-password', async (req: Request, res: Response) => {
  try {
    const body: ChangePasswordRequest = req.body;
    const { userId, oldPassword, newPassword } = body;

    if (!userId || !newPassword) {
      return res.status(400).json({ error: '用户ID和新密码必填' });
    }

    const user = getUserById(userId);
    if (!user) {
      return res.status(404).json({ error: '用户不存在' });
    }

    if (oldPassword) {
      const oldMatch = await verifyPassword(oldPassword, user.passwordHash);
      if (!oldMatch) {
        return res.status(401).json({ error: '旧密码错误' });
      }
    }

    const validation = await validatePassword({
      password: newPassword,
      userId,
    });

    if (validation.error) {
      return res.status(validation.error.status).json({
        error: validation.error.message,
      });
    }

    if (!validation.result?.passed) {
      return res.status(400).json({
        error: '新密码不符合策略要求',
        reasons: validation.result?.reasons || [],
        score: validation.result?.score || 0,
      });
    }

    const policy = getPolicyById(user.policyId);
    if (!policy) {
      return res.status(500).json({ error: '无法获取用户策略' });
    }

    const newHash = await hashPassword(newPassword);
    changePasswordAtomically(userId, newHash, policy.historyCount);

    return res.json({
      success: true,
      message: '密码修改成功',
      score: validation.result.score,
    });
  } catch (err) {
    console.error('Change password error:', err);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
