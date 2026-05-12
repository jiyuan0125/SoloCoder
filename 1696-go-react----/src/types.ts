export type WarehouseType = 'central' | 'area';

export interface Warehouse {
  id: number;
  name: string;
  type: WarehouseType;
  created_at: string;
}

export interface MaterialBatch {
  id: number;
  warehouse_id: number;
  name: string;
  specification: string;
  category: string;
  quantity: number;
  inbound_date: string;
  expiry_date: string | null;
  supplier: string;
  created_at: string;
}

export interface AllocationStatus {
  PENDING_APPROVAL: 'pending_approval';
  APPROVED: 'approved';
  IN_TRANSIT: 'in_transit';
  ARRIVED: 'arrived';
  RECEIVED: 'received';
}

export const ALLOCATION_STATUSES: AllocationStatus[keyof AllocationStatus][] = [
  'pending_approval',
  'approved',
  'in_transit',
  'arrived',
  'received',
];

export interface Allocation {
  id: number;
  from_warehouse_id: number;
  to_warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  requested_quantity: number;
  approved_quantity: number | null;
  status: AllocationStatus[keyof AllocationStatus];
  supplier: string;
  created_at: string;
}

export interface AllocationInTransit {
  id: number;
  allocation_id: number;
  warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  quantity: number;
  supplier: string;
  inbound_date: string;
  expiry_date: string | null;
}

export interface IssueRecord {
  id: number;
  warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  quantity: number;
  issued_at: string;
  voided: boolean;
  voided_at: string | null;
}

export interface IssueLine {
  id: number;
  issue_id: number;
  batch_id: number;
  quantity: number;
}

export interface DailyLedger {
  id: number;
  ledger_date: string;
  warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  opening_quantity: number;
  inbound_quantity: number;
  outbound_quantity: number;
  closing_quantity: number;
  created_at: string;
}

export interface InventoryDifference {
  id: number;
  warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  expected_quantity: number;
  actual_quantity: number;
  difference_quantity: number;
  recorded_at: string;
}

export interface CheckTask {
  id: number;
  warehouse_id: number;
  material_name: string;
  specification: string;
  category: string;
  expected_quantity: number;
  actual_quantity: number;
  difference_quantity: number;
  created_at: string;
  resolved: boolean;
}

export interface DailyIssueSummary {
  category: string;
  material_name: string;
  specification: string;
  total_quantity: number;
}
