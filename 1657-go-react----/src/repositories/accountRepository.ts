import db from '../db';
import { Account } from '../types';
import { v4 as uuidv4 } from 'uuid';

export class AccountRepository {
  private static instance: AccountRepository;

  static getInstance(): AccountRepository {
    if (!AccountRepository.instance) {
      AccountRepository.instance = new AccountRepository();
    }
    return AccountRepository.instance;
  }

  findOrCreate(userId: string): Account {
    const existing = db.prepare('SELECT * FROM accounts WHERE userId = ?').get(userId);
    if (existing) return this.mapRow(existing);

    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO accounts (id, userId, frozen, frozenAt)
      VALUES (?, ?, 0, NULL)
    `);
    stmt.run(id, userId);
    return this.findById(id)!;
  }

  findById(id: string): Account | null {
    const row = db.prepare('SELECT * FROM accounts WHERE id = ?').get(id);
    return row ? this.mapRow(row) : null;
  }

  findByUserId(userId: string): Account | null {
    const row = db.prepare('SELECT * FROM accounts WHERE userId = ?').get(userId);
    return row ? this.mapRow(row) : null;
  }

  freeze(id: string): Account | null {
    const existing = this.findById(id);
    if (!existing) return null;

    db.prepare(`
      UPDATE accounts SET frozen = 1, frozenAt = ? WHERE id = ?
    `).run(Date.now(), id);
    return this.findById(id);
  }

  unfreeze(id: string): Account | null {
    const existing = this.findById(id);
    if (!existing) return null;

    db.prepare(`
      UPDATE accounts SET frozen = 0, frozenAt = NULL WHERE id = ?
    `).run(id);
    return this.findById(id);
  }

  private mapRow(row: any): Account {
    return {
      id: row.id,
      userId: row.userId,
      frozen: !!row.frozen,
      frozenAt: row.frozenAt
    };
  }
}
