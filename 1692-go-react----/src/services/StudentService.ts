import { Repository, In } from 'typeorm';
import { AppDataSource } from '../config/database';
import { User } from '../models/User';
import { Course } from '../models/Course';
import { Enrollment } from '../models/Enrollment';
import { CoursePrerequisite } from '../models/CoursePrerequisite';
import { CourseStatus, UserRole, EnrollmentStatus } from '../types';
import { v4 as uuidv4 } from 'uuid';

export interface StudentRegistrationRequest {
  name: string;
  email: string;
}

export class StudentService {
  private userRepository: Repository<User>;
  private courseRepository: Repository<Course>;
  private enrollmentRepository: Repository<Enrollment>;
  private coursePrerequisiteRepository: Repository<CoursePrerequisite>;

  constructor() {
    this.userRepository = AppDataSource.getRepository(User);
    this.courseRepository = AppDataSource.getRepository(Course);
    this.enrollmentRepository = AppDataSource.getRepository(Enrollment);
    this.coursePrerequisiteRepository = AppDataSource.getRepository(CoursePrerequisite);
  }

  async registerStudent(request: StudentRegistrationRequest): Promise<User> {
    const existingUser = await this.userRepository.findOne({
      where: { email: request.email }
    });

    if (existingUser) {
      const error = new Error('邮箱已被注册');
      (error as any).statusCode = 409;
      throw error;
    }

    const student = this.userRepository.create({
      id: uuidv4(),
      name: request.name,
      email: request.email,
      role: UserRole.STUDENT
    });

    return this.userRepository.save(student);
  }

  async getStudentById(studentId: string): Promise<User> {
    const student = await this.userRepository.findOne({
      where: { id: studentId, role: UserRole.STUDENT }
    });

    if (!student) {
      const error = new Error('学员不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    return student;
  }

  async getAllStudents(): Promise<User[]> {
    return this.userRepository.find({
      where: { role: UserRole.STUDENT },
      order: { createdAt: 'DESC' }
    });
  }

  async enrollCourse(studentId: string, courseId: string): Promise<Enrollment> {
    const student = await this.userRepository.findOne({
      where: { id: studentId, role: UserRole.STUDENT }
    });

    if (!student) {
      const error = new Error('学员不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    const course = await this.courseRepository.findOne({
      where: { id: courseId },
      relations: ['enrollments']
    });

    if (!course) {
      const error = new Error('课程不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    if (course.status !== CourseStatus.ENROLLING) {
      const error = new Error('课程不在招生中');
      (error as any).statusCode = 400;
      throw error;
    }

    const existingEnrollment = await this.enrollmentRepository.findOne({
      where: {
        studentId,
        courseId,
        status: In([
          EnrollmentStatus.ENROLLED,
          EnrollmentStatus.WAITING_LIST,
          EnrollmentStatus.PENDING,
          EnrollmentStatus.COMPLETED
        ])
      }
    });

    if (existingEnrollment) {
      const error = new Error('您已报名该课程');
      (error as any).statusCode = 409;
      throw error;
    }

    const prerequisites = await this.coursePrerequisiteRepository.find({
      where: { courseId }
    });

    if (prerequisites.length > 0) {
      const prerequisiteCourseIds = prerequisites.map(p => p.prerequisiteCourseId);
      
      const completedPrerequisites = await this.enrollmentRepository.count({
        where: {
          studentId,
          courseId: In(prerequisiteCourseIds),
          status: EnrollmentStatus.COMPLETED
        }
      });

      if (completedPrerequisites < prerequisiteCourseIds.length) {
        const error = new Error('未满足前置课程要求');
        (error as any).statusCode = 403;
        throw error;
      }
    }

    const enrolledCount = course.enrollments.filter(
      e => e.status === EnrollmentStatus.ENROLLED
    ).length;

    let enrollmentStatus: EnrollmentStatus;
    let waitingListPosition = 0;

    if (enrolledCount < course.maxEnrollment) {
      enrollmentStatus = EnrollmentStatus.ENROLLED;
    } else {
      const waitingListCount = course.enrollments.filter(
        e => e.status === EnrollmentStatus.WAITING_LIST
      ).length;
      enrollmentStatus = EnrollmentStatus.WAITING_LIST;
      waitingListPosition = waitingListCount + 1;
    }

    const enrollment = this.enrollmentRepository.create({
      studentId,
      courseId,
      status: enrollmentStatus,
      waitingListPosition,
      enrolledAt: enrollmentStatus === EnrollmentStatus.ENROLLED ? new Date() : null
    });

    return this.enrollmentRepository.save(enrollment);
  }

  async dropCourse(studentId: string, courseId: string): Promise<Enrollment> {
    const enrollment = await this.enrollmentRepository.findOne({
      where: {
        studentId,
        courseId
      },
      relations: ['course']
    });

    if (!enrollment) {
      const error = new Error('未找到该报名记录');
      (error as any).statusCode = 404;
      throw error;
    }

    if (enrollment.status !== EnrollmentStatus.ENROLLED && enrollment.status !== EnrollmentStatus.WAITING_LIST) {
      const error = new Error('无法退课');
      (error as any).statusCode = 400;
      throw error;
    }

    const course = enrollment.course;
    const startDate = new Date(course.startDate);
    const now = new Date();
    const daysUntilStart = Math.ceil(
      (startDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
    );

    if (daysUntilStart <= 3) {
      const error = new Error('开课前 3 天内无法退课');
      (error as any).statusCode = 400;
      throw error;
    }

    const wasEnrolled = enrollment.status === EnrollmentStatus.ENROLLED;
    
    enrollment.status = EnrollmentStatus.DROPPED;
    enrollment.droppedAt = new Date();
    enrollment.waitingListPosition = 0;

    await this.enrollmentRepository.save(enrollment);

    if (wasEnrolled) {
      await this.processWaitingList(courseId);
    }

    return enrollment;
  }

  async processWaitingList(courseId: string): Promise<void> {
    const course = await this.courseRepository.findOne({
      where: { id: courseId },
      relations: ['enrollments']
    });

    if (!course) {
      return;
    }

    const enrolledCount = course.enrollments.filter(
      e => e.status === EnrollmentStatus.ENROLLED
    ).length;

    if (enrolledCount >= course.maxEnrollment) {
      return;
    }

    const waitingList = course.enrollments
      .filter(e => e.status === EnrollmentStatus.WAITING_LIST)
      .sort((a, b) => a.waitingListPosition - b.waitingListPosition);

    for (const waitingEnrollment of waitingList) {
      if (enrolledCount >= course.maxEnrollment) {
        break;
      }

      waitingEnrollment.status = EnrollmentStatus.ENROLLED;
      waitingEnrollment.enrolledAt = new Date();
      waitingEnrollment.waitingListPosition = 0;

      await this.enrollmentRepository.save(waitingEnrollment);
      
      const currentEnrolledCount = course.enrollments.filter(
        e => e.status === EnrollmentStatus.ENROLLED
      ).length;
    }

    const remainingWaitingList = course.enrollments
      .filter(e => e.status === EnrollmentStatus.WAITING_LIST)
      .sort((a, b) => a.waitingListPosition - b.waitingListPosition);

    for (let i = 0; i < remainingWaitingList.length; i++) {
      remainingWaitingList[i].waitingListPosition = i + 1;
      await this.enrollmentRepository.save(remainingWaitingList[i]);
    }
  }

  async getStudentEnrollments(studentId: string): Promise<Enrollment[]> {
    return this.enrollmentRepository.find({
      where: { studentId },
      relations: ['course'],
      order: { createdAt: 'DESC' }
    });
  }
}
