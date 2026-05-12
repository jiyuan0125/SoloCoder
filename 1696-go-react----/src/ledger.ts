import db from './database';
import { todayDate } from './utils';

export const updateLedgerForInbound = (
  warehouseId: number,
  name: string,
  specification: string,
  category: string,
  quantity: number
): void => {
  const date = todayDate();
  const key = {
    ledger_date: date,
    warehouse_id: warehouseId,
    material_name: name,
    specification,
    category,
  };

  const existing = db
    .prepare(
      `SELECT id, opening_quantity, inbound_quantity, closing_quantity
       FROM daily_ledger
       WHERE ledger_date = ? AND warehouse_id = ? AND material_name = ?
         AND specification = ? AND category = ?`
    )
    .get(date, warehouseId, name, specification, category) as
    | {
        id: number;
        opening_quantity: number;
        inbound_quantity: number;
        closing_quantity: number;
      }
    | undefined;

  if (existing) {
    db.prepare(
      `UPDATE daily_ledger
       SET inbound_quantity = inbound_quantity + ?,
           closing_quantity = opening_quantity + inbound_quantity + ? - outbound_quantity
       WHERE id = ?`
    ).run(quantity, quantity, existing.id);
  } else {
    const previous = db
      .prepare(
        `SELECT closing_quantity FROM daily_ledger
         WHERE warehouse_id = ? AND material_name = ? AND specification = ? AND category = ?
           AND ledger_date < ?
         ORDER BY ledger_date DESC
         LIMIT 1`
      )
      .get(warehouseId, name, specification, category, date) as
      | { closing_quantity: number }
      | undefined;

    const opening = previous ? previous.closing_quantity : 0;
    db.prepare(
      `INSERT INTO daily_ledger
       (ledger_date, warehouse_id, material_name, specification, category,
        opening_quantity, inbound_quantity, outbound_quantity, closing_quantity)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(date, warehouseId, name, specification, category, opening, quantity, 0, opening + quantity);
  }
};

export const updateLedgerForOutbound = (
  warehouseId: number,
  name: string,
  specification: string,
  category: string,
  quantity: number
): void => {
  const date = todayDate();
  const existing = db
    .prepare(
      `SELECT id, opening_quantity, inbound_quantity, closing_quantity, outbound_quantity
       FROM daily_ledger
       WHERE ledger_date = ? AND warehouse_id = ? AND material_name = ?
         AND specification = ? AND category = ?`
    )
    .get(date, warehouseId, name, specification, category) as
    | {
        id: number;
        opening_quantity: number;
        inbound_quantity: number;
        outbound_quantity: number;
        closing_quantity: number;
      }
    | undefined;

  if (existing) {
    db.prepare(
      `UPDATE daily_ledger
       SET outbound_quantity = outbound_quantity + ?,
           closing_quantity = opening_quantity + inbound_quantity - (outbound_quantity + ?)
       WHERE id = ?`
    ).run(quantity, quantity, existing.id);
  } else {
    const previous = db
      .prepare(
        `SELECT closing_quantity FROM daily_ledger
         WHERE warehouse_id = ? AND material_name = ? AND specification = ? AND category = ?
           AND ledger_date < ?
         ORDER BY ledger_date DESC
         LIMIT 1`
      )
      .get(warehouseId, name, specification, category, date) as
      | { closing_quantity: number }
      | undefined;

    const opening = previous ? previous.closing_quantity : 0;
    db.prepare(
      `INSERT INTO daily_ledger
       (ledger_date, warehouse_id, material_name, specification, category,
        opening_quantity, inbound_quantity, outbound_quantity, closing_quantity)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(date, warehouseId, name, specification, category, opening, 0, quantity, opening - quantity);
  }
};
