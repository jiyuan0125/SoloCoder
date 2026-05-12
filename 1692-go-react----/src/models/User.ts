import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, OneToMany } from 'typeorm';
import { UserRole, InstructorStatus } from '../types';
import { Course } from './Course';
import { Enrollment } from './Enrollment';
import { Evaluation } from './Evaluation';
import { Certificate } from './Certificate';
import { Notification } from './Notification';

@Entity()
export class User {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ unique: true })
  email: string;

  @Column({ nullable: true })
  password: string;

  @Column({
    type: 'simple-enum',
    enum: UserRole,
    default: UserRole.STUDENT
  })
  role: UserRole;

  @Column({
    type: 'simple-enum',
    enum: InstructorStatus,
    nullable: true
  })
  instructorStatus: InstructorStatus | null;

  @Column({ type: 'text', nullable: true })
  qualifications: string | null;

  @Column({ type: 'text', nullable: true })
  expertise: string | null;

  @Column({ type: 'text', nullable: true })
  teachingExperience: string | null;

  @Column({ type: 'decimal', precision: 3, scale: 2, default: null, nullable: true })
  averageRating: number | null;

  @Column({ default: 0 })
  ratingCount: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0, nullable: true })
  totalRatingScore: number | null;

  @CreateDateColumn()
  createdAt: Date;

  @OneToMany(() => Course, course => course.instructor)
  courses: Course[];

  @OneToMany(() => Enrollment, enrollment => enrollment.student)
  enrollments: Enrollment[];

  @OneToMany(() => Evaluation, evaluation => evaluation.student)
  evaluations: Evaluation[];

  @OneToMany(() => Certificate, certificate => certificate.student)
  certificates: Certificate[];

  @OneToMany(() => Notification, notification => notification.student)
  notifications: Notification[];
}
