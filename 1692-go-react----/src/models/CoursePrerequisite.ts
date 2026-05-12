import { Entity, PrimaryGeneratedColumn, ManyToOne, JoinColumn, Column } from 'typeorm';
import { Course } from './Course';

@Entity()
export class CoursePrerequisite {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  courseId: string;

  @Column()
  prerequisiteCourseId: string;

  @ManyToOne(() => Course, course => course.prerequisites)
  @JoinColumn({ name: 'courseId' })
  course: Course;

  @ManyToOne(() => Course, course => course.dependentCourses)
  @JoinColumn({ name: 'prerequisiteCourseId' })
  prerequisiteCourse: Course;
}
