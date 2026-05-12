import { Router } from 'express';
import { taskController } from '../controllers/task';

const router = Router();

router.get('/', taskController.list);
router.post('/', taskController.create);

router.get('/:id', taskController.get);
router.patch('/:id', taskController.update);
router.delete('/:id', taskController.remove);

router.post('/:id/start', taskController.start);
router.patch('/:id/status', taskController.updateStatus);

router.get('/:id/report', taskController.listReports);
router.get('/:id/report/latest', taskController.getLatestReport);

router.get('/:id/manual-items', taskController.getManualItems);

export default router;
