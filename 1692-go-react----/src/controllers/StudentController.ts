import { Request, Response } from 'express';
import { StudentService, StudentRegistrationRequest } from '../services/StudentService';
import { EnrollmentRequest } from '../types';

export class StudentController {
  private studentService: StudentService;

  constructor() {
    this.studentService = new StudentService();
  }

  registerStudent = async (req: Request, res: Response): Promise<void> => {
    try {
      const request = req.body as StudentRegistrationRequest;
      const student = await this.studentService.registerStudent(request);
      res.status(201).json(student);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '学员注册失败'
      });
    }
  };

  getStudent = async (req: Request, res: Response): Promise<void> => {
    try {
      const { studentId } = req.params;
      const student = await this.studentService.getStudentById(studentId);
      res.status(200).json(student);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '获取学员信息失败'
      });
    }
  };

  getAllStudents = async (req: Request, res: Response): Promise<void> => {
    try {
      const students = await this.studentService.getAllStudents();
      res.status(200).json(students);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '获取学员列表失败'
      });
    }
  };

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
