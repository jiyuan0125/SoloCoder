import { AppDataSource } from "../data-source";
import { MessageTemplate } from "../entities/MessageTemplate";
import { Message } from "../entities/Message";
import { Language } from "../entities/types";
import { validateTemplateVariables, extractVariables } from "../utils/template";

export class MessageTemplateService {
  private templateRepository = AppDataSource.getRepository(MessageTemplate);
  private messageRepository = AppDataSource.getRepository(Message);

  async create(data: {
    typeId: string;
    language: Language;
    subject: string;
    content: string;
  }): Promise<MessageTemplate | null> {
    if (!validateTemplateVariables(data.subject) || !validateTemplateVariables(data.content)) {
      return null;
    }

    const variables = [
      ...extractVariables(data.subject),
      ...extractVariables(data.content),
    ];
    const uniqueVariables = Array.from(new Set(variables));

    const template = this.templateRepository.create({
      typeId: data.typeId,
      language: data.language,
      subject: data.subject,
      content: data.content,
      variables: uniqueVariables.length > 0 ? JSON.stringify(uniqueVariables) : undefined,
    });

    return await this.templateRepository.save(template);
  }

  async findAll(): Promise<MessageTemplate[]> {
    return await this.templateRepository.find({
      relations: ["type"],
      order: { createdAt: "DESC" },
    });
  }

  async findById(id: string): Promise<MessageTemplate | null> {
    return await this.templateRepository.findOne({
      where: { id },
      relations: ["type"],
    });
  }

  async update(
    id: string,
    data: {
      subject?: string;
      content?: string;
    }
  ): Promise<MessageTemplate | null> {
    const template = await this.findById(id);
    if (!template) return null;

    if (data.subject !== undefined) {
      if (!validateTemplateVariables(data.subject)) return null;
      template.subject = data.subject;
    }
    if (data.content !== undefined) {
      if (!validateTemplateVariables(data.content)) return null;
      template.content = data.content;
    }

    const variables = [
      ...extractVariables(template.subject),
      ...extractVariables(template.content),
    ];
    const uniqueVariables = Array.from(new Set(variables));
    template.variables = uniqueVariables.length > 0 ? JSON.stringify(uniqueVariables) : undefined;

    return await this.templateRepository.save(template);
  }

  async delete(id: string): Promise<{ success: boolean; hasReferences: boolean }> {
    const messageCount = await this.messageRepository.count({
      where: { templateId: id },
    });

    if (messageCount > 0) {
      return { success: false, hasReferences: true };
    }

    const result = await this.templateRepository.delete(id);
    return {
      success: result.affected !== undefined && result.affected > 0,
      hasReferences: false,
    };
  }
}
