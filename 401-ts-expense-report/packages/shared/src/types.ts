export type ExpenseCategory = 'transport' | 'dining' | 'accommodation' | 'office' | 'other';

export type ExpenseStatus = 'pending' | 'approved' | 'paid' | 'rejected';

export type ApprovalLevel = 'dept_manager' | 'finance_director';

export type UserRole = 'employee' | 'dept_manager' | 'finance_director' | 'admin';

export type AuditActionType = 
  | 'create_expense'
  | 'update_expense'
  | 'delete_expense'
  | 'submit_expense'
  | 'approve_expense'
  | 'reject_expense'
  | 'batch_approve'
  | 'pay_expense'
  | 'update_amount'
  | 'delete_audit_log';

export interface User {
  id: string;
  name: string;
  role: UserRole;
  department?: string | undefined;
}

export interface ExpenseItem {
  id: string;
  employeeId: string;
  employeeName: string;
  date: string;
  amount: number;
  category: ExpenseCategory;
  reason: string;
  voucherFileName: string;
  status: ExpenseStatus;
  needsSupplementaryInfo: boolean;
  supplementaryInfo?: string | undefined;
  approvalLevel?: ApprovalLevel | undefined;
  createdAt: string;
  updatedAt: string;
}

export interface ApprovalRecord {
  id: string;
  expenseId: string;
  approverId: string;
  approverName: string;
  approvalLevel: ApprovalLevel;
  action: 'approve' | 'reject';
  comment?: string | undefined;
  createdAt: string;
}

export interface OperationLog {
  id: string;
  userId: string;
  userName: string;
  action: string;
  targetType: 'expense' | 'user' | 'system';
  targetId?: string | undefined;
  details?: Record<string, unknown> | undefined;
  createdAt: string;
}

export interface AuditLog {
  id: string;
  userId: string;
  userName: string;
  action: AuditActionType;
  targetType: 'expense' | 'audit_log' | 'system';
  targetId?: string | undefined;
  oldValue?: Record<string, unknown> | undefined;
  newValue?: Record<string, unknown> | undefined;
  ipAddress?: string | undefined;
  createdAt: string;
}
