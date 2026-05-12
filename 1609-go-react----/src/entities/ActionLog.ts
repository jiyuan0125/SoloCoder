import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
} from "typeorm";

@Entity()
export class ActionLog {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "uuid" })
  messageId!: string;

  @Column({ type: "varchar" })
  action!: string;

  @Column({ type: "varchar" })
  operatorId!: string;

  @Column({ type: "text", nullable: true })
  remark?: string;

  @CreateDateColumn()
  createdAt!: Date;
}
