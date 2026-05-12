export interface Client {
  id: string;
  name: string;
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  scopes: string[];
  createdAt: number;
}

export interface AuthorizationCode {
  code: string;
  clientId: string;
  userId: string;
  scopes: string[];
  redirectUri: string;
  expiresAt: number;
  used: boolean;
  issuedAt: number;
}

export interface Token {
  accessToken: string;
  refreshToken: string;
  clientId: string;
  userId: string;
  scopes: string[];
  accessTokenExpiresAt: number;
  refreshTokenExpiresAt: number;
  accessTokenRevoked: boolean;
  refreshTokenRevoked: boolean;
  refreshTokenUsed: boolean;
  authorizationCode?: string;
  issuedAt: number;
}
