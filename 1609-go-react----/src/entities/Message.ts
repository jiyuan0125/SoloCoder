import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
} from "typeorm";
import { MessageTemplate } from "./MessageTemplate";
import { MessageStatus } from "./types";

@Entity()
export class Message {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "uuid" })
  templateId!: string;

  @ManyToOne(() => MessageTemplate)
  @JoinColumn({ name: "templateId" })
  template!: MessageTemplate;

  @Column({ type: "varchar" })
  recipientId!: string;

  @Column({ type: "varchar" })
  subject!: string;

  @Column({ type: "text" })
  content!: string;

  @Column({ type: "text", nullable: true })
  variablesJson?: string;

  @Column({ type: "boolean", default: false })
  isRead!: boolean;

  @Column({ type: "boolean", default: false })
  isPinned!: boolean;

  @Column({ type: "boolean", default: false })
  isDeleted!: boolean;

  @Column({
    type: "varchar",
    enum: MessageStatus,
    nullable: true,
  })
  todoStatus?: MessageStatus;

  @Column({ type: "datetime", nullable: true })
  expireAt?: Date;

  @Column({ type: "varchar", nullable: true })
  processedBy?: string;

  @Column({ type: "datetime", nullable: true })
  processedAt?: Date;

  @CreateDateColumn()
  createdAt!: Date;

  @UpdateDateColumn()
  updatedAt!: Date;
}
