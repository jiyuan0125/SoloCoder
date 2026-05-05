import type {
  ExpenseCategory,
  ExpenseStatus,
  ExpenseItem,
  OperationLog,
  AuditLog,
  User,
} from './types';

export interface CreateExpenseRequest {
  employeeId: string;
  employeeName: string;
  date: string;
  amount: number;
  category: ExpenseCategory;
  reason: string;
  voucherFileName: string;
}

export interface CreateExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface UpdateExpenseRequest {
  expenseId: string;
  date?: string;
  amount?: number;
  category?: ExpenseCategory;
  reason?: string;
  voucherFileName?: string;
  supplementaryInfo?: string;
}

export interface UpdateExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface SubmitExpenseRequest {
  expenseId: string;
}

export interface SubmitExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface ApproveExpenseRequest {
  expenseId: string;
  approverId: string;
  approverName: string;
  comment?: string;
}

export interface ApproveExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
  needsSecondLevel?: boolean;
}

export interface RejectExpenseRequest {
  expenseId: string;
  approverId: string;
  approverName: string;
  comment?: string;
}

export interface RejectExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface BatchApproveRequest {
  expenseIds: string[];
  approverId: string;
  approverName: string;
  comment?: string;
}

export interface BatchApproveResponse {
  success: boolean;
  approved: string[];
  skipped: string[];
  failed: Array<{
    expenseId: string;
    code: number;
    message: string;
  }>;
}

export interface PayExpenseRequest {
  expenseId: string;
  operatorId: string;
  operatorName: string;
}

export interface PayExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface GetExpenseRequest {
  expenseId: string;
}

export interface GetExpenseResponse {
  success: boolean;
  data?: ExpenseItem;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface ListExpensesRequest {
  startDate?: string;
  endDate?: string;
  status?: ExpenseStatus;
  employeeId?: string;
  page?: number;
  pageSize?: number;
}

export interface ListExpensesResponse {
  success: boolean;
  data: {
    items: ExpenseItem[];
    total: number;
    page: number;
    pageSize: number;
  };
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface DeleteExpenseRequest {
  expenseId: string;
  operatorId: string;
  operatorName: string;
}

export interface DeleteExpenseResponse {
  success: boolean;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface GetOperationLogsRequest {
  startDate?: string;
  endDate?: string;
  userId?: string;
  page?: number;
  pageSize?: number;
}

export interface GetOperationLogsResponse {
  success: boolean;
  data: {
    items: OperationLog[];
    total: number;
    page: number;
    pageSize: number;
  };
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface GetAuditLogsRequest {
  startDate?: string;
  endDate?: string;
  userId?: string;
  action?: string;
  page?: number;
  pageSize?: number;
}

export interface GetAuditLogsResponse {
  success: boolean;
  data: {
    items: AuditLog[];
    total: number;
    page: number;
    pageSize: number;
  };
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface UserLoginRequest {
  userId: string;
}

export interface UserLoginResponse {
  success: boolean;
  data?: User;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}
