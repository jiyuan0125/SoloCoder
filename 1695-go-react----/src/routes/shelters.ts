import { Router, Request, Response } from 'express';
import { runQuery, runInsert, runUpdate } from '../database';
import { Shelter, ShelterStatus } from '../types';

const router = Router();

const isValidStatus = (status: string): boolean => {
  return Object.values(ShelterStatus).includes(status as ShelterStatus);
};

const getShelterById = async (id: number): Promise<Shelter | null> => {
  const shelters = await runQuery<Shelter[]>('SELECT * FROM shelters WHERE id = ?', [id]);
  return shelters.length > 0 ? shelters[0] : null;
};

router.get('/', async (_req: Request, res: Response) => {
  try {
    const shelters = await runQuery<Shelter[]>('SELECT * FROM shelters ORDER BY id');
    const result = shelters.map(shelter => ({
      ...shelter,
      remaining: shelter.capacity - shelter.occupied,
      occupancyRate: Math.round((shelter.occupied / shelter.capacity) * 100)
    }));
    res.json(result);
  } catch (error) {
    console.error('获取避难场所列表失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const shelter = await getShelterById(id);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    res.json({
      ...shelter,
      remaining: shelter.capacity - shelter.occupied,
      occupancyRate: Math.round((shelter.occupied / shelter.capacity) * 100)
    });
  } catch (error) {
    console.error('获取避难场所详情失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const { name, location, type, area, capacity, facilities, manager, status, latitude, longitude } = req.body;

    if (!name || !location || !type || !area || capacity === undefined || !facilities || !manager) {
      return res.status(400).json({ error: '缺少必要字段' });
    }

    if (capacity <= 0) {
      return res.status(400).json({ error: '可容纳人数必须为正整数' });
    }

    const finalStatus = status || ShelterStatus.AVAILABLE;
    if (!isValidStatus(finalStatus)) {
      return res.status(400).json({ error: '状态无效，只能是 AVAILABLE, MAINTENANCE, FULL, UNAVAILABLE' });
    }

    const existing = await runQuery<Shelter[]>('SELECT id FROM shelters WHERE name = ?', [name]);
    if (existing.length > 0) {
      return res.status(409).json({ error: '场所名称已存在' });
    }

    const shelterId = await runInsert(
      `INSERT INTO shelters (name, location, type, area, capacity, facilities, manager, status, latitude, longitude)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [name, location, type, area, capacity, facilities, manager, finalStatus, latitude || null, longitude || null]
    );

    const shelter = await getShelterById(shelterId);
    res.status(201).json(shelter);
  } catch (error: any) {
    console.error('创建避难场所失败:', error);
    if (error.message?.includes('UNIQUE constraint failed')) {
      return res.status(409).json({ error: '场所名称已存在' });
    }
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const shelter = await getShelterById(id);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    const { name, location, type, area, capacity, facilities, manager, status, latitude, longitude, occupied } = req.body;

    if (capacity !== undefined && capacity <= 0) {
      return res.status(400).json({ error: '可容纳人数必须为正整数' });
    }

    if (status && !isValidStatus(status)) {
      return res.status(400).json({ error: '状态无效，只能是 AVAILABLE, MAINTENANCE, FULL, UNAVAILABLE' });
    }

    if (name && name !== shelter.name) {
      const existing = await runQuery<Shelter[]>('SELECT id FROM shelters WHERE name = ? AND id != ?', [name, id]);
      if (existing.length > 0) {
        return res.status(409).json({ error: '场所名称已存在' });
      }
    }

    const finalName = name || shelter.name;
    const finalLocation = location || shelter.location;
    const finalType = type || shelter.type;
    const finalArea = area !== undefined ? area : shelter.area;
    const finalCapacity = capacity !== undefined ? capacity : shelter.capacity;
    const finalOccupied = occupied !== undefined ? occupied : shelter.occupied;
    const finalFacilities = facilities || shelter.facilities;
    const finalManager = manager || shelter.manager;
    const finalStatus = status || shelter.status;
    const finalLat = latitude !== undefined ? latitude : shelter.latitude;
    const finalLng = longitude !== undefined ? longitude : shelter.longitude;

    if (finalOccupied > finalCapacity) {
      return res.status(400).json({ error: '已占用人数不能超过可容纳人数' });
    }

    await runUpdate(
      `UPDATE shelters SET name = ?, location = ?, type = ?, area = ?, capacity = ?, occupied = ?, facilities = ?, manager = ?, status = ?, latitude = ?, longitude = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [finalName, finalLocation, finalType, finalArea, finalCapacity, finalOccupied, finalFacilities, finalManager, finalStatus, finalLat, finalLng, id]
    );

    const updated = await getShelterById(id);
    res.json(updated);
  } catch (error: any) {
    console.error('更新避难场所失败:', error);
    if (error.message?.includes('UNIQUE constraint failed')) {
      return res.status(409).json({ error: '场所名称已存在' });
    }
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const shelter = await getShelterById(id);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    await runUpdate('DELETE FROM shelters WHERE id = ?', [id]);
    res.status(204).send();
  } catch (error) {
    console.error('删除避难场所失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:id/assign', async (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id);
    if (isNaN(id)) {
      return res.status(404).json({ error: 'ID 无效' });
    }

    const shelter = await getShelterById(id);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    if (shelter.status === ShelterStatus.MAINTENANCE) {
      return res.status(400).json({ error: '维护中的避难场所不能被选为疏散目标' });
    }

    if (shelter.status === ShelterStatus.UNAVAILABLE) {
      return res.status(400).json({ error: '不可用的避难场所不能被选为疏散目标' });
    }

    const { peopleCount, reason } = req.body;
    if (!peopleCount || peopleCount <= 0 || !Number.isInteger(peopleCount)) {
      return res.status(400).json({ error: '疏散人数必须为正整数' });
    }

    const remaining = shelter.capacity - shelter.occupied;
    if (peopleCount > remaining) {
      return res.status(400).json({ error: `容量不足，剩余容量: ${remaining}，请求人数: ${peopleCount}` });
    }

    const newOccupied = shelter.occupied + peopleCount;
    const occupancyRate = (newOccupied / shelter.capacity) * 100;
    let newStatus = shelter.status;

    if (newOccupied >= shelter.capacity) {
      newStatus = ShelterStatus.FULL;
    }

    await runUpdate(
      `UPDATE shelters SET occupied = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
      [newOccupied, newStatus, id]
    );

    await runInsert(
      `INSERT INTO assignments (shelter_id, people_count, reason) VALUES (?, ?, ?)`,
      [id, peopleCount, reason || null]
    );

    const updated = await getShelterById(id);
    const response: any = {
      shelter: {
        ...updated,
        remaining: shelter.capacity - newOccupied,
        occupancyRate: Math.round(occupancyRate)
      }
    };

    if (occupancyRate >= 90 && occupancyRate < 100) {
      response.warning = '即将满员';
      const nearShelters = await runQuery<Shelter[]>(
        `SELECT * FROM shelters 
         WHERE id != ? AND status = 'AVAILABLE' AND occupied < capacity 
         ORDER BY capacity DESC LIMIT 3`,
        [id]
      );
      if (nearShelters.length > 0) {
        response.recommendedAlternatives = nearShelters.map(s => ({
          id: s.id,
          name: s.name,
          remaining: s.capacity - s.occupied
        }));
      }
    }

    res.json(response);
  } catch (error) {
    console.error('分配人数失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id/materials', async (req: Request, res: Response) => {
  try {
    const shelterId = parseInt(req.params.id);
    if (isNaN(shelterId)) {
      return res.status(404).json({ error: '避难场所ID无效' });
    }

    const shelter = await getShelterById(shelterId);
    if (!shelter) {
      return res.status(404).json({ error: '避难场所不存在' });
    }

    const materials = await runQuery<any[]>(
      `SELECT * FROM materials WHERE shelter_id = ? ORDER BY name`,
      [shelterId]
    );

    const now = new Date();
    const thirtyDaysLater = new Date();
    thirtyDaysLater.setDate(thirtyDaysLater.getDate() + 30);

    const result = materials.map(material => {
      const expiryDate = new Date(material.expiry_date);
      let expiryStatus = 'normal';
      if (expiryDate < now) {
        expiryStatus = 'expired';
      } else if (expiryDate <= thirtyDaysLater) {
        expiryStatus = 'near_expiry';
      }

      return {
        ...material,
        expiryStatus,
        isLowStock: material.quantity < material.min_stock
      };
    });

    res.json(result);
  } catch (error) {
    console.error('获取物资列表失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
