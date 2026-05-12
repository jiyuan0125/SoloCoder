import { Router } from 'express';
import { taskController } from '../controllers/taskController';

const router = Router();

router.post('/', taskController.createTask);
router.get('/', taskController.getAllTasks);
router.get('/:id', taskController.getTaskById);
router.put('/:id', taskController.updateTask);
router.delete('/:id', taskController.deleteTask);
router.get('/:id/executions', taskController.getTaskExecutions);
router.post('/:id/trigger', taskController.triggerTask);

export default router;
