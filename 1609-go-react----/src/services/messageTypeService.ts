import { AppDataSource } from "../data-source";
import { MessageType } from "../entities/MessageType";
import { MessageCategory, InAppType } from "../entities/types";

export class MessageTypeService {
  private repository = AppDataSource.getRepository(MessageType);

  async create(data: {
    category: MessageCategory;
    subType?: InAppType;
    name: string;
    description?: string;
    isTodo?: boolean;
  }): Promise<MessageType> {
    const type = this.repository.create({
      category: data.category,
      subType: data.category === MessageCategory.IN_APP ? data.subType : undefined,
      name: data.name,
      description: data.description,
      isTodo: data.isTodo || false,
    });

    return await this.repository.save(type);
  }

  async findAll(): Promise<MessageType[]> {
    return await this.repository.find({
      order: { createdAt: "DESC" },
    });
  }

  async findById(id: string): Promise<MessageType | null> {
    return await this.repository.findOne({ where: { id } });
  }

  async update(
    id: string,
    data: {
      name?: string;
      description?: string;
      isTodo?: boolean;
    }
  ): Promise<MessageType | null> {
    const type = await this.findById(id);
    if (!type) return null;

    if (data.name !== undefined) type.name = data.name;
    if (data.description !== undefined) type.description = data.description;
    if (data.isTodo !== undefined) type.isTodo = data.isTodo;

    return await this.repository.save(type);
  }

  async delete(id: string): Promise<boolean> {
    const result = await this.repository.delete(id);
    return result.affected !== undefined && result.affected > 0;
  }
}
