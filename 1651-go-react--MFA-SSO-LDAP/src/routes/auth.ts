import { Router, Request, Response } from 'express';
import { requireAuth } from '../middleware/auth';
import { 
  verifyLoginAndGetToken, getUserById, enableMfa, disableMfa,
  getUserByUsername, withLoginLock, createUser
} from '../services/userService';
import { logAuth } from '../services/logService';
import { 
  getAppById, isCallbackUrlRegistered, createSsoSession, 
  buildCallbackUrl, logoutUserFromAllApps
} from '../services/ssoService';
import { verifyToken } from '../utils/auth';

const router = Router();

router.post('/login', async (req: Request, res: Response) => {
  const { username, password, totp, appId, callbackUrl } = req.body;
  
  if (!username || !password) {
    res.status(400).json({ error: '用户名和密码不能为空' });
    return;
  }
  
  if (appId) {
    const app = getAppById(appId);
    if (!app) {
      res.status(400).json({ error: '无效的应用ID' });
      return;
    }
    if (callbackUrl && !isCallbackUrlRegistered(appId, callbackUrl)) {
      res.status(400).json({ error: '回调URL未注册' });
      return;
    }
  }
  
  let userLocked = false;
  let existingUser = getUserByUsername(username);
  if (existingUser) {
    await withLoginLock(String(existingUser.id), async () => {
      const result = verifyLoginAndGetToken(username, password, totp);
      
      if (!result.success) {
        if (result.errorCode === 'forbidden') {
          userLocked = true;
          res.status(403).json({ error: result.message });
          logAuth(existingUser!.id, username, 'login', false, req);
        } else if (result.errorCode === 'mfa_required') {
          res.status(200).json({ 
            mfaRequired: true, 
            message: result.message 
          });
          logAuth(existingUser!.id, username, 'login_password_success', true, req);
        } else {
          res.status(401).json({ error: result.message });
          logAuth(existingUser!.id, username, 'login', false, req);
        }
        return;
      }
      
      logAuth(result.user.id, username, 'login', true, req);
      
      if (appId && callbackUrl) {
        const app = getAppById(appId)!;
        const sso = createSsoSession(result.user.id, username, appId, callbackUrl);
        const redirectUrl = buildCallbackUrl(callbackUrl, sso.authToken);
        res.json({ 
          token: result.token, 
          redirectUrl,
          sessionId: sso.sessionId
        });
      } else {
        res.json({ token: result.token });
      }
    });
  } else {
    res.status(401).json({ error: '用户名或密码错误' });
    logAuth(null, username, 'login', false, req);
  }
});

router.post('/token/verify', (req: Request, res: Response) => {
  const { token } = req.body;
  
  if (!token) {
    res.status(400).json({ error: 'Token不能为空' });
    return;
  }
  
  const result = verifyToken(token);
  
  if (!result.valid) {
    if (result.error === 'malformed') {
      res.status(400).json({ error: 'Token格式错误', valid: false });
    } else {
      res.status(401).json({ 
        error: result.error === 'expired' ? 'Token已过期' : 'Token签名不合法', 
        valid: false 
      });
    }
    return;
  }
  
  res.json({ 
    valid: true, 
    payload: {
      userId: result.payload.userId,
      username: result.payload.username,
      iat: result.payload.iat,
      exp: result.payload.exp
    }
  });
});

router.post('/mfa/enable', requireAuth, (req: Request, res: Response) => {
  if (!req.user) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const user = getUserById(req.user.userId);
  if (!user) {
    res.status(404).json({ error: '用户不存在' });
    return;
  }
  
  if (user.mfa_enabled) {
    res.status(400).json({ error: 'MFA已启用' });
    return;
  }
  
  const { secret, uri } = enableMfa(user.id);
  logAuth(user.id, user.username, 'mfa_enable', true, req);
  
  res.json({ secret, uri });
});

router.post('/mfa/disable', requireAuth, (req: Request, res: Response) => {
  if (!req.user) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const user = getUserById(req.user.userId);
  if (!user) {
    res.status(404).json({ error: '用户不存在' });
    return;
  }
  
  if (!user.mfa_enabled) {
    res.status(400).json({ error: 'MFA未启用' });
    return;
  }
  
  disableMfa(user.id);
  logAuth(user.id, user.username, 'mfa_disable', true, req);
  
  res.json({ success: true });
});

router.post('/sso/login', async (req: Request, res: Response) => {
  const { appId, callbackUrl, state } = req.body;
  
  if (!appId || !callbackUrl) {
    res.status(400).json({ error: 'appId和callbackUrl不能为空' });
    return;
  }
  
  const app = getAppById(appId);
  if (!app) {
    res.status(400).json({ error: '无效的应用ID' });
    return;
  }
  
  if (!isCallbackUrlRegistered(appId, callbackUrl)) {
    res.status(400).json({ error: '回调URL未注册' });
    return;
  }
  
  res.json({
    appId,
    appName: app.name,
    callbackUrl,
    state: state || null
  });
});

router.post('/sso/complete', requireAuth, (req: Request, res: Response) => {
  if (!req.user) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const { appId, callbackUrl, state } = req.body;
  
  if (!appId || !callbackUrl) {
    res.status(400).json({ error: 'appId和callbackUrl不能为空' });
    return;
  }
  
  const app = getAppById(appId);
  if (!app) {
    res.status(400).json({ error: '无效的应用ID' });
    return;
  }
  
  if (!isCallbackUrlRegistered(appId, callbackUrl)) {
    res.status(400).json({ error: '回调URL未注册' });
    return;
  }
  
  const sso = createSsoSession(req.user.userId, req.user.username, appId, callbackUrl);
  let redirectUrl = buildCallbackUrl(callbackUrl, sso.authToken);
  
  if (state) {
    const separator = redirectUrl.includes('?') ? '&' : '?';
    redirectUrl = `${redirectUrl}${separator}state=${encodeURIComponent(state)}`;
  }
  
  logAuth(req.user.userId, req.user.username, 'sso_login', true, req);
  
  res.json({
    redirectUrl,
    sessionId: sso.sessionId
  });
});

router.post('/logout', requireAuth, async (req: Request, res: Response) => {
  if (!req.user) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const result = await logoutUserFromAllApps(req.user.userId);
  logAuth(req.user.userId, req.user.username, 'logout', true, req);
  
  res.json({
    success: true,
    sessionsCleared: result.total,
    appsNotified: result.succeeded,
    appsFailed: result.failed
  });
});

router.post('/register', (req: Request, res: Response) => {
  const { username, password } = req.body;
  
  if (!username || !password) {
    res.status(400).json({ error: '用户名和密码不能为空' });
    return;
  }
  
  if (password.length < 6) {
    res.status(400).json({ error: '密码至少6个字符' });
    return;
  }
  
  const existing = getUserByUsername(username);
  if (existing) {
    res.status(409).json({ error: '用户名已存在' });
    return;
  }
  
  const user = createUser(username, password);
  logAuth(user.id, username, 'register', true, req);
  
  res.json({
    id: user.id,
    username: user.username
  });
});

router.get('/me', requireAuth, (req: Request, res: Response) => {
  if (!req.user) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const user = getUserById(req.user.userId);
  if (!user) {
    res.status(404).json({ error: '用户不存在' });
    return;
  }
  
  res.json({
    id: user.id,
    username: user.username,
    mfaEnabled: user.mfa_enabled === 1,
    createdAt: user.created_at
  });
});

export default router;
