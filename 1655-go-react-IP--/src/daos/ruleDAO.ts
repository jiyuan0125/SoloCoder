import db from '../database';
import { Rule, RuleType, RuleCondition } from '../types';

interface RuleRow {
  id: string;
  type: RuleType;
  ip_pattern: string;
  created_at: number;
  expires_at: number | null;
  enabled: number;
}

interface ConditionRow {
  id: string;
  rule_id: string;
  key: string;
  value: string;
  operator: 'equals' | 'contains' | 'regex';
}

function mapRule(row: RuleRow): Rule {
  return {
    id: row.id,
    type: row.type,
    ipPattern: row.ip_pattern,
    createdAt: row.created_at,
    expiresAt: row.expires_at ?? undefined,
    enabled: row.enabled === 1
  };
}

function mapCondition(row: ConditionRow): RuleCondition {
  return {
    id: row.id,
    ruleId: row.rule_id,
    key: row.key,
    value: row.value,
    operator: row.operator
  };
}

export const ruleDAO = {
  getAll(): Rule[] {
    const rows = db.prepare('SELECT * FROM rules WHERE enabled = 1').all() as RuleRow[];
    return rows.map(mapRule);
  },

  getById(id: string): Rule | null {
    const row = db.prepare('SELECT * FROM rules WHERE id = ?').get(id) as RuleRow | undefined;
    return row ? mapRule(row) : null;
  },

  getByType(type: RuleType): Rule[] {
    const rows = db.prepare(`
      SELECT * FROM rules 
      WHERE type = ? AND enabled = 1
    `).all(type) as RuleRow[];
    return rows.map(mapRule);
  },

  create(rule: Omit<Rule, 'createdAt'>): Rule {
    const now = Date.now();
    db.prepare(`
      INSERT INTO rules (id, type, ip_pattern, created_at, expires_at, enabled)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(
      rule.id,
      rule.type,
      rule.ipPattern,
      now,
      rule.expiresAt ?? null,
      rule.enabled ? 1 : 0
    );
    return { ...rule, createdAt: now };
  },

  update(id: string, updates: Partial<Omit<Rule, 'id' | 'createdAt'>>): Rule | null {
    const existing = this.getById(id);
    if (!existing) return null;

    const fields: string[] = [];
    const values: any[] = [];

    if (updates.type !== undefined) {
      fields.push('type = ?');
      values.push(updates.type);
    }
    if (updates.ipPattern !== undefined) {
      fields.push('ip_pattern = ?');
      values.push(updates.ipPattern);
    }
    if (updates.expiresAt !== undefined) {
      fields.push('expires_at = ?');
      values.push(updates.expiresAt ?? null);
    }
    if (updates.enabled !== undefined) {
      fields.push('enabled = ?');
      values.push(updates.enabled ? 1 : 0);
    }

    if (fields.length === 0) return existing;

    values.push(id);
    db.prepare(`UPDATE rules SET ${fields.join(', ')} WHERE id = ?`).run(...values);
    return this.getById(id);
  },

  delete(id: string): boolean {
    const result = db.prepare('DELETE FROM rules WHERE id = ?').run(id);
    return result.changes > 0;
  },

  findConflictingRule(ipPattern: string, excludeId?: string): Rule | null {
    let query = `
      SELECT r1.* FROM rules r1
      JOIN rules r2 ON r1.ip_pattern = r2.ip_pattern AND r1.type != r2.type
      WHERE r1.enabled = 1 AND r2.enabled = 1 AND r1.ip_pattern = ?
    `;
    const params: any[] = [ipPattern];

    if (excludeId) {
      query += ' AND r1.id != ? AND r2.id != ?';
      params.push(excludeId, excludeId);
    }

    query += ' LIMIT 1';
    const row = db.prepare(query).get(...params) as RuleRow | undefined;
    return row ? mapRule(row) : null;
  }
};

export const conditionDAO = {
  getByRuleId(ruleId: string): RuleCondition[] {
    const rows = db.prepare('SELECT * FROM rule_conditions WHERE rule_id = ?').all(ruleId) as ConditionRow[];
    return rows.map(mapCondition);
  },

  getById(id: string): RuleCondition | null {
    const row = db.prepare('SELECT * FROM rule_conditions WHERE id = ?').get(id) as ConditionRow | undefined;
    return row ? mapCondition(row) : null;
  },

  create(condition: RuleCondition): RuleCondition {
    db.prepare(`
      INSERT INTO rule_conditions (id, rule_id, key, value, operator)
      VALUES (?, ?, ?, ?, ?)
    `).run(
      condition.id,
      condition.ruleId,
      condition.key,
      condition.value,
      condition.operator
    );
    return condition;
  },

  update(id: string, updates: Partial<Omit<RuleCondition, 'id' | 'ruleId'>>): RuleCondition | null {
    const existing = this.getById(id);
    if (!existing) return null;

    const fields: string[] = [];
    const values: any[] = [];

    if (updates.key !== undefined) {
      fields.push('key = ?');
      values.push(updates.key);
    }
    if (updates.value !== undefined) {
      fields.push('value = ?');
      values.push(updates.value);
    }
    if (updates.operator !== undefined) {
      fields.push('operator = ?');
      values.push(updates.operator);
    }

    if (fields.length === 0) return existing;

    values.push(id);
    db.prepare(`UPDATE rule_conditions SET ${fields.join(', ')} WHERE id = ?`).run(...values);
    return this.getById(id);
  },

  delete(id: string): boolean {
    const result = db.prepare('DELETE FROM rule_conditions WHERE id = ?').run(id);
    return result.changes > 0;
  },

  deleteByRuleId(ruleId: string): boolean {
    const result = db.prepare('DELETE FROM rule_conditions WHERE rule_id = ?').run(ruleId);
    return result.changes > 0;
  }
};
