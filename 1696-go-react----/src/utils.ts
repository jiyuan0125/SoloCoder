import db from './database';

export class HttpError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

export const todayDate = (): string => {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, '0');
  const d = String(now.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
};

export const nowDateTime = (): string => {
  return new Date().toISOString().replace('T', ' ').substring(0, 19);
};

export const getStock = (
  warehouseId: number,
  name: string,
  specification: string,
  category: string
): number => {
  const row = db
    .prepare(
      `SELECT COALESCE(SUM(quantity), 0) AS total FROM material_batches
       WHERE warehouse_id = ? AND name = ? AND specification = ? AND category = ?`
    )
    .get(warehouseId, name, specification, category) as { total: number };
  return row.total;
};

export const getInTransitQuantity = (
  warehouseId: number,
  name: string,
  specification: string,
  category: string
): number => {
  const row = db
    .prepare(
      `SELECT COALESCE(SUM(quantity), 0) AS total FROM allocation_in_transit
       WHERE warehouse_id = ? AND material_name = ? AND specification = ? AND category = ?`
    )
    .get(warehouseId, name, specification, category) as { total: number };
  return row.total;
};

export const getAvailableStock = (
  warehouseId: number,
  name: string,
  specification: string,
  category: string
): number => {
  const stock = getStock(warehouseId, name, specification, category);
  const inTransit = getInTransitQuantity(warehouseId, name, specification, category);
  return stock - inTransit;
};

export const nextAllocationStatus = (
  current: string
): string | null => {
  const order = [
    'pending_approval',
    'approved',
    'in_transit',
    'arrived',
    'received',
  ];
  const idx = order.indexOf(current);
  if (idx === -1 || idx >= order.length - 1) {
    return null;
  }
  return order[idx + 1];
};

export const validateMaterialInput = (input: {
  name: string;
  quantity: number;
}): void => {
  if (!input.name || String(input.name).trim() === '') {
    throw new HttpError(400, '物资名称不能为空');
  }
  if (!Number.isFinite(input.quantity) || input.quantity <= 0) {
    throw new HttpError(400, '数量必须为大于零的数值');
  }
};
