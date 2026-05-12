export enum CategoryType {
  FEATURE_SUGGESTION = '功能建议',
  BUG_REPORT = 'Bug报告',
  EXPERIENCE_COMPLAINT = '体验投诉',
  OTHER = '其他'
}

export enum Priority {
  URGENT = '紧急',
  NORMAL = '普通',
  LOW = '低'
}

export enum FeedbackStatus {
  NEW = '新建',
  PENDING = '待处理',
  PROCESSING = '处理中',
  RESOLVED = '已解决',
  CLOSED = '已关闭'
}

export interface Category {
  id: number;
  name: CategoryType;
  createdAt: string;
}

export interface Feedback {
  id: number;
  userId: number;
  title: string;
  description: string;
  categoryId: number;
  priority: Priority;
  status: FeedbackStatus;
  assigneeId: number | null;
  internalNote: string;
  isTimeout: boolean;
  resolvedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface StatusLog {
  id: number;
  feedbackId: number;
  fromStatus: FeedbackStatus;
  toStatus: FeedbackStatus;
  operatorId: number;
  operatorRole: 'admin' | 'user';
  createdAt: string;
}

export interface Attachment {
  id: number;
  feedbackId: number;
  fileName: string;
  filePath: string;
  fileSize: number;
  createdAt: string;
}

export interface CreateFeedbackRequest {
  userId: number;
  title: string;
  description: string;
  categoryId: number;
}

export interface UpdatePriorityRequest {
  priority: Priority;
}

export interface UpdateStatusRequest {
  toStatus: FeedbackStatus;
  operatorId: number;
  operatorRole: 'admin' | 'user';
}

export interface BatchClaimRequest {
  feedbackIds: number[];
  assigneeId: number;
}

export interface BatchResolveRequest {
  feedbackIds: number[];
  operatorId: number;
}

export interface NotificationResult {
  success: boolean;
  message: string;
}
