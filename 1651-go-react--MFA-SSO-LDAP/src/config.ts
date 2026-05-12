export const config = {
  port: process.env.PORT ? parseInt(process.env.PORT, 10) : 3000,
  jwtSecret: process.env.JWT_SECRET || 'auth-center-jwt-secret-key-change-me',
  jwtExpiresIn: '1h',
  jwtExpiresInSeconds: 3600,
  
  ssoSecret: process.env.SSO_SECRET || 'auth-center-sso-secret-key-change-me',
  ssoTokenExpiresInSeconds: 300,
  
  loginMaxFailedAttempts: 5,
  loginFreezeDurationMinutes: 15,
  
  mfaMaxFailedAttempts: 5,
  mfaFreezeDurationMinutes: 10,
  
  totpWindow: 1,
  sessionExpiresInSeconds: 86400,
};

export type UserRow = {
  id: number;
  username: string;
  password_hash: string;
  mfa_enabled: number;
  mfa_secret: string | null;
  mfa_failed_attempts: number;
  mfa_frozen_until: number;
  login_failed_attempts: number;
  login_frozen_until: number;
  created_at: number;
};

export type AppRow = {
  id: string;
  name: string;
  callback_url: string;
  secret: string;
  created_at: number;
};

export type SessionRow = {
  id: string;
  user_id: number;
  app_id: string;
  callback_url: string | null;
  token: string | null;
  created_at: number;
  expires_at: number;
};

export type AuthLogRow = {
  id: number;
  user_id: number | null;
  username: string | null;
  action: string;
  success: number;
  ip: string | null;
  user_agent: string | null;
  created_at: number;
};
