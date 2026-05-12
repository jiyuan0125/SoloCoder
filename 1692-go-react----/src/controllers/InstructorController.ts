import { Request, Response } from 'express';
import { InstructorService } from '../services/InstructorService';
import { InstructorRegistrationRequest, EvaluationRequest } from '../types';

export class InstructorController {
  private instructorService: InstructorService;

  constructor() {
    this.instructorService = new InstructorService();
  }

  registerInstructor = async (req: Request, res: Response): Promise<void> => {
    try {
      const request = req.body as InstructorRegistrationRequest & { email?: string };
      const email = request.email || `instructor_${Date.now()}@example.com`;
      const instructor = await this.instructorService.registerInstructor(request, email);
      res.status(201).json(instructor);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '讲师注册失败'
      });
    }
  };

  approveInstructor = async (req: Request, res: Response): Promise<void> => {
    try {
      const { instructorId } = req.params;
      const instructor = await this.instructorService.approveInstructor(instructorId);
      res.status(200).json(instructor);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '讲师审核失败'
      });
    }
  };

  getInstructor = async (req: Request, res: Response): Promise<void> => {
    try {
      const { instructorId } = req.params;
      const instructor = await this.instructorService.getInstructorById(instructorId);
      res.status(200).json(instructor);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '获取讲师信息失败'
      });
    }
  };

  getAllInstructors = async (req: Request, res: Response): Promise<void> => {
    try {
      const instructors = await this.instructorService.getAllInstructors();
      res.status(200).json(instructors);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '获取讲师列表失败'
      });
    }
  };

  getInstructorCourses = async (req: Request, res: Response): Promise<void> => {
    try {
      const { instructorId } = req.params;
      const courses = await this.instructorService.getInstructorCourses(instructorId);
      res.status(200).json(courses);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '获取讲师课程失败'
      });
    }
  };

  rateInstructor = async (req: Request, res: Response): Promise<void> => {
    try {
      const { courseId } = req.params;
      const request = req.body as EvaluationRequest;
      const evaluation = await this.instructorService.rateInstructor(
        request.studentId,
        courseId,
        request.rating,
        request.comments
      );
      res.status(201).json(evaluation);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '评价讲师失败'
      });
    }
  };
}
