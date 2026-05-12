import { Router } from 'express';
import { elderController } from '../controllers/elderController';
import { assessmentController } from '../controllers/assessmentController';
import { callController } from '../controllers/callController';
import { visitController } from '../controllers/visitController';
import { healthController } from '../controllers/healthController';

const router = Router();

router.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
});

router.post('/elders', elderController.createElder);
router.get('/elders', elderController.getAllElders);
router.get('/elders/:id', elderController.getElder);

router.post('/assessments', assessmentController.createAssessment);
router.get('/assessments/:id', assessmentController.getAssessment);
router.get('/elders/:elderId/assessments', assessmentController.getAssessmentsByElder);
router.get('/elders/:elderId/latest-assessment', assessmentController.getLatestAssessment);

router.post('/calls', callController.createCall);
router.get('/calls/pending', callController.getPendingCalls);
router.get('/calls/:id', callController.getCall);
router.post('/calls/:id/respond', callController.respondToCall);
router.post('/calls/:id/check-timeout', callController.checkTimeout);
router.get('/calls/:id/notifications', callController.getNotifications);
router.post('/calls/:id/close', callController.closeCall);
router.get('/elders/:elderId/calls', callController.getCallsByElder);

router.get('/elders/:elderId/visits', visitController.getVisitPlansByElder);
router.get('/elders/:elderId/visits/pending', visitController.getPendingVisitPlans);
router.get('/visits/reported', visitController.getReportedVisits);
router.get('/visits/:id', visitController.getVisitPlan);
router.post('/visits/:id/complete', visitController.completeVisit);

router.post('/health-records', healthController.createHealthRecord);
router.get('/health-records/abnormal', healthController.getAbnormalRecords);
router.get('/health-records/:id', healthController.getHealthRecord);
router.get('/elders/:elderId/health-records', healthController.getHealthRecordsByElder);

export default router;
