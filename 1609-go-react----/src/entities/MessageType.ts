import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from "typeorm";
import { MessageCategory, InAppType } from "./types";

@Entity()
export class MessageType {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({
    type: "varchar",
    enum: MessageCategory,
  })
  category!: MessageCategory;

  @Column({
    type: "varchar",
    enum: InAppType,
    nullable: true,
  })
  subType?: InAppType;

  @Column({ type: "varchar" })
  name!: string;

  @Column({ type: "text", nullable: true })
  description?: string;

  @Column({ type: "boolean", default: false })
  isTodo!: boolean;

  @CreateDateColumn()
  createdAt!: Date;

  @UpdateDateColumn()
  updatedAt!: Date;
}
