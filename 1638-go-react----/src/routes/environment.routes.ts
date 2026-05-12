import { Router } from 'express';
import { environmentController } from '../controllers/environment.controller';

const router = Router();

router.post('/', environmentController.create);
router.get('/', environmentController.list);
router.get('/:id', environmentController.getById);
router.put('/:id', environmentController.update);
router.delete('/:id', environmentController.delete);
router.post('/:id/status', environmentController.updateStatus);
router.post('/:id/clone', environmentController.clone);
router.get('/:id/config-history', environmentController.getConfigHistory);
router.post('/:id/rollback', environmentController.rollback);

export { router as environmentRoutes };
