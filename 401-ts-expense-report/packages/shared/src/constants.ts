export const EXPENSE_CATEGORIES = ['transport', 'dining', 'accommodation', 'office', 'other'] as const;

export const EXPENSE_STATUSES = ['pending', 'approved', 'paid', 'rejected'] as const;

export const DINING_LIMIT_CENTS = 800 * 100;

export const ACCOMMODATION_NIGHT_LIMIT_CENTS = 500 * 100;

export const LARGE_AMOUNT_THRESHOLD_CENTS = 5000 * 100;

export const DEFAULT_PAGE = 1;
export const DEFAULT_PAGE_SIZE = 20;
export const MAX_PAGE_SIZE = 100;

export const SERVER_PORT = 3000;
export const SERVER_HOST = 'localhost';

export const CATEGORY_LABELS: Record<string, string> = {
  transport: '交通',
  dining: '餐饮',
  accommodation: '住宿',
  office: '办公',
  other: '其他',
};

export const STATUS_LABELS: Record<string, string> = {
  pending: '待审批',
  approved: '已批准',
  paid: '已打款',
  rejected: '已驳回',
};

export const ROLE_LABELS: Record<string, string> = {
  employee: '员工',
  dept_manager: '部门经理',
  finance_director: '财务总监',
  admin: '管理员',
};
