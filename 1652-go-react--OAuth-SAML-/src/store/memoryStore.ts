import { Client, AuthorizationCode, Token } from '../types';

class MemoryStore {
  private clients: Map<string, Client> = new Map();
  private clientsByName: Map<string, Client> = new Map();
  private authorizationCodes: Map<string, AuthorizationCode> = new Map();
  private tokens: Map<string, Token> = new Map();
  private tokensByRefreshToken: Map<string, Token> = new Map();

  addClient(client: Client): void {
    this.clients.set(client.clientId, client);
    this.clientsByName.set(client.name, client);
  }

  getClientByClientId(clientId: string): Client | undefined {
    return this.clients.get(clientId);
  }

  getClientByName(name: string): Client | undefined {
    return this.clientsByName.get(name);
  }

  addAuthorizationCode(code: AuthorizationCode): void {
    this.authorizationCodes.set(code.code, code);
  }

  getAuthorizationCode(code: string): AuthorizationCode | undefined {
    return this.authorizationCodes.get(code);
  }

  updateAuthorizationCode(code: AuthorizationCode): void {
    this.authorizationCodes.set(code.code, code);
  }

  addToken(token: Token): void {
    this.tokens.set(token.accessToken, token);
    this.tokensByRefreshToken.set(token.refreshToken, token);
  }

  getTokenByAccessToken(accessToken: string): Token | undefined {
    return this.tokens.get(accessToken);
  }

  getTokenByRefreshToken(refreshToken: string): Token | undefined {
    return this.tokensByRefreshToken.get(refreshToken);
  }

  updateToken(token: Token): void {
    this.tokens.set(token.accessToken, token);
    this.tokensByRefreshToken.set(token.refreshToken, token);
  }

  revokeTokensByAuthorizationCode(code: string): void {
    for (const token of this.tokens.values()) {
      if (token.authorizationCode === code) {
        token.accessTokenRevoked = true;
        token.refreshTokenRevoked = true;
        this.tokens.set(token.accessToken, token);
        this.tokensByRefreshToken.set(token.refreshToken, token);
      }
    }
  }

  revokeTokensByRefreshToken(refreshToken: string): void {
    const originalToken = this.tokensByRefreshToken.get(refreshToken);
    if (!originalToken) return;
    for (const token of this.tokens.values()) {
      if (
        token.userId === originalToken.userId && token.clientId === originalToken.clientId
      ) {
        token.accessTokenRevoked = true;
        token.refreshTokenRevoked = true;
        this.tokens.set(token.accessToken, token);
        this.tokensByRefreshToken.set(token.refreshToken, token);
      }
    }
  }

  revokeTokensByAccessToken(accessToken: string): void {
    const originalToken = this.tokens.get(accessToken);
    if (originalToken) {
      originalToken.accessTokenRevoked = true;
      originalToken.refreshTokenRevoked = true;
      this.tokens.set(accessToken, originalToken);
      this.tokensByRefreshToken.set(originalToken.refreshToken, originalToken);
    }
  }

  revokeToken(accessToken: string): void {
    const token = this.tokens.get(accessToken);
    if (token) {
      token.accessTokenRevoked = true;
      token.refreshTokenRevoked = true;
    }
  }
}

export const store = new MemoryStore();
