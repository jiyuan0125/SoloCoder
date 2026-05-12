import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Client } from '../types';

export function getOrCreateClient(name: string, phone: string): Client {
  let client = db.prepare('SELECT * FROM clients WHERE phone = ?').get(phone) as any;
  
  if (!client) {
    const id = uuidv4();
    db.prepare('INSERT INTO clients (id, name, phone) VALUES (?, ?, ?)').run(id, name, phone);
    client = db.prepare('SELECT * FROM clients WHERE id = ?').get(id) as any;
  } else {
    db.prepare('UPDATE clients SET name = ? WHERE phone = ?').run(name, phone);
    client = db.prepare('SELECT * FROM clients WHERE phone = ?').get(phone) as any;
  }

  return {
    ...client,
    no_show_count: Number(client.no_show_count)
  };
}

export function getClientByPhone(phone: string): Client | undefined {
  const row = db.prepare('SELECT * FROM clients WHERE phone = ?').get(phone) as any;
  if (!row) return undefined;
  return {
    ...row,
    no_show_count: Number(row.no_show_count)
  };
}

export function isClientBanned(clientId: string): boolean {
  const row = db.prepare('SELECT banned_until FROM clients WHERE id = ?').get(clientId) as any;
  if (!row || !row.banned_until) return false;
  return new Date(row.banned_until) > new Date();
}

export function incrementNoShow(clientId: string): number {
  const client = db.prepare('SELECT * FROM clients WHERE id = ?').get(clientId) as any;
  const newCount = Number(client.no_show_count) + 1;
  
  let bannedUntil: string | null = null;
  if (newCount >= 3) {
    const date = new Date();
    date.setDate(date.getDate() + 30);
    bannedUntil = date.toISOString();
  }

  db.prepare(`
    UPDATE clients 
    SET no_show_count = ?, banned_until = ?
    WHERE id = ?
  `).run(newCount, bannedUntil, clientId);

  return newCount;
}
