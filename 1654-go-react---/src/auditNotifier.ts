import { AuditNotification } from './types';

export interface AuditNotifierConfig {
  endpoint?: string;
  timeoutMs?: number;
  enabled?: boolean;
}

const DEFAULT_CONFIG: Required<AuditNotifierConfig> = {
  endpoint: '',
  timeoutMs: 5000,
  enabled: true,
};

export class AuditNotifier {
  private config: Required<AuditNotifierConfig>;
  private notifications: AuditNotification[] = [];

  constructor(config?: AuditNotifierConfig) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  async sendNotification(notification: AuditNotification): Promise<void> {
    if (!this.config.enabled) {
      this.notifications.push(notification);
      return;
    }

    this.notifications.push(notification);

    if (!this.config.endpoint) {
      console.log(
        `[Audit] Notification recorded: action=${notification.action}, sessionId=${notification.sessionId}, userId=${notification.userId}`
      );
      return;
    }

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(
        () => controller.abort(),
        this.config.timeoutMs
      );

      const response = await fetch(this.config.endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(notification),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(
          `Audit notification failed with status ${response.status}`
        );
      }
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Unknown error';
      throw new Error(`Failed to send audit notification: ${message}`);
    }
  }

  getNotifications(): AuditNotification[] {
    return [...this.notifications];
  }

  clearNotifications(): void {
    this.notifications = [];
  }
}
