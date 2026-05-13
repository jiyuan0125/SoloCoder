import { Router } from 'express';
import { StudentController } from '../controllers/StudentController';

export class StudentRoutes {
  public router: Router;
  private studentController: StudentController;

  constructor() {
    this.router = Router();
    this.studentController = new StudentController();
    this.initializeRoutes();
  }

  private initializeRoutes(): void {
    this.router.post('/', this.studentController.registerStudent);
    this.router.get('/', this.studentController.getAllStudents);
    this.router.get('/:studentId', this.studentController.getStudent);
    this.router.post('/enroll', this.studentController.enrollCourse);
    this.router.delete('/:studentId/courses/:courseId', this.studentController.dropCourse);
    this.router.get('/:studentId/enrollments', this.studentController.getEnrollments);
  }
}
