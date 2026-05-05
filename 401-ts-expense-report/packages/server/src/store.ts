import type {
  ExpenseItem,
  ApprovalRecord,
  OperationLog,
  AuditLog,
  User,
} from '@expense-report/shared';

interface DataStore {
  expenses: Map<string, ExpenseItem>;
  approvalRecords: Map<string, ApprovalRecord[]>;
  operationLogs: OperationLog[];
  auditLogs: AuditLog[];
  users: Map<string, User>;
}

const store: DataStore = {
  expenses: new Map(),
  approvalRecords: new Map(),
  operationLogs: [],
  auditLogs: [],
  users: new Map(),
};

export function getStore(): DataStore {
  return store;
}

export function initDefaultUsers(): void {
  const defaultUsers: User[] = [
    { id: 'emp1', name: '张三', role: 'employee', department: '技术部' },
    { id: 'emp2', name: '李四', role: 'employee', department: '市场部' },
    { id: 'mgr1', name: '王经理', role: 'dept_manager', department: '技术部' },
    { id: 'mgr2', name: '刘经理', role: 'dept_manager', department: '市场部' },
    { id: 'fd1', name: '陈总监', role: 'finance_director', department: '财务部' },
    { id: 'admin1', name: '系统管理员', role: 'admin', department: '行政部' },
  ];

  for (const user of defaultUsers) {
    store.users.set(user.id, user);
  }
}

export function getUserById(userId: string): User | undefined {
  return store.users.get(userId);
}

export function getExpenseById(expenseId: string): ExpenseItem | undefined {
  return store.expenses.get(expenseId);
}

export function getAllExpenses(): ExpenseItem[] {
  return Array.from(store.expenses.values());
}

export function saveExpense(expense: ExpenseItem): void {
  store.expenses.set(expense.id, expense);
}

export function deleteExpense(expenseId: string): boolean {
  return store.expenses.delete(expenseId);
}

export function getApprovalRecords(expenseId: string): ApprovalRecord[] {
  return store.approvalRecords.get(expenseId) ?? [];
}

export function addApprovalRecord(record: ApprovalRecord): void {
  const records = store.approvalRecords.get(record.expenseId) ?? [];
  records.push(record);
  store.approvalRecords.set(record.expenseId, records);
}

export function clearApprovalRecords(expenseId: string): void {
  store.approvalRecords.delete(expenseId);
}

export function addOperationLog(log: OperationLog): void {
  store.operationLogs.push(log);
}

export function getAllOperationLogs(): OperationLog[] {
  return [...store.operationLogs];
}

export function addAuditLog(log: AuditLog): void {
  store.auditLogs.push(log);
}

export function getAllAuditLogs(): AuditLog[] {
  return [...store.auditLogs];
}
