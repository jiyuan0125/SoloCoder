import db from '../db';
import { Rule } from '../types';
import { v4 as uuidv4 } from 'uuid';

export class RuleRepository {
  private static instance: RuleRepository;

  static getInstance(): RuleRepository {
    if (!RuleRepository.instance) {
      RuleRepository.instance = new RuleRepository();
    }
    return RuleRepository.instance;
  }

  create(rule: Omit<Rule, 'id' | 'createdAt' | 'updatedAt'>): Rule {
    const now = Date.now();
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO rules (id, name, condition, weight, priority, enabled, createdAt, updatedAt)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(id, rule.name, rule.condition, rule.weight, rule.priority, rule.enabled ? 1 : 0, now, now);
    return this.findById(id)!;
  }

  findById(id: string): Rule | null {
    const row = db.prepare('SELECT * FROM rules WHERE id = ?').get(id);
    return row ? this.mapRow(row) : null;
  }

  findAllEnabled(): Rule[] {
    const rows = db.prepare('SELECT * FROM rules WHERE enabled = 1 ORDER BY priority ASC').all();
    return rows.map(row => this.mapRow(row));
  }

  findAll(): Rule[] {
    const rows = db.prepare('SELECT * FROM rules ORDER BY priority ASC').all();
    return rows.map(row => this.mapRow(row));
  }

  update(id: string, updates: Partial<Omit<Rule, 'id' | 'createdAt'>>): Rule | null {
    const existing = this.findById(id);
    if (!existing) return null;

    const updatesWithDate = { ...updates, updatedAt: Date.now() };
    const fields: string[] = [];
    const values: any[] = [];

    for (const [key, value] of Object.entries(updatesWithDate)) {
      fields.push(`${key} = ?`);
      if (key === 'enabled') {
        values.push(value ? 1 : 0);
      } else {
        values.push(value);
      }
    }

    values.push(id);

    db.prepare(`UPDATE rules SET ${fields.join(', ')} WHERE id = ?`).run(...values);
    return this.findById(id);
  }

  delete(id: string): boolean {
    const result = db.prepare('DELETE FROM rules WHERE id = ?').run(id);
    return (result as any).changes > 0;
  }

  private mapRow(row: any): Rule {
    return {
      id: row.id,
      name: row.name,
      condition: row.condition,
      weight: row.weight,
      priority: row.priority,
      enabled: !!row.enabled,
      createdAt: row.createdAt,
      updatedAt: row.updatedAt
    };
  }
}
