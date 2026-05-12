import { CategoryType, FeedbackStatus, Priority, NotificationResult } from './types';

export function validateTitle(title: string): boolean {
  return title.length >= 5 && title.length <= 100;
}

export function validateDescription(description: string): boolean {
  return description.length >= 20 && description.length <= 2000;
}

const URGENT_KEYWORDS = ['崩溃', '闪退', '白屏', '数据丢失'];

export function determinePriority(
  title: string,
  description: string,
  categoryName: CategoryType
): Priority {
  if (categoryName === CategoryType.FEATURE_SUGGESTION) {
    return Priority.LOW;
  }

  if (categoryName === CategoryType.BUG_REPORT) {
    const combinedText = `${title} ${description}`;
    const hasUrgentKeyword = URGENT_KEYWORDS.some(keyword => combinedText.includes(keyword));
    if (hasUrgentKeyword) {
      return Priority.URGENT;
    }
  }

  return Priority.NORMAL;
}

export const STATUS_TRANSITIONS: Record<FeedbackStatus, FeedbackStatus[]> = {
  [FeedbackStatus.NEW]: [FeedbackStatus.PENDING],
  [FeedbackStatus.PENDING]: [FeedbackStatus.PROCESSING],
  [FeedbackStatus.PROCESSING]: [FeedbackStatus.RESOLVED],
  [FeedbackStatus.RESOLVED]: [FeedbackStatus.CLOSED],
  [FeedbackStatus.CLOSED]: []
};

export function isValidStatusTransition(from: FeedbackStatus, to: FeedbackStatus): boolean {
  const allowedTransitions = STATUS_TRANSITIONS[from];
  return allowedTransitions.includes(to);
}

export function getAllowedTransitions(status: FeedbackStatus): FeedbackStatus[] {
  return STATUS_TRANSITIONS[status];
}

export function getPriorityOrder(priority: Priority): number {
  switch (priority) {
    case Priority.URGENT: return 0;
    case Priority.NORMAL: return 1;
    case Priority.LOW: return 2;
    default: return 3;
  }
}

export const USER_ACCOUNTS = new Map<number, { active: boolean }>([
  [1, { active: true }],
  [2, { active: true }],
  [3, { active: false }]
]);

export function isUserActive(userId: number): boolean {
  const user = USER_ACCOUNTS.get(userId);
  return user ? user.active : true;
}

export async function sendNotification(userId: number, message: string): Promise<NotificationResult> {
  if (!isUserActive(userId)) {
    return {
      success: false,
      message: '用户账号已注销，无法发送通知'
    };
  }

  console.log(`[Notification] 发送通知给用户 ${userId}: ${message}`);
  return {
    success: true,
    message: '通知发送成功'
  };
}

export function logInternalMessage(message: string): void {
  console.log(`[Internal Log] ${new Date().toISOString()} - ${message}`);
}

export function getDaysBetween(date1: string, date2: string): number {
  const d1 = new Date(date1);
  const d2 = new Date(date2);
  const diffTime = Math.abs(d2.getTime() - d1.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return diffDays;
}

export function getMinutesBetween(date1: string, date2: string): number {
  const d1 = new Date(date1);
  const d2 = new Date(date2);
  const diffTime = Math.abs(d2.getTime() - d1.getTime());
  const diffMinutes = Math.ceil(diffTime / (1000 * 60));
  return diffMinutes;
}
