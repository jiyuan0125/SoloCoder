import { Router, Request, Response } from 'express';
import { runQuery } from '../database';
import { Shelter, ShelterStatus } from '../types';

const router = Router();

const DISTANCE_WEIGHT = 0.5;
const CAPACITY_WEIGHT = 0.3;
const FACILITY_WEIGHT = 0.2;

const calculateDistance = (
  lat1: number,
  lon1: number,
  lat2: number,
  lon2: number
): number => {
  const R = 6371;
  const dLat = ((lat2 - lat1) * Math.PI) / 180;
  const dLon = ((lon2 - lon1) * Math.PI) / 180;
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLon / 2) *
      Math.sin(dLon / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
};

const countFacilities = (facilities: string): number => {
  if (!facilities) return 0;
  const facilityList = facilities.split(/[,，;；、\s]+/).filter(f => f.trim());
  return facilityList.length;
};

const normalize = (value: number, min: number, max: number): number => {
  if (max - min === 0) return 1;
  return (value - min) / (max - min);
};

router.post('/recommend', async (req: Request, res: Response) => {
  try {
    const { latitude, longitude, peopleCount, preferredFacilities } = req.body;

    if (latitude === undefined || longitude === undefined) {
      return res.status(400).json({ error: '必须提供当前位置的经纬度' });
    }

    if (peopleCount === undefined || peopleCount <= 0 || !Number.isInteger(peopleCount)) {
      return res.status(400).json({ error: '疏散人数必须为正整数' });
    }

    const availableShelters = await runQuery<Shelter[]>(
      `SELECT * FROM shelters 
       WHERE status IN (?, ?) 
       AND (capacity - occupied) >= ?
       ORDER BY id`,
      [ShelterStatus.AVAILABLE, ShelterStatus.FULL === ShelterStatus.FULL ? ShelterStatus.AVAILABLE : ShelterStatus.AVAILABLE, peopleCount]
    );

    const validShelters = availableShelters.filter(
      s => s.status === ShelterStatus.AVAILABLE && (s.capacity - s.occupied) >= peopleCount
    );

    if (validShelters.length === 0) {
      return res.json({
        message: '暂无可用的避难场所',
        recommendations: []
      });
    }

    let maxDistance = 0;
    let minDistance = Infinity;
    let maxRemainingRatio = 0;
    let minRemainingRatio = Infinity;
    let maxFacilityCount = 0;
    let minFacilityCount = Infinity;

    const sheltersWithMetrics = validShelters.map(shelter => {
      const hasCoord = shelter.latitude !== null && shelter.latitude !== undefined &&
                       shelter.longitude !== null && shelter.longitude !== undefined;
      
      const distance = hasCoord
        ? calculateDistance(latitude, longitude, shelter.latitude!, shelter.longitude!)
        : 1000;

      const remaining = shelter.capacity - shelter.occupied;
      const remainingRatio = remaining / shelter.capacity;
      const facilityCount = countFacilities(shelter.facilities);

      maxDistance = Math.max(maxDistance, distance);
      minDistance = Math.min(minDistance, distance);
      maxRemainingRatio = Math.max(maxRemainingRatio, remainingRatio);
      minRemainingRatio = Math.min(minRemainingRatio, remainingRatio);
      maxFacilityCount = Math.max(maxFacilityCount, facilityCount);
      minFacilityCount = Math.min(minFacilityCount, facilityCount);

      const facilityMatchScore = preferredFacilities
        ? preferredFacilities.split(/[,，;；、\s]+/).filter((pref: string) => 
            shelter.facilities.includes(pref.trim())
          ).length
        : 0;

      return {
        shelter,
        distance,
        remaining,
        remainingRatio,
        facilityCount,
        facilityMatchScore
      };
    });

    const scoredShelters = sheltersWithMetrics.map(item => {
      const distanceScore = maxDistance === minDistance 
        ? 0.5 
        : 1 - normalize(item.distance, minDistance, maxDistance);

      const capacityScore = maxRemainingRatio === minRemainingRatio
        ? 0.5
        : normalize(item.remainingRatio, minRemainingRatio, maxRemainingRatio);

      const facilityScore = maxFacilityCount === minFacilityCount
        ? 0.5
        : normalize(item.facilityCount, minFacilityCount, maxFacilityCount);

      const totalScore =
        distanceScore * DISTANCE_WEIGHT +
        capacityScore * CAPACITY_WEIGHT +
        facilityScore * FACILITY_WEIGHT;

      return {
        shelter: {
          id: item.shelter.id,
          name: item.shelter.name,
          location: item.shelter.location,
          type: item.shelter.type,
          area: item.shelter.area,
          capacity: item.shelter.capacity,
          occupied: item.shelter.occupied,
          remaining: item.remaining,
          facilities: item.shelter.facilities,
          manager: item.shelter.manager,
          status: item.shelter.status,
          latitude: item.shelter.latitude,
          longitude: item.shelter.longitude
        },
        distance: item.distance,
        score: {
          total: Math.round(totalScore * 100) / 100,
          distance: Math.round(distanceScore * 100) / 100,
          capacity: Math.round(capacityScore * 100) / 100,
          facility: Math.round(facilityScore * 100) / 100
        }
      };
    });

    scoredShelters.sort((a, b) => b.score.total - a.score.total);

    res.json({
      message: '推荐方案已生成',
      peopleCount,
      weights: {
        distance: DISTANCE_WEIGHT,
        capacity: CAPACITY_WEIGHT,
        facility: FACILITY_WEIGHT
      },
      recommendations: scoredShelters
    });
  } catch (error) {
    console.error('生成疏散推荐失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/status', async (_req: Request, res: Response) => {
  try {
    const allShelters = await runQuery<Shelter[]>('SELECT * FROM shelters');
    
    const total = allShelters.length;
    const available = allShelters.filter(s => s.status === ShelterStatus.AVAILABLE).length;
    const maintenance = allShelters.filter(s => s.status === ShelterStatus.MAINTENANCE).length;
    const full = allShelters.filter(s => s.status === ShelterStatus.FULL).length;
    const unavailable = allShelters.filter(s => s.status === ShelterStatus.UNAVAILABLE).length;
    
    const totalCapacity = allShelters.reduce((sum, s) => sum + s.capacity, 0);
    const totalOccupied = allShelters.reduce((sum, s) => sum + s.occupied, 0);
    const totalRemaining = totalCapacity - totalOccupied;

    res.json({
      summary: {
        total,
        available,
        maintenance,
        full,
        unavailable
      },
      capacity: {
        total: totalCapacity,
        occupied: totalOccupied,
        remaining: totalRemaining,
        utilizationRate: totalCapacity > 0 ? Math.round((totalOccupied / totalCapacity) * 100) : 0
      },
      shelters: allShelters.map(s => ({
        id: s.id,
        name: s.name,
        status: s.status,
        capacity: s.capacity,
        occupied: s.occupied,
        remaining: s.capacity - s.occupied,
        occupancyRate: Math.round((s.occupied / s.capacity) * 100)
      }))
    });
  } catch (error) {
    console.error('获取系统状态失败:', error);
    res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
