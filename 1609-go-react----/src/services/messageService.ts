import { AppDataSource } from "../data-source";
import { Message } from "../entities/Message";
import { MessageTemplate } from "../entities/MessageTemplate";
import { ActionLog } from "../entities/ActionLog";
import { MessageStatus, TODO_ACTIONS, MessageCategory } from "../entities/types";
import { renderTemplate } from "../utils/template";
import { logPushNotification, logEmailNotification } from "../utils/logger";

export class MessageService {
  private messageRepository = AppDataSource.getRepository(Message);
  private templateRepository = AppDataSource.getRepository(MessageTemplate);
  private actionLogRepository = AppDataSource.getRepository(ActionLog);

  async sendMessage(data: {
    templateId: string;
    recipientId: string;
    variables: Record<string, string>;
    isPinned?: boolean;
    expireAt?: Date;
  }): Promise<Message | null> {
    const template = await this.templateRepository.findOne({
      where: { id: data.templateId },
      relations: ["type"],
    });

    if (!template) return null;

    const subject = renderTemplate(template.subject, data.variables);
    const content = renderTemplate(template.content, data.variables);

    const message = this.messageRepository.create({
      templateId: data.templateId,
      recipientId: data.recipientId,
      subject,
      content,
      variablesJson: JSON.stringify(data.variables),
      isPinned: data.isPinned || false,
      todoStatus: template.type.isTodo ? MessageStatus.PENDING : undefined,
      expireAt: data.expireAt,
    });

    const savedMessage = await this.messageRepository.save(message);

    if (template.type.category === MessageCategory.PUSH) {
      logPushNotification(savedMessage.id, savedMessage.recipientId, savedMessage.subject);
    } else if (template.type.category === MessageCategory.EMAIL) {
      logEmailNotification(savedMessage.id, savedMessage.recipientId, savedMessage.subject);
    }

    return savedMessage;
  }

  async getUnreadCount(recipientId: string): Promise<number> {
    return await this.messageRepository.count({
      where: {
        recipientId,
        isRead: false,
        isDeleted: false,
      },
    });
  }

  async getMessages(
    recipientId: string,
    filter: "all" | "unread" | "read" = "all"
  ): Promise<Message[]> {
    const where: any = {
      recipientId,
      isDeleted: false,
    };

    if (filter === "unread") {
      where.isRead = false;
    } else if (filter === "read") {
      where.isRead = true;
    }

    await this.checkAndMarkExpiredTodos(recipientId);

    const messages = await this.messageRepository.find({
      where,
      relations: ["template", "template.type"],
      order: { createdAt: "DESC" },
    });

    const pinned = messages.filter((m) => m.isPinned);
    const unpinned = messages.filter((m) => !m.isPinned);

    return [...pinned, ...unpinned];
  }

  async markAsRead(recipientId: string, messageIds: string[]): Promise<void> {
    const messages = await this.messageRepository.find({
      where: messageIds.map((id) => ({
        id,
        recipientId,
        isDeleted: false,
        isRead: false,
      })),
    });

    if (messages.length === 0) return;

    for (const message of messages) {
      message.isRead = true;
    }

    await this.messageRepository.save(messages);
  }

  async markAllAsRead(recipientId: string): Promise<void> {
    await this.messageRepository
      .createQueryBuilder()
      .update(Message)
      .set({ isRead: true })
      .where("recipientId = :recipientId AND isDeleted = :isDeleted AND isRead = :isRead", {
        recipientId,
        isDeleted: false,
        isRead: false,
      })
      .execute();
  }

  async deleteMessage(recipientId: string, messageId: string): Promise<boolean> {
    const message = await this.messageRepository.findOne({
      where: { id: messageId, recipientId, isDeleted: false },
    });

    if (!message) return false;

    message.isDeleted = true;
    await this.messageRepository.save(message);
    return true;
  }

  async executeAction(
    recipientId: string,
    messageId: string,
    action: string,
    operatorId: string
  ): Promise<{ success: boolean; errorCode?: number; errorMessage?: string }> {
    if (!TODO_ACTIONS.includes(action)) {
      return { success: false, errorCode: 400, errorMessage: "Invalid action" };
    }

    const message = await this.messageRepository.findOne({
      where: { id: messageId, recipientId, isDeleted: false },
      relations: ["template", "template.type"],
    });

    if (!message) {
      return { success: false, errorCode: 404, errorMessage: "Message not found" };
    }

    if (!message.template.type.isTodo) {
      return { success: false, errorCode: 400, errorMessage: "Not a todo message" };
    }

    if (message.todoStatus === MessageStatus.EXPIRED) {
      return { success: false, errorCode: 400, errorMessage: "该消息已过期" };
    }

    if (message.todoStatus !== MessageStatus.PENDING) {
      return { success: false, errorCode: 400, errorMessage: "Message already processed" };
    }

    message.todoStatus = MessageStatus.PROCESSED;
    message.processedBy = operatorId;
    message.processedAt = new Date();
    message.isRead = true;

    await this.messageRepository.save(message);

    const log = this.actionLogRepository.create({
      messageId,
      action,
      operatorId,
    });
    await this.actionLogRepository.save(log);

    return { success: true };
  }

  private async checkAndMarkExpiredTodos(recipientId: string): Promise<void> {
    const now = new Date();
    await this.messageRepository
      .createQueryBuilder()
      .update(Message)
      .set({ todoStatus: MessageStatus.EXPIRED })
      .where(
        "recipientId = :recipientId AND isDeleted = :isDeleted AND todoStatus = :todoStatus AND expireAt IS NOT NULL AND expireAt < :now",
        {
          recipientId,
          isDeleted: false,
          todoStatus: MessageStatus.PENDING,
          now,
        }
      )
      .execute();
  }

  async findById(recipientId: string, id: string): Promise<Message | null> {
    await this.checkAndMarkExpiredTodos(recipientId);
    return await this.messageRepository.findOne({
      where: { id, recipientId, isDeleted: false },
      relations: ["template", "template.type"],
    });
  }
}
