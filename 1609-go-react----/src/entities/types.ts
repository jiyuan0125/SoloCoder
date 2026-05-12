export enum MessageCategory {
  IN_APP = "in_app",
  PUSH = "push",
  EMAIL = "email",
}

export enum InAppType {
  SYSTEM_NOTIFICATION = "system_notification",
  APPROVAL_NOTIFICATION = "approval_notification",
  TASK_REMINDER = "task_reminder",
}

export enum MessageStatus {
  PENDING = "pending",
  PROCESSED = "processed",
  EXPIRED = "expired",
}

export enum Language {
  ZH = "zh",
  EN = "en",
}

export const TODO_ACTIONS = ["approve", "reject"];
