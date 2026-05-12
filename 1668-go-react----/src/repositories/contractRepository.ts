import { db } from '../database';
import { Contract, ContractStatus } from '../types';

export interface ContractRow {
  id: string;
  code: string;
  name: string;
  status: string;
  created_at: string;
}

export function mapContractRow(row: ContractRow): Contract {
  return {
    id: row.id,
    code: row.code,
    name: row.name,
    status: row.status as ContractStatus,
    createdAt: row.created_at
  };
}

export function findContractById(id: string): Contract | undefined {
  const stmt = db.prepare(`SELECT * FROM contracts WHERE id = ?`);
  const row = stmt.get(id) as ContractRow | undefined;
  return row ? mapContractRow(row) : undefined;
}

export function insertContract(contract: Contract): void {
  const stmt = db.prepare(`
    INSERT INTO contracts (id, code, name, status, created_at)
    VALUES (?, ?, ?, ?, ?)
  `);
  stmt.run(
    contract.id,
    contract.code,
    contract.name,
    contract.status,
    contract.createdAt
  );
}

export function getAllContracts(): Contract[] {
  const stmt = db.prepare(`SELECT * FROM contracts ORDER BY created_at DESC`);
  const rows = stmt.all() as ContractRow[];
  return rows.map(mapContractRow);
}
