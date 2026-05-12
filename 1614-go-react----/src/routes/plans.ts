import { Router } from 'express';
import { planController } from '../controllers/planController';

const router = Router();

router.post('/', planController.createPlan);
router.get('/', planController.getAllPlans);
router.get('/:id', planController.getPlan);
router.put('/:id', planController.updatePlan);
router.delete('/:id', planController.deletePlan);

router.post('/:id/start', planController.startGrayscale);
router.post('/:id/full-release', planController.fullRelease);
router.post('/:id/complete', planController.complete);
router.post('/:id/pause', planController.pause);
router.post('/:id/resume', planController.resume);
router.post('/:id/rollback', planController.rollback);

router.post('/:id/record', planController.recordRequest);
router.get('/:id/error-rate', planController.getErrorRate);

router.post('/:id/resolve-version', planController.resolveVersion);
router.post('/resolve-version', planController.resolveVersionByAppId);

export default router;
