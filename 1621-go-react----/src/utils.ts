import crypto from 'crypto';
import fs from 'fs';

const isValidCron = require('cron-validator').isValidCron;

export function validateCronExpression(cron: string): { valid: boolean; reason?: string } {
  try {
    const valid = isValidCron(cron);
    if (valid) {
      return { valid: true };
    }
    return { valid: false, reason: 'Invalid cron expression format' };
  } catch (error) {
    return { 
      valid: false, 
      reason: error instanceof Error ? error.message : 'Unknown cron validation error' 
    };
  }
}

export function generateId(): string {
  return crypto.randomUUID();
}

export async function calculateSHA256(filePath: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash('sha256');
    const stream = fs.createReadStream(filePath);
    
    stream.on('data', (data) => hash.update(data));
    stream.on('end', () => resolve(hash.digest('hex')));
    stream.on('error', (err) => reject(err));
  });
}

export function getCurrentTime(): string {
  return new Date().toISOString();
}
