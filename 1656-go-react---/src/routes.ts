import express, { Request, Response } from 'express';
import { validateDeviceInfo, generateDeviceFingerprint } from './fingerprint';
import { checkDeviceTrust, getAllTrustRecords, revokeTrustById } from './trustService';
import { DeviceInfo } from './types';

const router = express.Router();

interface LoginRequest {
  userId: string;
  deviceInfo: Partial<DeviceInfo>;
  verificationCode?: string;
}

router.post('/api/login', async (req: Request, res: Response) => {
  try {
    const { userId, deviceInfo, verificationCode } = req.body as LoginRequest;

    if (!userId) {
      return res.status(400).json({ error: '用户ID不能为空' });
    }

    if (!validateDeviceInfo(deviceInfo)) {
      return res.status(400).json({ error: '设备信息不完整' });
    }

    const deviceFingerprint = generateDeviceFingerprint(deviceInfo);

    const result = await checkDeviceTrust(
      userId,
      deviceFingerprint,
      async () => {
        if (verificationCode && verificationCode.length > 0) {
          return true;
        }
        return false;
      }
    );

    if (result.message) {
      return res.status(401).json({ error: result.message });
    }

    if (result.alreadyVerifying) {
      return res.status(409).json({ error: '设备正在验证中，请稍后重试' });
    }

    if (result.requiresVerification) {
      return res.status(403).json({
        error: '需要额外验证',
        deviceFingerprint,
        message: '请提供验证码进行设备验证'
      });
    }

    if (result.trusted) {
      return res.status(200).json({
        message: '登录成功',
        deviceFingerprint,
        trusted: true
      });
    }

    return res.status(401).json({ error: '验证失败' });

  } catch (error) {
    console.error('Login error:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/api/admin/trust-records', async (_req: Request, res: Response) => {
  try {
    const records = await getAllTrustRecords();
    return res.status(200).json(records);
  } catch (error) {
    console.error('Get trust records error:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.delete('/api/admin/trust-records/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      return res.status(400).json({ error: '无效的记录ID' });
    }

    const success = await revokeTrustById(id);

    if (!success) {
      return res.status(404).json({ error: '信任记录不存在' });
    }

    return res.status(200).json({ message: '设备信任已撤销' });

  } catch (error) {
    console.error('Revoke trust error:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
