import { Router } from 'express';
import multer from 'multer';
import { createFeedback, getAllFeedbacks, getFeedbackById, updatePriority, updateStatus, getStatusLogs, deleteFeedback, batchClaim, batchResolve } from '../controllers/feedbackController';

const router = Router();

const upload = multer({
  dest: 'uploads/',
  limits: {
    fileSize: 10 * 1024 * 1024
  },
  fileFilter: (_req, _file, cb) => {
    cb(null, true);
  }
});

router.get('/', getAllFeedbacks);
router.get('/:id', getFeedbackById);
router.post('/', upload.array('attachments'), createFeedback);
router.put('/:id/priority', updatePriority);
router.put('/:id/status', updateStatus);
router.get('/:id/status-logs', getStatusLogs);
router.delete('/:id', deleteFeedback);

router.post('/batch-claim', batchClaim);
router.post('/batch-resolve', batchResolve);

export default router;
