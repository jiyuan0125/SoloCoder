import { Repository, In } from 'typeorm';
import { AppDataSource } from '../config/database';
import { User } from '../models/User';
import { Evaluation } from '../models/Evaluation';
import { Course } from '../models/Course';
import { UserRole, InstructorStatus, CourseStatus, InstructorRegistrationRequest } from '../types';
import { v4 as uuidv4 } from 'uuid';

export class InstructorService {
  private userRepository: Repository<User>;
  private evaluationRepository: Repository<Evaluation>;
  private courseRepository: Repository<Course>;

  constructor() {
    this.userRepository = AppDataSource.getRepository(User);
    this.evaluationRepository = AppDataSource.getRepository(Evaluation);
    this.courseRepository = AppDataSource.getRepository(Course);
  }

  async registerInstructor(request: InstructorRegistrationRequest, email: string): Promise<User> {
    const existingUser = await this.userRepository.findOne({
      where: { email }
    });

    if (existingUser) {
      existingUser.role = UserRole.INSTRUCTOR;
      existingUser.instructorStatus = InstructorStatus.PENDING;
      existingUser.qualifications = request.qualifications;
      existingUser.expertise = request.expertise;
      existingUser.teachingExperience = request.teachingExperience;
      existingUser.name = request.name;
      return this.userRepository.save(existingUser);
    }

    const instructor = this.userRepository.create({
      id: uuidv4(),
      name: request.name,
      email,
      role: UserRole.INSTRUCTOR,
      instructorStatus: InstructorStatus.PENDING,
      qualifications: request.qualifications,
      expertise: request.expertise,
      teachingExperience: request.teachingExperience
    });

    return this.userRepository.save(instructor);
  }

  async approveInstructor(instructorId: string): Promise<User> {
    const instructor = await this.userRepository.findOne({
      where: { id: instructorId }
    });

    if (!instructor) {
      const error = new Error('讲师不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    if (instructor.role !== UserRole.INSTRUCTOR) {
      const error = new Error('该用户不是讲师');
      (error as any).statusCode = 400;
      throw error;
    }

    instructor.instructorStatus = InstructorStatus.APPROVED;
    return this.userRepository.save(instructor);
  }

  async getInstructorById(instructorId: string): Promise<User> {
    const instructor = await this.userRepository.findOne({
      where: { id: instructorId, role: UserRole.INSTRUCTOR }
    });

    if (!instructor) {
      const error = new Error('讲师不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    return instructor;
  }

  async getAllInstructors(): Promise<User[]> {
    return this.userRepository.find({
      where: { role: UserRole.INSTRUCTOR },
      order: { createdAt: 'DESC' }
    });
  }

  async getInstructorCourses(instructorId: string): Promise<Course[]> {
    await this.getInstructorById(instructorId);

    return this.courseRepository.find({
      where: { instructorId },
      order: { createdAt: 'DESC' }
    });
  }

  async rateInstructor(
    studentId: string,
    courseId: string,
    rating: number,
    comments?: string
  ): Promise<Evaluation> {
    if (rating < 1 || rating > 5) {
      const error = new Error('评分必须在 1 到 5 分之间');
      (error as any).statusCode = 400;
      throw error;
    }

    const course = await this.courseRepository.findOne({
      where: { id: courseId }
    });

    if (!course) {
      const error = new Error('课程不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    const instructor = await this.getInstructorById(course!.instructorId);

    const existingEvaluation = await this.evaluationRepository.findOne({
      where: {
        studentId,
        courseId
      }
    });

    if (existingEvaluation) {
      const error = new Error('您已经评价过该课程');
      (error as any).statusCode = 400;
      throw error;
    }

    const evaluation = this.evaluationRepository.create({
      studentId,
      courseId,
      instructorId: course!.instructorId,
      rating,
      comments: comments ?? null
    });

    const savedEvaluation = await this.evaluationRepository.save(evaluation);

    const newTotalScore = (instructor.totalRatingScore ?? 0) + rating;
    const newRatingCount = instructor.ratingCount + 1;
    const newAverageRating = newTotalScore / newRatingCount;

    instructor.totalRatingScore = newTotalScore;
    instructor.ratingCount = newRatingCount;
    instructor.averageRating = parseFloat(newAverageRating.toFixed(2));

    if (newAverageRating < 3.5) {
      instructor.instructorStatus = InstructorStatus.SUSPENDED;
    }

    await this.userRepository.save(instructor);

    return savedEvaluation;
  }
}
