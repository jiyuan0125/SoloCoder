import { Request, Response } from 'express';
import { SnapshotService } from '../services/SnapshotService';
import { RestoreService } from '../services/RestoreService';

export class SnapshotController {
  static async getAllSnapshots(req: Request, res: Response): Promise<void> {
    try {
      const snapshots = SnapshotService.getInstance().getAllSnapshots();
      res.json(snapshots);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async getSnapshot(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const snapshot = SnapshotService.getInstance().getSnapshot(id);
      
      if (!snapshot) {
        res.status(404).json({ error: 'Snapshot not found' });
        return;
      }

      res.json(snapshot);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async deleteSnapshot(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const result = SnapshotService.getInstance().deleteSnapshot(id);
      
      if (result.error) {
        if (result.conflict) {
          res.status(409).json({ 
            error: result.error,
            reason: '快照被恢复操作引用中'
          });
          return;
        }
        
        if (result.error === 'Snapshot not found') {
          res.status(404).json({ error: result.error });
          return;
        }
        
        res.status(400).json({ error: result.error });
        return;
      }

      res.status(204).send();
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async initiateRestore(req: Request, res: Response): Promise<void> {
    try {
      const { snapshotId } = req.params;
      const result = RestoreService.getInstance().initiateRestore(snapshotId);
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      res.status(201).json(result.operation);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async confirmRestore(req: Request, res: Response): Promise<void> {
    try {
      const { operationId } = req.params;
      const { confirm, acknowledgeDataLoss } = req.body;
      
      if (typeof confirm !== 'boolean' || typeof acknowledgeDataLoss !== 'boolean') {
        res.status(400).json({ 
          error: 'Confirm restore requires both "confirm" and "acknowledgeDataLoss" boolean fields' 
        });
        return;
      }

      const result = RestoreService.getInstance().confirmRestore(operationId, {
        confirm,
        acknowledgeDataLoss
      });
      
      if (result.error) {
        res.status(400).json({ error: result.error });
        return;
      }

      res.json({ success: true });
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async getRestoreOperation(req: Request, res: Response): Promise<void> {
    try {
      const { operationId } = req.params;
      const operation = RestoreService.getInstance().getOperation(operationId);
      
      if (!operation) {
        res.status(404).json({ error: 'Restore operation not found' });
        return;
      }

      res.json(operation);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }

  static async getAllRestoreOperations(req: Request, res: Response): Promise<void> {
    try {
      const operations = RestoreService.getInstance().getAllOperations();
      res.json(operations);
    } catch (error) {
      res.status(500).json({ error: error instanceof Error ? error.message : 'Internal server error' });
    }
  }
}
