import { Repository } from 'typeorm';
import { AppDataSource } from '../config/database';
import { Certificate } from '../models/Certificate';
import { Enrollment } from '../models/Enrollment';
import { Evaluation } from '../models/Evaluation';
import { Course } from '../models/Course';
import { User } from '../models/User';
import { EnrollmentStatus, CourseStatus, UserRole } from '../types';

export class CertificateService {
  private certificateRepository: Repository<Certificate>;
  private enrollmentRepository: Repository<Enrollment>;
  private evaluationRepository: Repository<Evaluation>;
  private courseRepository: Repository<Course>;
  private userRepository: Repository<User>;

  constructor() {
    this.certificateRepository = AppDataSource.getRepository(Certificate);
    this.enrollmentRepository = AppDataSource.getRepository(Enrollment);
    this.evaluationRepository = AppDataSource.getRepository(Evaluation);
    this.courseRepository = AppDataSource.getRepository(Course);
    this.userRepository = AppDataSource.getRepository(User);
  }

  async issueCertificate(studentId: string, courseId: string): Promise<Certificate> {
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

    if (enrollment.status !== EnrollmentStatus.COMPLETED) {
      const error = new Error('课程未完成，无法颁发证书');
      (error as any).statusCode = 400;
      throw error;
    }

    const existingCertificate = await this.certificateRepository.findOne({
      where: {
        studentId,
        courseId
      }
    });

    if (existingCertificate) {
      const error = new Error('该课程证书已颁发');
      (error as any).statusCode = 400;
      throw error;
    }

    const totalHours = enrollment.course.totalHours;
    const attendanceRate = totalHours > 0 
      ? (enrollment.attendedHours / totalHours) * 100 
      : 0;

    const evaluation = await this.evaluationRepository.findOne({
      where: {
        studentId,
        courseId
      }
    });

    const minAttendanceRate = 80;
    const isIssued = attendanceRate >= minAttendanceRate && 
      (evaluation === null || evaluation.rating >= 2);

    const certificateNumber = await this.generateCertificateNumber();

    const certificate = this.certificateRepository.create({
      certificateNumber,
      studentId,
      courseId,
      issueDate: new Date(),
      attendanceRate: parseFloat(attendanceRate.toFixed(2)),
      evaluationRating: evaluation?.rating ?? null,
      isIssued
    });

    return this.certificateRepository.save(certificate);
  }

  private async generateCertificateNumber(): Promise<string> {
    const currentYear = new Date().getFullYear().toString();
    const prefix = 'PE';
    let maxAttempts = 100;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const certificatesThisYear = await this.certificateRepository
        .createQueryBuilder('certificate')
        .where('certificate.certificateNumber LIKE :pattern', {
          pattern: `${prefix}${currentYear}%`
        })
        .getCount();

      const nextNumber = certificatesThisYear + 1 + attempt;
      const sequentialNumber = nextNumber.toString().padStart(4, '0');
      const certificateNumber = `${prefix}${currentYear}${sequentialNumber}`;

      const existing = await this.certificateRepository.findOne({
        where: { certificateNumber }
      });

      if (!existing) {
        return certificateNumber;
      }
    }

    const timestamp = Date.now().toString().slice(-4);
    const backupNumber = `${prefix}${currentYear}${timestamp}`;
    const existing = await this.certificateRepository.findOne({
      where: { certificateNumber: backupNumber }
    });

    if (!existing) {
      return backupNumber;
    }

    throw new Error('生成证书编号失败，请稍后重试');
  }

  async getCertificateById(certificateId: string): Promise<Certificate> {
    const certificate = await this.certificateRepository.findOne({
      where: { id: certificateId },
      relations: ['student', 'course']
    });

    if (!certificate) {
      const error = new Error('证书不存在');
      (error as any).statusCode = 404;
      throw error;
    }

    return certificate;
  }

  async getStudentCertificates(studentId: string): Promise<Certificate[]> {
    return this.certificateRepository.find({
      where: { studentId },
      relations: ['course'],
      order: { createdAt: 'DESC' }
    });
  }
}
