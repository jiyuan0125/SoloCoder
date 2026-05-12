import { db } from '../database';
import { HealthRecord, CreateHealthRecordInput } from '../types';
import { generateId, now } from '../utils';
import { elderService } from './elderService';

const SYSTOLIC_LIMIT = 140;
const DIASTOLIC_LIMIT = 90;
const BLOOD_SUGAR_LIMIT = 7.0;

function mapHealthRecord(row: any): HealthRecord {
  return {
    id: row.id,
    elderId: row.elder_id,
    systolic: row.systolic,
    diastolic: row.diastolic,
    bloodSugar: row.blood_sugar,
    recordDate: row.record_date,
    isAbnormal: row.is_abnormal === 1,
    medicalAdvice: row.medical_advice,
    createdAt: row.created_at,
  };
}

export const healthService = {
  getHealthRecordById(id: string): HealthRecord | null {
    const row = db.prepare('SELECT * FROM health_records WHERE id = ?').get(id);
    return row ? mapHealthRecord(row) : null;
  },

  getHealthRecordsByElderId(elderId: string): HealthRecord[] {
    const rows = db.prepare(
      'SELECT * FROM health_records WHERE elder_id = ? ORDER BY record_date DESC'
    ).all(elderId);
    return rows.map(mapHealthRecord);
  },

  isBloodPressureAbnormal(systolic: number | null, diastolic: number | null): boolean {
    if (systolic === null || diastolic === null) return false;
    return systolic > SYSTOLIC_LIMIT || diastolic > DIASTOLIC_LIMIT;
  },

  isBloodSugarAbnormal(bloodSugar: number | null): boolean {
    if (bloodSugar === null) return false;
    return bloodSugar > BLOOD_SUGAR_LIMIT;
  },

  getRecentRecords(elderId: string, limit: number): HealthRecord[] {
    const rows = db.prepare(`
      SELECT * FROM health_records 
      WHERE elder_id = ? 
      ORDER BY record_date DESC 
      LIMIT ?
    `).all(elderId, limit);
    return rows.map(mapHealthRecord);
  },

  checkConsecutiveAbnormal(
    elderId: string,
    type: 'bloodPressure' | 'bloodSugar',
    consecutiveCount: number = 3
  ): { isAbnormal: boolean; records: HealthRecord[] } {
    const recentRecords = healthService.getRecentRecords(elderId, consecutiveCount);

    if (recentRecords.length < consecutiveCount) {
      return { isAbnormal: false, records: [] };
    }

    let abnormalCount = 0;
    const abnormalRecords: HealthRecord[] = [];

    for (const record of recentRecords) {
      let hasData = false;
      let isAbnormal = false;

      if (type === 'bloodPressure') {
        if (record.systolic !== null && record.diastolic !== null) {
          hasData = true;
          isAbnormal = healthService.isBloodPressureAbnormal(record.systolic, record.diastolic);
        }
      } else {
        if (record.bloodSugar !== null) {
          hasData = true;
          isAbnormal = healthService.isBloodSugarAbnormal(record.bloodSugar);
        }
      }

      if (!hasData) {
        continue;
      }

      if (isAbnormal) {
        abnormalCount++;
        abnormalRecords.push(record);
      } else {
        break;
      }
    }

    return {
      isAbnormal: abnormalCount >= consecutiveCount,
      records: abnormalRecords,
    };
  },

  createHealthRecord(input: CreateHealthRecordInput): HealthRecord | null {
    const elder = elderService.getElderById(input.elderId);
    if (!elder) return null;

    const hasBloodPressure = input.systolic !== null && input.diastolic !== null;
    const hasBloodSugar = input.bloodSugar !== null;

    let isAbnormal = false;
    let medicalAdvice: string | null = null;

    const id = generateId();
    const timestamp = now();
    const recordDate = input.recordDate || timestamp;

    db.prepare(`
      INSERT INTO health_records (
        id, elder_id, systolic, diastolic, blood_sugar, record_date, 
        is_abnormal, medical_advice, created_at
      ) VALUES (?, ?, ?, ?, ?, ?, 0, NULL, ?)
    `).run(
      id,
      input.elderId,
      input.systolic,
      input.diastolic,
      input.bloodSugar,
      recordDate,
      timestamp
    );

    if (hasBloodPressure) {
      const bpCheck = healthService.checkConsecutiveAbnormal(input.elderId, 'bloodPressure', 3);
      if (bpCheck.isAbnormal) {
        isAbnormal = true;
        medicalAdvice = `连续3次血压异常（收缩压>140或舒张压>90），建议就医检查。异常记录日期：${bpCheck.records
          .map((r) => r.recordDate.split('T')[0])
          .join('、')}`;
      }
    }

    if (hasBloodSugar) {
      const sugarCheck = healthService.checkConsecutiveAbnormal(input.elderId, 'bloodSugar', 3);
      if (sugarCheck.isAbnormal) {
        isAbnormal = true;
        const sugarAdvice = `连续3次血糖异常（>7.0），建议就医检查。异常记录日期：${sugarCheck.records
          .map((r) => r.recordDate.split('T')[0])
          .join('、')}`;
        medicalAdvice = medicalAdvice ? `${medicalAdvice}；${sugarAdvice}` : sugarAdvice;
      }
    }

    if (isAbnormal) {
      db.prepare(`
        UPDATE health_records 
        SET is_abnormal = 1, medical_advice = ?
        WHERE id = ?
      `).run(medicalAdvice, id);
    }

    return healthService.getHealthRecordById(id);
  },

  getAbnormalRecords(elderId?: string): HealthRecord[] {
    let query = 'SELECT * FROM health_records WHERE is_abnormal = 1 ORDER BY record_date DESC';
    const params: any[] = [];

    if (elderId) {
      query = 'SELECT * FROM health_records WHERE elder_id = ? AND is_abnormal = 1 ORDER BY record_date DESC';
      params.push(elderId);
    }

    const rows = db.prepare(query).all(...params);
    return rows.map(mapHealthRecord);
  },
};
