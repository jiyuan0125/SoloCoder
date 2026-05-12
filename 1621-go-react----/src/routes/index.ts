import { Router } from 'express';
import { BackupTaskController } from '../controllers/BackupTaskController';
import { SnapshotController } from '../controllers/SnapshotController';

const router = Router();

router.post('/tasks', BackupTaskController.createTask);
router.get('/tasks', BackupTaskController.getAllTasks);
router.get('/tasks/:id', BackupTaskController.getTask);
router.put('/tasks/:id/enabled', BackupTaskController.toggleTask);
router.put('/tasks/:id/cron', BackupTaskController.updateCron);
router.delete('/tasks/:id', BackupTaskController.deleteTask);
router.post('/tasks/:id/backup', BackupTaskController.triggerManualBackup);

router.get('/snapshots', SnapshotController.getAllSnapshots);
router.get('/snapshots/:id', SnapshotController.getSnapshot);
router.delete('/snapshots/:id', SnapshotController.deleteSnapshot);

router.post('/snapshots/:snapshotId/restore', SnapshotController.initiateRestore);
router.post('/restore/:operationId/confirm', SnapshotController.confirmRestore);
router.get('/restore/:operationId', SnapshotController.getRestoreOperation);
router.get('/restore', SnapshotController.getAllRestoreOperations);

export default router;
