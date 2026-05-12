import { Router, Request, Response } from 'express';
import { store } from '../store/memoryStore';
import { AuthorizationCode, Token } from '../types';
import {
  generateAuthorizationCode,
  generateAccessToken,
  generateRefreshToken,
  generateRandomString
} from '../utils/random';
import {
  AUTHORIZATION_CODE_EXPIRES_IN_SECONDS,
  ACCESS_TOKEN_EXPIRES_IN_SECONDS,
  REFRESH_TOKEN_EXPIRES_IN_SECONDS
} from '../utils/constants';

const router = Router();

router.get('/authorize/:clientId', (req: Request, res: Response) => {
  const { clientId } = req.params;
  const { redirectUri, scope, responseType } = req.query;

  const client = store.getClientByClientId(clientId);
  if (!client) {
    return res.status(404).json({ error: '客户端不存在' });
  }

  if (responseType !== 'code') {
    return res.status(400).json({ error: '仅支持授权码流程' });
  }

  if (typeof redirectUri !== 'string' || redirectUri !== client.redirectUri) {
    return res.status(400).json({ error: '重定向URI不匹配' });
  }

  const requestedScopes = typeof scope === 'string' ? scope.split(' ') : [];
  for (const s of requestedScopes) {
    if (!client.scopes.includes(s)) {
      return res.status(400).json({ error: '请求了未注册的权限范围' });
    }
  }

  return res.status(200).json({
    message: '请登录并授权',
    clientName: client.name,
    scopes: requestedScopes
  });
});

router.post('/authorize/:clientId/confirm', (req: Request, res: Response) => {
  const { clientId } = req.params;
  const { userId, redirectUri, scope } = req.body;

  const client = store.getClientByClientId(clientId);
  if (!client) {
    return res.status(404).json({ error: '客户端不存在' });
  }

  if (!userId || typeof userId !== 'string') {
    return res.status(400).json({ error: '缺少用户ID' });
  }

  const requestedScopes = Array.isArray(scope) ? scope : (typeof scope === 'string' ? scope.split(' ') : []);
  for (const s of requestedScopes) {
    if (!client.scopes.includes(s)) {
      return res.status(400).json({ error: '请求了未注册的权限范围' });
    }
  }

  const code = generateAuthorizationCode();
  const authorizationCode: AuthorizationCode = {
    code,
    clientId,
    userId,
    scopes: requestedScopes,
    redirectUri: client.redirectUri,
    expiresAt: Date.now() + AUTHORIZATION_CODE_EXPIRES_IN_SECONDS * 1000,
    used: false,
    issuedAt: Date.now()
  };

  store.addAuthorizationCode(authorizationCode);

  const redirectUrl = new URL(client.redirectUri);
  redirectUrl.searchParams.set('code', code);

  return res.status(200).json({
    redirectUri: redirectUrl.toString()
  });
});

router.post('/authorize/:clientId/callback', (req: Request, res: Response) => {
  const { clientId } = req.params;
  const { code, clientSecret } = req.body;

  const client = store.getClientByClientId(clientId);
  if (!client) {
    return res.status(404).json({ error: '客户端不存在' });
  }

  if (!code || typeof code !== 'string') {
    return res.status(400).json({ error: '缺少授权码' });
  }

  if (!clientSecret || clientSecret !== client.clientSecret) {
    return res.status(400).json({ error: '客户端密钥错误' });
  }

  const authorizationCode = store.getAuthorizationCode(code);
  if (!authorizationCode) {
    return res.status(400).json({ error: '授权码无效' });
  }

  if (authorizationCode.used) {
    store.revokeTokensByAuthorizationCode(code);
    return res.status(400).json({ error: '授权码已被使用' });
  }

  if (Date.now() > authorizationCode.expiresAt) {
    return res.status(400).json({ error: '授权码已过期' });
  }

  authorizationCode.used = true;
  store.updateAuthorizationCode(authorizationCode);

  const accessToken = generateAccessToken();
  const refreshToken = generateRefreshToken();

  const token: Token = {
    accessToken,
    refreshToken,
    clientId,
    userId: authorizationCode.userId,
    scopes: authorizationCode.scopes,
    accessTokenExpiresAt: Date.now() + ACCESS_TOKEN_EXPIRES_IN_SECONDS * 1000,
    refreshTokenExpiresAt: Date.now() + REFRESH_TOKEN_EXPIRES_IN_SECONDS * 1000,
    accessTokenRevoked: false,
    refreshTokenRevoked: false,
    refreshTokenUsed: false,
    authorizationCode: code,
    issuedAt: Date.now()
  };

  store.addToken(token);

  return res.status(200).json({
    accessToken,
    tokenType: 'Bearer',
    expiresIn: ACCESS_TOKEN_EXPIRES_IN_SECONDS,
    refreshToken,
    scope: token.scopes.join(' ')
  });
});

router.post('/token/refresh', (req: Request, res: Response) => {
  const { refreshToken: oldRefreshToken } = req.body;

  if (!oldRefreshToken || typeof oldRefreshToken !== 'string') {
    return res.status(400).json({ error: '缺少刷新令牌' });
  }

  const token = store.getTokenByRefreshToken(oldRefreshToken);
  if (!token) {
    return res.status(400).json({ error: '刷新令牌无效' });
  }

  if (token.refreshTokenRevoked || Date.now() > token.refreshTokenExpiresAt) {
    return res.status(400).json({ error: '刷新令牌已失效' });
  }

  if (token.refreshTokenUsed) {
    store.revokeTokensByRefreshToken(oldRefreshToken);
    return res.status(400).json({ error: '刷新令牌已被使用' });
  }

  token.refreshTokenUsed = true;

  const newAccessToken = generateAccessToken();
  const newRefreshToken = generateRefreshToken();

  const newToken: Token = {
    accessToken: newAccessToken,
    refreshToken: newRefreshToken,
    clientId: token.clientId,
    userId: token.userId,
    scopes: token.scopes,
    accessTokenExpiresAt: Date.now() + ACCESS_TOKEN_EXPIRES_IN_SECONDS * 1000,
    refreshTokenExpiresAt: Date.now() + REFRESH_TOKEN_EXPIRES_IN_SECONDS * 1000,
    accessTokenRevoked: false,
    refreshTokenRevoked: false,
    refreshTokenUsed: false,
    authorizationCode: token.authorizationCode,
    issuedAt: Date.now()
  };

  store.updateToken(token);
  store.addToken(newToken);

  return res.status(200).json({
    accessToken: newAccessToken,
    tokenType: 'Bearer',
    expiresIn: ACCESS_TOKEN_EXPIRES_IN_SECONDS,
    refreshToken: newRefreshToken,
    scope: newToken.scopes.join(' ')
  });
});

router.post('/token/revoke', (req: Request, res: Response) => {
  const { accessToken } = req.body;

  if (!accessToken || typeof accessToken !== 'string') {
    return res.status(400).json({ error: '缺少访问令牌' });
  }

  const token = store.getTokenByAccessToken(accessToken);
  if (!token) {
    return res.status(404).json({ error: '令牌不存在' });
  }

  store.revokeToken(accessToken);
  store.updateToken(token);

  return res.status(200).json({ message: '令牌已撤销' });
});

export { router as oauthRouter };
