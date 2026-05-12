import { Router } from 'express';
import { CourseController } from '../controllers/CourseController';

export class CourseRoutes {
  public router: Router;
  private courseController: CourseController;

  constructor() {
    this.router = Router();
    this.courseController = new CourseController();
    this.initializeRoutes();
  }

  private initializeRoutes(): void {
    this.router.post('/', this.courseController.createCourse);
    this.router.get('/', this.courseController.getAllCourses);
    this.router.get('/:courseId', this.courseController.getCourse);
    this.router.delete('/:courseId', this.courseController.cancelCourse);
    this.router.post('/check-under-enrolled', this.courseController.checkAndCancelUnderEnrolled);
  }
}
