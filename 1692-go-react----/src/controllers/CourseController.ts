import { Request, Response } from 'express';
import { CourseService } from '../services/CourseService';
import { CreateCourseRequest } from '../types';

export class CourseController {
  private courseService: CourseService;

  constructor() {
    this.courseService = new CourseService();
  }

  createCourse = async (req: Request, res: Response): Promise<void> => {
    try {
      const courseData = req.body as CreateCourseRequest;
      const course = await this.courseService.createCourse(courseData);
      res.status(201).json(course);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '创建课程失败'
      });
    }
  };

  getCourse = async (req: Request, res: Response): Promise<void> => {
    try {
      const { courseId } = req.params;
      const course = await this.courseService.getCourseById(courseId);
      res.status(200).json(course);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '获取课程失败'
      });
    }
  };

  getAllCourses = async (req: Request, res: Response): Promise<void> => {
    try {
      const courses = await this.courseService.getAllCourses();
      res.status(200).json(courses);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '获取课程列表失败'
      });
    }
  };

  cancelCourse = async (req: Request, res: Response): Promise<void> => {
    try {
      const { courseId } = req.params;
      const result = await this.courseService.cancelCourse(courseId);
      res.status(200).json({
        success: true,
        message: '课程已取消',
        failedNotifications: result.failedNotifications
      });
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '取消课程失败'
      });
    }
  };

  checkAndCancelUnderEnrolled = async (req: Request, res: Response): Promise<void> => {
    try {
      await this.courseService.checkAndCancelUnderEnrolledCourses();
      res.status(200).json({
        success: true,
        message: '已检查并取消招生不足的课程'
      });
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '检查课程失败'
      });
    }
  };
}
