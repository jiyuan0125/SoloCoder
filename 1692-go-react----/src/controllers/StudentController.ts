import { Request, Response } from 'express';
import { StudentService } from '../services/StudentService';
import { EnrollmentRequest } from '../types';

export class StudentController {
  private studentService: StudentService;

  constructor() {
    this.studentService = new StudentService();
  }

  enrollCourse = async (req: Request, res: Response): Promise<void> => {
    try {
      const request = req.body as EnrollmentRequest;
      const enrollment = await this.studentService.enrollCourse(
        request.studentId,
        request.courseId
      );
      res.status(201).json(enrollment);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '报名失败'
      });
    }
  };

  dropCourse = async (req: Request, res: Response): Promise<void> => {
    try {
      const { studentId, courseId } = req.params;
      const enrollment = await this.studentService.dropCourse(studentId, courseId);
      res.status(200).json(enrollment);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '退课失败'
      });
    }
  };

  getEnrollments = async (req: Request, res: Response): Promise<void> => {
    try {
      const { studentId } = req.params;
      const enrollments = await this.studentService.getStudentEnrollments(studentId);
      res.status(200).json(enrollments);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '获取报名记录失败'
      });
    }
  };
}
