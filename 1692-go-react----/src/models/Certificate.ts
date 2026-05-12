import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, ManyToOne, JoinColumn, Index } from 'typeorm';
import { User } from './User';
import { Course } from './Course';

@Entity()
@Index(['certificateNumber'], { unique: true })
export class Certificate {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  certificateNumber: string;

  @Column()
  studentId: string;

  @Column()
  courseId: string;

  @Column({ type: 'date' })
  issueDate: Date;

  @Column({ type: 'decimal', precision: 5, scale: 2 })
  attendanceRate: number;

  @Column({ type: 'int', nullable: true })
  evaluationRating: number | null;

  @Column({ type: 'boolean', default: true })
  isIssued: boolean;

  @CreateDateColumn()
  createdAt: Date;

  @ManyToOne(() => User, user => user.certificates)
  @JoinColumn({ name: 'studentId' })
  student: User;

  @ManyToOne(() => Course, course => course.certificates)
  @JoinColumn({ name: 'courseId' })
  course: Course;
}
