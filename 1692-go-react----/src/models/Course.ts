import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, UpdateDateColumn, ManyToOne, OneToMany, JoinColumn } from 'typeorm';
import { CourseStatus } from '../types';
import { User } from './User';
import { Enrollment } from './Enrollment';
import { Evaluation } from './Evaluation';
import { Certificate } from './Certificate';
import { Notification } from './Notification';
import { CoursePrerequisite } from './CoursePrerequisite';

@Entity()
export class Course {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  name: string;

  @Column()
  category: string;

  @Column({ type: 'text' })
  description: string;

  @Column()
  instructorId: string;

  @Column({ type: 'int' })
  totalHours: number;

  @Column({ type: 'decimal', precision: 5, scale: 2 })
  hoursPerSession: number;

  @Column({ type: 'int' })
  maxEnrollment: number;

  @Column({ type: 'date' })
  startDate: Date;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0 })
  fee: number;

  @Column({
    type: 'simple-enum',
    enum: CourseStatus,
    default: CourseStatus.DRAFT
  })
  status: CourseStatus;

  @Column({ type: 'boolean', default: false })
  isAutoCancelled: boolean;

  @Column({ type: 'datetime', nullable: true })
  cancelledAt: Date | null;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  @ManyToOne(() => User, user => user.courses)
  @JoinColumn({ name: 'instructorId' })
  instructor: User;

  @OneToMany(() => Enrollment, enrollment => enrollment.course)
  enrollments: Enrollment[];

  @OneToMany(() => Evaluation, evaluation => evaluation.course)
  evaluations: Evaluation[];

  @OneToMany(() => Certificate, certificate => certificate.course)
  certificates: Certificate[];

  @OneToMany(() => Notification, notification => notification.course)
  notifications: Notification[];

  @OneToMany(() => CoursePrerequisite, prerequisite => prerequisite.course)
  prerequisites: CoursePrerequisite[];

  @OneToMany(() => CoursePrerequisite, prerequisite => prerequisite.prerequisiteCourse)
  dependentCourses: CoursePrerequisite[];
}
