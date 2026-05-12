import { Router, Request, Response } from 'express';
import { runQuery, runInsert, runUpdate } from '../database';
import { Material, Shelter } from '../types';

const router = Router();

const getMaterialById = async (id: number): Promise<Material | null> => {
  const materials = await runQuery<Material[]>('SELECT * FROM materials WHERE id = ?', [id]);
  return materials.length > 0 ? materials[0] : null;
};

const getShelterById = async (id: number): Promise<Shelter | null> => {
  const shelters = await runQuery<Shelter[]>('SELECT * FROM shelters WHERE id = ?', [id]);
  return shelters.length > 0 ? shelters[0] : null;
};

const checkReplenishment = async (material: Material): Promise<void> => {
  if (material.quantity < material.min_stock) {
    const existingTodos = await runQuery<any[]>(
      `SELECT * FROM replenishment_todos 
       WHERE material_id = ? AND status = 'PENDING'`,
      [material.id]
    );

    if (existingTodos.length === 0) {
      await runInsert(
        `INSERT INTO replenishment_todos (material_id, shelter_id, material_name, current_stock, min_stock, required)
         VALUES (?, ?, ?, ?, ?, ?)`,
        [
          material.id,
          material.shelter_id,
          material.name,
          material.quantity,
          material.min_stock,
          material.min_stock - material.quantity
        ]
      );
    }
  }
};

const isExpired = (expiryDate: string): boolean => {
  return new Date(expiryDate) < new Date();
};

const isNearExpiry = (expiryDate: string): boolean => {
  const expiry = new Date(expiryDate);
  const now = new Date();
  const thirtyDaysLater = new Date();
  thirtyDaysLater.setDate(now.getDate() + 30);
  return expiry > now && expiry <= thirtyDaysLater;
};

router.get('/', async (_req: Request, res: Response) => {
  try {
    const materials = await runQuery<Material[]>('SELECT * FROM materials ORDER BY shelter_id, name');
    const result = materials.map(m => ({
      ...m,
      isExpired: isExpired(m.expiry_date),
      isNearExpiry: isNearExpiry(m.expiry_date),
      isLowStock: m.quantity < m.min_stock
    }));
    res.json(result);
  } catch (error) {
    console.error('获取物资列表失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const material = await getMaterialById(id);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    res.json({
      ...material,
      isExpired: isExpired(material.expiry_date),
      isNearExpiry: isNearExpiry(material.expiry_date),
      isLowStock: material.quantity < material.min_stock
    });
  } catch (error) {
    console.error('获取物资详情失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const { shelterId, name, quantity, expiryDate, minStock } = req.body;

    if (!shelterId) {
      return res.status(400).json({ error: '必须指定避难场所ID' });
    }

    if (!name || name.trim() === '') {
      return res.status(400).json({ error: '物资名称不能为空' });
    }

    if (quantity === undefined || quantity < 0 || !Number.isInteger(quantity)) {
      return res.status(400).json({ error: '库存数量必须为非负整数' });
    }

    if (minStock === undefined || minStock < 0 || !Number.isInteger(minStock)) {
      return res.status(400).json({ error: '最低储备量必须为非负整数' });
    }

    if (!expiryDate) {
      return res.status(400).json({ error: '必须指定保质期' });
    }

    const shelter = await getShelterById(shelterId);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    const materialId = await runInsert(
      `INSERT INTO materials (name, shelter_id, quantity, expiry_date, min_stock)
       VALUES (?, ?, ?, ?, ?)`,
      [name.trim(), shelterId, quantity, expiryDate, minStock]
    );

    const material = await getMaterialById(materialId);
    if (material) {
      await checkReplenishment(material);
    }

    res.status(201).json(material);
  } catch (error) {
    console.error('创建物资失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const material = await getMaterialById(id);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    const { name, quantity, expiryDate, minStock } = req.body;

    if (name !== undefined && name.trim() === '') {
      return res.status(400).json({ error: '物资名称不能为空' });
    }

    if (quantity !== undefined && (quantity < 0 || !Number.isInteger(quantity))) {
      return res.status(400).json({ error: '库存数量必须为非负整数' });
    }

    if (minStock !== undefined && (minStock < 0 || !Number.isInteger(minStock))) {
      return res.status(400).json({ error: '最低储备量必须为非负整数' });
    }

    const finalName = name !== undefined ? name.trim() : material.name;
    const finalQuantity = quantity !== undefined ? quantity : material.quantity;
    const finalExpiryDate = expiryDate || material.expiry_date;
    const finalMinStock = minStock !== undefined ? minStock : material.min_stock;

    await runUpdate(
      `UPDATE materials SET name = ?, quantity = ?, expiry_date = ?, min_stock = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [finalName, finalQuantity, finalExpiryDate, finalMinStock, id]
    );

    const updated = await getMaterialById(id);
    if (updated) {
      await checkReplenishment(updated);
    }

    res.json(updated);
  } catch (error) {
    console.error('更新物资失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const material = await getMaterialById(id);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    await runUpdate('DELETE FROM materials WHERE id = ?', [id]);
    res.status(204).send();
  } catch (error) {
    console.error('删除物资失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:id/issue', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const material = await getMaterialById(id);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    if (isExpired(material.expiry_date)) {
      return res.status(400).json({ error: '过期物资不能发放，必须报废' });
    }

    const { quantity } = req.body;
    if (!quantity || quantity <= 0 || !Number.isInteger(quantity)) {
      return res.status(400).json({ error: '发放数量必须为正整数' });
    }

    if (quantity > material.quantity) {
      return res.status(400).json({ error: `库存不足，当前库存: ${material.quantity}` });
    }

    const newQuantity = material.quantity - quantity;
    await runUpdate(
      `UPDATE materials SET quantity = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [newQuantity, id]
    );

    const updated = await getMaterialById(id);
    if (updated) {
      await checkReplenishment(updated);
    }

    res.json({
      success: true,
      material: updated,
      issued: quantity
    });
  } catch (error) {
    console.error('发放物资失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:id/scrap', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const material = await getMaterialById(id);
    if (!material) {
      return res.status(404).json({ error: '物资不存在' });
    }

    const { reason } = req.body;

    await runUpdate(
      `UPDATE materials SET quantity = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [id]
    );

    const updated = await getMaterialById(id);
    res.json({
      success: true,
      material: updated,
      reason: reason || '物资已报废'
    });
  } catch (error) {
    console.error('报废物资失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/replenishment/todos', async (_req: Request, res: Response) => {
  try {
    const todos = await runQuery<any[]>(
      `SELECT * FROM replenishment_todos ORDER BY created_at DESC`
    );
    res.json(todos);
  } catch (error) {
    console.error('获取补货待办失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
