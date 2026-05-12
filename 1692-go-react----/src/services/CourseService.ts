import { Repository, In } from 'typeorm';
import { AppDataSource } from '../config/database';
import { Course } from '../models/Course';
import { User } from '../models/User';
import { Enrollment } from '../models/Enrollment';
import { CoursePrerequisite } from '../models/CoursePrerequisite';
import { Notification } from '../models/Notification';
import { CourseStatus, UserRole, InstructorStatus, EnrollmentStatus, CreateCourseRequest } from '../types';

export class CourseService {
  private courseRepository: Repository<Course>;
  private userRepository: Repository<User>;
  private enrollmentRepository: Repository<Enrollment>;
  private coursePrerequisiteRepository: Repository<CoursePrerequisite>;
  private notificationRepository: Repository<Notification>;

  constructor() {
    this.courseRepository = AppDataSource.getRepository(Course);
    this.userRepository = AppDataSource.getRepository(User);
    this.enrollmentRepository = AppDataSource.getRepository(Enrollment);
    this.coursePrerequisiteRepository = AppDataSource.getRepository(CoursePrerequisite);
    this.notificationRepository = AppDataSource.getRepository(Notification);
  }

  async createCourse(request: CreateCourseRequest): Promise<Course> {
    if (request.totalHours <= 0 || request.maxEnrollment <= 0) {
      const error = new Error('课时数和报名人数必须大于 0');
      (error as any).statusCode = 400;
      throw error;
    }

    const existingCourse = await this.courseRepository.findOne({
      where: { name: request.name }
    });

    if (existingCourse) {
      const error = new Error('课程名称已存在');
      (error as any).statusCode = 409;
      throw error;
    }

    const instructor = await this.userRepository.findOne({
      where: { id: request.instructorId }
    });

    if (!instructor) {
      const error = new Error('讲师不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    if (instructor.role !== UserRole.INSTRUCTOR || instructor.instructorStatus !== InstructorStatus.APPROVED) {
      const error = new Error('讲师未通过审核');
      (error as any).statusCode = 400;
      throw error;
    }

    if (instructor.ratingCount > 0 && 
        (instructor.averageRating !== null && instructor.averageRating < 3.5)) {
      const error = new Error('讲师评分过低，暂停排课');
      (error as any).statusCode = 400;
      throw error;
    }

    const activeCourses = await this.courseRepository.count({
      where: {
        instructorId: request.instructorId,
        status: In([CourseStatus.ENROLLING, CourseStatus.IN_PROGRESS])
      }
    });

    if (activeCourses >= 3) {
      const error = new Error('讲师同时进行的课程不能超过 3 门');
      (error as any).statusCode = 400;
      throw error;
    }

    const cancelledSameNameCourse = await this.courseRepository
      .createQueryBuilder('course')
      .where('course.name = :name AND course.status = :status AND course.cancelledAt >= :date', {
        name: request.name,
        status: CourseStatus.CANCELLED,
        date: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString()
      })
      .getOne();

    if (cancelledSameNameCourse) {
      const error = new Error('该课程名称在取消后 7 天内无法重新开设');
      (error as any).statusCode = 400;
      throw error;
    }

    const course = this.courseRepository.create({
      name: request.name,
      category: request.category,
      description: request.description,
      instructorId: request.instructorId,
      totalHours: request.totalHours,
      hoursPerSession: request.hoursPerSession,
      maxEnrollment: request.maxEnrollment,
      startDate: new Date(request.startDate),
      fee: request.fee ?? 0,
      status: CourseStatus.ENROLLING
    });

    const savedCourse = await this.courseRepository.save(course);

    if (request.prerequisites && request.prerequisites.length > 0) {
      const prerequisites = request.prerequisites.map(prerequisiteCourseId =>
        this.coursePrerequisiteRepository.create({
          courseId: savedCourse.id,
          prerequisiteCourseId
        })
      );
      await this.coursePrerequisiteRepository.save(prerequisites);
    }

    return savedCourse;
  }

  async getCourseById(courseId: string): Promise<Course> {
    const course = await this.courseRepository.findOne({
      where: { id: courseId },
      relations: ['instructor', 'prerequisites', 'enrollments']
    });

    if (!course) {
      const error = new Error('课程不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    return course;
  }

  async getAllCourses(): Promise<Course[]> {
    return this.courseRepository.find({
      relations: ['instructor', 'enrollments'],
      order: { createdAt: 'DESC' }
    });
  }

  async cancelCourse(courseId: string): Promise<{ success: boolean; failedNotifications: number }> {
    const course = await this.getCourseById(courseId);

    if (course.status === CourseStatus.CANCELLED) {
      const error = new Error('课程已取消');
      (error as any).statusCode = 400;
      throw error;
    }

    const enrolledStudents = await this.enrollmentRepository.find({
      where: {
        courseId,
        status: In([EnrollmentStatus.ENROLLED, EnrollmentStatus.WAITING_LIST])
      },
      relations: ['student']
    });

    course.status = CourseStatus.CANCELLED;
    course.cancelledAt = new Date();
    await this.courseRepository.save(course);

    let failedCount = 0;

    for (const enrollment of enrolledStudents) {
      try {
        const currentEnrollment = await this.enrollmentRepository.findOne({
          where: { id: enrollment.id }
        });

        if (!currentEnrollment || currentEnrollment.status === EnrollmentStatus.DROPPED) {
          continue;
        }

        const notification = this.notificationRepository.create({
          studentId: enrollment.studentId,
          courseId,
          message: `您报名的课程「${course.name}」已被取消。`,
          type: 'COURSE_CANCELLED',
          isSent: false
        });

        try {
          notification.isSent = true;
          await this.notificationRepository.save(notification);
        } catch (err) {
          failedCount++;
          notification.errorMessage = (err as Error).message;
          notification.isSent = false;
          await this.notificationRepository.save(notification);
        }
      } catch (err) {
        failedCount++;
      }
    }

    return {
      success: true,
      failedNotifications: failedCount
    };
  }

  async checkAndCancelUnderEnrolledCourses(): Promise<void> {
    const sevenDaysFromNow = new Date();
    sevenDaysFromNow.setDate(sevenDaysFromNow.getDate() + 7);

    const enrollingCourses = await this.courseRepository.find({
      where: {
        status: CourseStatus.ENROLLING
      },
      relations: ['enrollments']
    });

    for (const course of enrollingCourses) {
      const startDate = new Date(course.startDate);
      const now = new Date();
      const daysUntilStart = Math.ceil(
        (startDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
      );

      if (daysUntilStart <= 7) {
        const enrolledCount = course.enrollments.filter(
          e => e.status === EnrollmentStatus.ENROLLED
        ).length;

        if (enrolledCount < 5) {
          await this.cancelCourse(course.id);
        }
      }
    }
  }
}
