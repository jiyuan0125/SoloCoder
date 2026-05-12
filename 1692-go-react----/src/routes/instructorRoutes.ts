import { Router } from 'express';
import { InstructorController } from '../controllers/InstructorController';

export class InstructorRoutes {
  public router: Router;
  private instructorController: InstructorController;

  constructor() {
    this.router = Router();
    this.instructorController = new InstructorController();
    this.initializeRoutes();
  }

  private initializeRoutes(): void {
    this.router.post('/', this.instructorController.registerInstructor);
    this.router.get('/', this.instructorController.getAllInstructors);
    this.router.get('/:instructorId', this.instructorController.getInstructor);
    this.router.post('/:instructorId/approve', this.instructorController.approveInstructor);
    this.router.get('/:instructorId/courses', this.instructorController.getInstructorCourses);
    this.router.post('/courses/:courseId/rate', this.instructorController.rateInstructor);
  }
}
