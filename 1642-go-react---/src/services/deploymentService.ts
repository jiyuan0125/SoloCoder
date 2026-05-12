import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import type { Deployment } from '../types';

export function createDeployment(version: string, deployedAt?: string): Deployment {
  if (!version) {
    const err = new Error('缺少必填字段: version');
    (err as any).statusCode = 400;
    throw err;
  }

  const parsedTime = deployedAt ? new Date(deployedAt).getTime() : Date.now();
  if (isNaN(parsedTime)) {
    const err = new Error('无效的 deployedAt 时间格式');
    (err as any).statusCode = 400;
    throw err;
  }

  const id = uuidv4();
  db.prepare(`
    INSERT INTO deployments (id, version, deployed_at)
    VALUES (?, ?, ?)
  `).run(id, version, parsedTime);

  return {
    id,
    version,
    deployedAt: parsedTime,
  };
}

export function listDeployments(): Deployment[] {
  const rows = db.prepare(`
    SELECT * FROM deployments ORDER BY deployed_at DESC
  `).all() as any[];

  return rows.map(row => ({
    id: row.id,
    version: row.version,
    deployedAt: row.deployed_at,
  }));
}
