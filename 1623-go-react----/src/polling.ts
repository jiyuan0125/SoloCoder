import { Response } from 'express';
import { Environment } from './types';

interface PollingConnection {
  res: Response;
  timeout: NodeJS.Timeout;
  appId: number;
  environment: Environment;
  version: number;
}

const connections = new Map<string, PollingConnection>();

function getConnectionKey(appId: number, environment: Environment, version: number, timestamp: number): string {
  return `${appId}:${environment}:${version}:${timestamp}`;
}

export function addConnection(
  appId: number,
  environment: Environment,
  version: number,
  res: Response,
  timeoutMs: number = 30000
): string {
  const key = getConnectionKey(appId, environment, version, Date.now());
  const timeout = setTimeout(() => {
    removeConnection(key);
    if (!res.headersSent) {
      res.status(304).end();
    }
  }, timeoutMs);

  connections.set(key, {
    res,
    timeout,
    appId,
    environment,
    version
  });

  return key;
}

export function removeConnection(key: string): void {
  const connection = connections.get(key);
  if (connection) {
    clearTimeout(connection.timeout);
    connections.delete(key);
  }
}

export function notifyConfigChange(appId: number, environment: Environment, newVersion: number): void {
  const keysToRemove: string[] = [];

  connections.forEach((conn, key) => {
    if (conn.appId === appId && conn.environment === environment && newVersion > conn.version) {
      if (!conn.res.headersSent) {
        conn.res.status(200).end();
      }
      keysToRemove.push(key);
    }
  });

  keysToRemove.forEach(key => removeConnection(key));
}

export function disconnectByApp(appId: number): void {
  const keysToRemove: string[] = [];

  connections.forEach((conn, key) => {
    if (conn.appId === appId) {
      if (!conn.res.headersSent) {
        conn.res.status(410).json({ message: 'App has been deleted' });
      }
      keysToRemove.push(key);
    }
  });

  keysToRemove.forEach(key => removeConnection(key));
}

export function disconnectByConfig(appId: number, environment: Environment, key: string): void {
  const keysToRemove: string[] = [];

  connections.forEach((conn, connKey) => {
    if (conn.appId === appId && conn.environment === environment) {
      if (!conn.res.headersSent) {
        conn.res.status(410).json({ message: `Config '${key}' has been deleted` });
      }
      keysToRemove.push(connKey);
    }
  });

  keysToRemove.forEach(k => removeConnection(k));
}
