import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, ManyToOne, JoinColumn } from 'typeorm';
import { EnrollmentStatus } from '../types';
import { User } from './User';
import { Course } from './Course';

@Entity()
export class Enrollment {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  studentId: string;

  @Column()
  courseId: string;

  @Column({
    type: 'simple-enum',
    enum: EnrollmentStatus,
    default: EnrollmentStatus.PENDING
  })
  status: EnrollmentStatus;

  @Column({ type: 'int', default: 0 })
  attendedHours: number;

  @Column({ type: 'int', default: 0 })
  totalSessions: number;

  @Column({ type: 'int', default: 0 })
  attendedSessions: number;

  @Column({ type: 'datetime', nullable: true })
  enrolledAt: Date | null;

  @Column({ type: 'datetime', nullable: true })
  droppedAt: Date | null;

  @Column({ type: 'datetime', nullable: true })
  completedAt: Date | null;

  @Column({ type: 'int', default: 0 })
  waitingListPosition: number;

  @CreateDateColumn()
  createdAt: Date;

  @ManyToOne(() => User, user => user.enrollments)
  @JoinColumn({ name: 'studentId' })
  student: User;

  @ManyToOne(() => Course, course => course.enrollments)
  @JoinColumn({ name: 'courseId' })
  course: Course;
}
