import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
} from "typeorm";
import { MessageType } from "./MessageType";
import { Language } from "./types";

@Entity()
export class MessageTemplate {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "uuid" })
  typeId!: string;

  @ManyToOne(() => MessageType)
  @JoinColumn({ name: "typeId" })
  type!: MessageType;

  @Column({
    type: "varchar",
    enum: Language,
  })
  language!: Language;

  @Column({ type: "varchar" })
  subject!: string;

  @Column({ type: "text" })
  content!: string;

  @Column({ type: "text", nullable: true })
  variables?: string;

  @CreateDateColumn()
  createdAt!: Date;

  @UpdateDateColumn()
  updatedAt!: Date;
}
