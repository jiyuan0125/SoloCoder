import { Router, Request, Response } from 'express';
import { runQuery, runInsert, runUpdate } from '../database';
import { Transfer, TransferStatus, Material, Shelter } from '../types';

const router = Router();

const getShelterById = async (id: number): Promise<Shelter | null> => {
  const shelters = await runQuery<Shelter[]>('SELECT * FROM shelters WHERE id = ?', [id]);
  return shelters.length > 0 ? shelters[0] : null;
};

const getMaterialById = async (id: number): Promise<Material | null> => {
  const materials = await runQuery<Material[]>('SELECT * FROM materials WHERE id = ?', [id]);
  return materials.length > 0 ? materials[0] : null;
};

const getTransferById = async (id: number): Promise<Transfer | null> => {
  const transfers = await runQuery<Transfer[]>('SELECT * FROM transfers WHERE id = ?', [id]);
  return transfers.length > 0 ? transfers[0] : null;
};

const statusFlow: Record<TransferStatus, TransferStatus | null> = {
  [TransferStatus.PENDING]: TransferStatus.APPROVED,
  [TransferStatus.APPROVED]: TransferStatus.IN_TRANSIT,
  [TransferStatus.IN_TRANSIT]: TransferStatus.ARRIVED,
  [TransferStatus.ARRIVED]: TransferStatus.STORED,
  [TransferStatus.STORED]: null
};

const isValidStatusTransition = (from: TransferStatus, to: TransferStatus): boolean => {
  return statusFlow[from] === to;
};

const isExpired = (expiryDate: string): boolean => {
  return new Date(expiryDate) < new Date();
};

router.get('/', async (_req: Request, res: Response) => {
  try {
    const transfers = await runQuery<Transfer[]>(
      'SELECT * FROM transfers ORDER BY created_at DESC'
    );
    res.json(transfers);
  } catch (error) {
    console.error('获取调拨记录失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const transfer = await getTransferById(id);
    if (!transfer) {
      return res.status(404).json({ error: '调拨记录不存在' });
    }

    res.json(transfer);
  } catch (error) {
    console.error('获取调拨详情失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/shelters/:shelterId/materials/:materialId/transfer/:targetId', async (req: Request, res: Response) => {
  try {
    const fromShelterId = parseInt(req.params.shelterId);
    const materialId = parseInt(req.params.materialId);
    const toShelterId = parseInt(req.params.targetId);

    if (isNaN(fromShelterId)) {
      return res.status(404).json({ error: '调出场所ID无效' });
    }

    if (isNaN(materialId)) {
      return res.status(404).json({ error: '物资ID无效' });
    }

    if (isNaN(toShelterId)) {
      return res.status(404).json({ error: '调入场所ID无效' });
    }

    const fromShelter = await getShelterById(fromShelterId);
    if (!fromShelter) {
      return res.status(404).json({ error: '调出场所不存在' });
    }

    const toShelter = await getShelterById(toShelterId);
    if (!toShelter) {
      return res.status(404).json({ error: '调入场所不存在' });
    }

    const material = await getMaterialById(materialId);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    if (material.shelter_id !== fromShelterId) {
      return res.status(404).json({ error: '物资不属于调出场所' });
    }

    if (isExpired(material.expiry_date)) {
      return res.status(400).json({ error: '过期物资不能调拨' });
    }

    const { quantity } = req.body;
    if (!quantity || quantity <= 0 || !Number.isInteger(quantity)) {
      return res.status(400).json({ error: '调拨数量必须为正整数' });
    }

    if (quantity > material.quantity) {
      return res.status(400).json({ error: `调出数量超过库存，当前库存: ${material.quantity}` });
    }

    const transferId = await runInsert(
      `INSERT INTO transfers (from_shelter_id, to_shelter_id, material_id, material_name, quantity, status)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [fromShelterId, toShelterId, materialId, material.name, quantity, TransferStatus.PENDING]
    );

    const transfer = await getTransferById(transferId);
    res.status(201).json(transfer);
  } catch (error) {
    console.error('创建调拨记录失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/:id/status', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const transfer = await getTransferById(id);
    if (!transfer) {
      return res.status(404).json({ error: '调拨记录不存在' });
    }

    const { status } = req.body;
    if (!status) {
      return res.status(400).json({ error: '必须指定状态' });
    }

    const currentStatus = transfer.status;
    const targetStatus = status as TransferStatus;

    if (!Object.values(TransferStatus).includes(targetStatus)) {
      return res.status(400).json({ error: '无效的调拨状态' });
    }

    if (currentStatus === targetStatus) {
      return res.json(transfer);
    }

    if (!isValidStatusTransition(currentStatus, targetStatus)) {
      return res.status(400).json({
        error: `非法状态跳转，当前状态: ${currentStatus}，不能跳转到: ${targetStatus}`,
        validNext: statusFlow[currentStatus] ? [statusFlow[currentStatus]] : []
      });
    }

    const material = await getMaterialById(transfer.material_id);

    if (targetStatus === TransferStatus.APPROVED) {
      if (!material) {
        return res.status(404).json({ error: '物资不存在' });
      }
      if (transfer.quantity > material.quantity) {
        return res.status(400).json({ error: `库存不足，当前库存: ${material.quantity}` });
      }

      await runUpdate(
        `UPDATE materials SET quantity = quantity - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
        [transfer.quantity, transfer.material_id]
      );
    }

    if (targetStatus === TransferStatus.STORED) {
      if (!material) {
        return res.status(404).json({ error: '物资不存在' });
      }

      const existingMaterials = await runQuery<Material[]>(
        `SELECT * FROM materials WHERE shelter_id = ? AND name = ?`,
        [transfer.to_shelter_id, material.name]
      );

      if (existingMaterials.length > 0) {
        await runUpdate(
          `UPDATE materials SET quantity = quantity + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
          [transfer.quantity, existingMaterials[0].id]
        );
      } else {
        await runInsert(
          `INSERT INTO materials (name, shelter_id, quantity, expiry_date, min_stock)
           VALUES (?, ?, ?, ?, ?)`,
          [material.name, transfer.to_shelter_id, transfer.quantity, material.expiry_date, material.min_stock]
        );
      }
    }

    await runUpdate(
      `UPDATE transfers SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [targetStatus, id]
    );

    const updated = await getTransferById(id);
    res.json(updated);
  } catch (error) {
    console.error('更新调拨状态失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:id/cancel', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const transfer = await getTransferById(id);
    if (!transfer) {
      return res.status(404).json({ error: '调拨记录不存在' });
    }

    if (transfer.status === TransferStatus.PENDING || transfer.status === TransferStatus.APPROVED) {
      if (transfer.status === TransferStatus.APPROVED) {
        await runUpdate(
          `UPDATE materials SET quantity = quantity + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
          [transfer.quantity, transfer.material_id]
        );
      }

      await runUpdate(
        `UPDATE transfers SET status = 'CANCELLED', updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
        [id]
      );
    } else {
      return res.status(400).json({ error: '当前状态不允许取消调拨' });
    }

    const updated = await getTransferById(id);
    res.json(updated);
  } catch (error) {
    console.error('取消调拨失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
