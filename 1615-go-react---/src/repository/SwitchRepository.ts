import { getDatabase } from '../database/db';
import type { FeatureSwitch, SwitchHistory, ServiceReference, SwitchConditions, ValueType } from '../types';

interface SwitchRow {
  id: number;
  key: string;
  name: string;
  description: string;
  value_type: string;
  boolean_value: number;
  json_value: string | null;
  conditions: string;
  is_deleted: number;
  created_by: string;
  created_at: number;
  updated_by: string;
  updated_at: number;
  version: number;
}

interface HistoryRow {
  id: number;
  switch_key: string;
  operation: string;
  old_value: string | null;
  new_value: string | null;
  operator: string;
  timestamp: number;
  version: number;
}

interface RefRow {
  id: number;
  service_name: string;
  switch_key: string;
  is_active: number;
  created_at: number;
  last_poll_at: number;
}

function parseConditions(conditionsStr: string): SwitchConditions {
  try {
    return JSON.parse(conditionsStr) as SwitchConditions;
  } catch {
    return {};
  }
}

function mapSwitchRow(row: SwitchRow): FeatureSwitch {
  return {
    id: row.id,
    key: row.key,
    name: row.name,
    description: row.description,
    valueType: row.value_type as ValueType,
    booleanValue: row.boolean_value === 1,
    jsonValue: row.json_value,
    conditions: parseConditions(row.conditions),
    isDeleted: row.is_deleted === 1,
    createdBy: row.created_by,
    createdAt: row.created_at,
    updatedBy: row.updated_by,
    updatedAt: row.updated_at,
    version: row.version
  };
}

function mapHistoryRow(row: HistoryRow): SwitchHistory {
  return {
    id: row.id,
    switchKey: row.switch_key,
    operation: row.operation as 'create' | 'update' | 'delete',
    oldValue: row.old_value,
    newValue: row.new_value,
    operator: row.operator,
    timestamp: row.timestamp,
    version: row.version
  };
}

function mapRefRow(row: RefRow): ServiceReference {
  return {
    id: row.id,
    serviceName: row.service_name,
    switchKey: row.switch_key,
    isActive: row.is_active === 1,
    createdAt: row.created_at,
    lastPollAt: row.last_poll_at
  };
}

export class SwitchRepository {
  private db = getDatabase();

  findByKey(key: string): FeatureSwitch | null {
    const row = this.db.prepare<{ key: string }, SwitchRow>(`
      SELECT * FROM switches WHERE key = @key
    `).get({ key });
    return row ? mapSwitchRow(row) : null;
  }

  findAll(): FeatureSwitch[] {
    const stmt = this.db.prepare(`SELECT * FROM switches ORDER BY updated_at DESC`);
    const rows = stmt.all() as SwitchRow[];
    return rows.map(mapSwitchRow);
  }

  findByKeys(keys: string[]): FeatureSwitch[] {
    if (keys.length === 0) return [];
    const placeholders = keys.map(() => '?').join(',');
    const rows = this.db.prepare<string[], SwitchRow>(`
      SELECT * FROM switches WHERE key IN (${placeholders})
    `).all(...keys);
    return rows.map(mapSwitchRow);
  }

  findByVersionAfter(version: number): { updated: FeatureSwitch[]; deleted: FeatureSwitch[] } {
    const all = this.db.prepare<{ version: number }, SwitchRow>(`
      SELECT * FROM switches WHERE version > @version
    `).all({ version });
    const updated = all.filter(s => !s.is_deleted);
    const deleted = all.filter(s => s.is_deleted);
    return {
      updated: updated.map(mapSwitchRow),
      deleted: deleted.map(mapSwitchRow)
    };
  }

  getCurrentVersion(): number {
    const stmt = this.db.prepare(`SELECT last_version FROM sync_state WHERE id = 1`);
    const row = stmt.get() as { last_version: number } | undefined;
    return row?.last_version ?? 0;
  }

  create(params: {
    key: string;
    name: string;
    description: string;
    valueType: ValueType;
    booleanValue: boolean;
    jsonValue: string | null;
    conditions: string;
    operator: string;
  }): FeatureSwitch {
    const now = Date.now();
    const tx = this.db.transaction(() => {
      const version = this.getCurrentVersion() + 1;
      const result = this.db.prepare<{
        key: string; name: string; description: string; valueType: string;
        booleanValue: number; jsonValue: string | null; conditions: string;
        operator: string; createdAt: number; updatedAt: number; version: number;
      }, SwitchRow>(`
        INSERT INTO switches
          (key, name, description, value_type, boolean_value, json_value, conditions,
           created_by, created_at, updated_by, updated_at, version)
        VALUES
          (@key, @name, @description, @valueType, @booleanValue, @jsonValue, @conditions,
           @operator, @createdAt, @operator, @updatedAt, @version)
        RETURNING *
      `).get({
        key: params.key,
        name: params.name,
        description: params.description,
        valueType: params.valueType,
        booleanValue: params.booleanValue ? 1 : 0,
        jsonValue: params.jsonValue,
        conditions: params.conditions,
        operator: params.operator,
        createdAt: now,
        updatedAt: now,
        version
      });

      this.db.prepare(`
        INSERT INTO switch_history
          (switch_key, operation, old_value, new_value, operator, timestamp, version)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `).run(params.key, 'create', null, params.jsonValue ?? String(params.booleanValue), params.operator, now, version);

      this.db.prepare(`UPDATE sync_state SET last_version = ?, updated_at = ? WHERE id = 1`).run(version, now);

      return result!;
    });
    const row = tx();
    return mapSwitchRow(row);
  }

  update(key: string, updates: {
    name?: string;
    description?: string;
    booleanValue?: boolean;
    jsonValue?: string | null;
    conditions?: string;
    operator: string;
  }): FeatureSwitch | null {
    const existing = this.findByKey(key);
    if (!existing) return null;

    const tx = this.db.transaction(() => {
      const version = this.getCurrentVersion() + 1;
      const now = Date.now();

      const fields: string[] = ['updated_by = ?', 'updated_at = ?', 'version = ?'];
      const values: (string | number | boolean | null)[] = [updates.operator, now, version];

      if (updates.name !== undefined) {
        fields.push('name = ?');
        values.push(updates.name);
      }
      if (updates.description !== undefined) {
        fields.push('description = ?');
        values.push(updates.description);
      }
      if (updates.booleanValue !== undefined) {
        fields.push('boolean_value = ?');
        values.push(updates.booleanValue ? 1 : 0);
      }
      if (updates.jsonValue !== undefined) {
        fields.push('json_value = ?');
        values.push(updates.jsonValue);
      }
      if (updates.conditions !== undefined) {
        fields.push('conditions = ?');
        values.push(updates.conditions);
      }

      values.push(key);

      const query = `UPDATE switches SET ${fields.join(', ')} WHERE key = ? RETURNING *`;
      const result = this.db.prepare(query).get(...values) as SwitchRow;

      const oldValue = existing.valueType === 'json' ? existing.jsonValue : String(existing.booleanValue);
      const newBool = updates.booleanValue !== undefined ? updates.booleanValue : existing.booleanValue;
      const newJson = updates.jsonValue !== undefined ? updates.jsonValue : existing.jsonValue;
      const newValue = existing.valueType === 'json' ? newJson : String(newBool);

      this.db.prepare(`
        INSERT INTO switch_history
          (switch_key, operation, old_value, new_value, operator, timestamp, version)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `).run(key, 'update', oldValue, newValue, updates.operator, now, version);

      this.db.prepare(`UPDATE sync_state SET last_version = ?, updated_at = ? WHERE id = 1`).run(version, now);

      return result;
    });

    const row = tx();
    return row ? mapSwitchRow(row) : null;
  }

  softDelete(key: string, operator: string): FeatureSwitch | null {
    const existing = this.findByKey(key);
    if (!existing) return null;
    if (existing.isDeleted) return existing;

    const tx = this.db.transaction(() => {
      const version = this.getCurrentVersion() + 1;
      const now = Date.now();

      const result = this.db.prepare<{ operator: string; updatedAt: number; version: number; key: string }, SwitchRow>(`
        UPDATE switches
        SET is_deleted = 1, updated_by = @operator, updated_at = @updatedAt, version = @version
        WHERE key = @key
        RETURNING *
      `).get({ operator, updatedAt: now, version, key });

      const oldValue = existing.valueType === 'json' ? existing.jsonValue : String(existing.booleanValue);

      this.db.prepare(`
        INSERT INTO switch_history
          (switch_key, operation, old_value, new_value, operator, timestamp, version)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `).run(key, 'delete', oldValue, null, operator, now, version);

      this.db.prepare(`UPDATE sync_state SET last_version = ?, updated_at = ? WHERE id = 1`).run(version, now);

      return result;
    });

    const row = tx();
    return row ? mapSwitchRow(row) : null;
  }

  hasActiveReferences(key: string): boolean {
    const row = this.db.prepare<{ key: string }, { count: number }>(`
      SELECT COUNT(*) as count FROM service_references
      WHERE switch_key = @key AND is_active = 1
    `).get({ key });
    return (row?.count ?? 0) > 0;
  }

  upsertServiceReference(serviceName: string, switchKey: string): ServiceReference {
    const now = Date.now();
    const existing = this.db.prepare<{ serviceName: string; switchKey: string }, RefRow>(`
      SELECT * FROM service_references WHERE service_name = @serviceName AND switch_key = @switchKey
    `).get({ serviceName, switchKey });

    if (existing) {
      this.db.prepare(`UPDATE service_references SET is_active = 1, last_poll_at = ? WHERE id = ?`).run(now, existing.id);
      return this.getServiceReferenceById(existing.id)!;
    }

    const result = this.db.prepare<{
      serviceName: string; switchKey: string; isActive: number; createdAt: number; lastPollAt: number;
    }, RefRow>(`
      INSERT INTO service_references (service_name, switch_key, is_active, created_at, last_poll_at)
      VALUES (@serviceName, @switchKey, @isActive, @createdAt, @lastPollAt)
      RETURNING *
    `).get({ serviceName, switchKey, isActive: 1, createdAt: now, lastPollAt: now });

    return mapRefRow(result!);
  }

  getServiceReferenceById(id: number): ServiceReference | null {
    const row = this.db.prepare<{ id: number }, RefRow>(`
      SELECT * FROM service_references WHERE id = @id
    `).get({ id });
    return row ? mapRefRow(row) : null;
  }

  getHistory(key: string, limit: number = 50): SwitchHistory[] {
    const rows = this.db.prepare<{ key: string; limit: number }, HistoryRow>(`
      SELECT * FROM switch_history
      WHERE switch_key = @key
      ORDER BY timestamp DESC
      LIMIT @limit
    `).all({ key, limit });
    return rows.map(mapHistoryRow);
  }
}
