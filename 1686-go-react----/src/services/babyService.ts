import { v4 as uuidv4 } from 'uuid';
import { db } from '../db';
import { Baby, BabyLog, FeedingType } from '../types';
import { getMotherById, formatDate } from './motherService';

const VALID_FEEDING_TYPES: FeedingType[] = ['breast', 'formula', 'mixed'];

export const isValidFeedingType = (type: string): type is FeedingType => {
  return VALID_FEEDING_TYPES.includes(type as FeedingType);
};

export const getBabyById = (id: string): Baby | undefined => {
  const row = db.prepare('SELECT * FROM babies WHERE id = ?').get(id);
  return row as Baby | undefined;
};

export const createBaby = (
  motherId: string,
  name: string,
  birthWeight: number,
  birthLength: number,
  feedingType: string,
  birthDate: string
): Baby => {
  const mother = getMotherById(motherId);
  if (!mother) {
    throw new Error('MOTHER_NOT_FOUND');
  }

  if (!isValidFeedingType(feedingType)) {
    throw new Error('INVALID_FEEDING_TYPE');
  }

  if (birthWeight <= 0) {
    throw new Error('NEGATIVE_WEIGHT');
  }

  if (birthLength <= 0) {
    throw new Error('INVALID_LENGTH');
  }

  const id = uuidv4();

  const insertStmt = db.prepare(`
    INSERT INTO babies (id, mother_id, name, birth_weight, birth_length, feeding_type, birth_date)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `);

  insertStmt.run(id, motherId, name, birthWeight, birthLength, feedingType, birthDate);

  return {
    id,
    mother_id: motherId,
    name,
    birth_weight: birthWeight,
    birth_length: birthLength,
    feeding_type: feedingType as FeedingType,
    birth_date: birthDate
  };
};

export const addBabyLog = (
  babyId: string,
  logDate: string,
  feedingTime?: string,
  feedingAmount?: number,
  diaperChange?: boolean,
  temperature?: number,
  sleepDuration?: number,
  jaundiceIndex?: number,
  currentWeight?: number
): BabyLog => {
  const baby = getBabyById(babyId);
  if (!baby) {
    throw new Error('BABY_NOT_FOUND');
  }

  if (temperature !== undefined && temperature < 0) {
    throw new Error('NEGATIVE_TEMPERATURE');
  }

  if (currentWeight !== undefined && currentWeight < 0) {
    throw new Error('NEGATIVE_WEIGHT');
  }

  const anomalyNotes: string[] = [];

  if (currentWeight !== undefined) {
    const weightLoss = baby.birth_weight - currentWeight;
    const lossPercentage = (weightLoss / baby.birth_weight) * 100;
    if (lossPercentage > 7) {
      anomalyNotes.push(`体重异常: 较出生体重下降${lossPercentage.toFixed(1)}%，超过7%`);
    }
  }

  if (jaundiceIndex !== undefined) {
    if (jaundiceIndex > 15) {
      const prevLogs = db.prepare(`
        SELECT jaundice_index FROM baby_logs 
        WHERE baby_id = ? AND log_date < ? AND jaundice_index IS NOT NULL
        ORDER BY log_date DESC LIMIT 1
      `).get(babyId, logDate) as { jaundice_index: number } | undefined;

      if (prevLogs && prevLogs.jaundice_index > 15) {
        anomalyNotes.push('黄疸指数连续两天超过15，建议转院');
      } else {
        anomalyNotes.push('黄疸指数超过12，偏高，建议通知医生');
      }
    } else if (jaundiceIndex > 12) {
      anomalyNotes.push('黄疸指数超过12，偏高，建议通知医生');
    }
  }

  const id = uuidv4();
  const notes = anomalyNotes.length > 0 ? anomalyNotes.join('; ') : undefined;

  const insertStmt = db.prepare(`
    INSERT INTO baby_logs (id, baby_id, log_date, feeding_time, feeding_amount, diaper_change, 
                          temperature, sleep_duration, jaundice_index, current_weight, anomaly_notes)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  insertStmt.run(
    id, babyId, logDate, feedingTime, feedingAmount, diaperChange ? 1 : 0,
    temperature, sleepDuration, jaundiceIndex, currentWeight, notes
  );

  return {
    id,
    baby_id: babyId,
    log_date: logDate,
    feeding_time: feedingTime,
    feeding_amount: feedingAmount,
    diaper_change: diaperChange,
    temperature,
    sleep_duration: sleepDuration,
    jaundice_index: jaundiceIndex,
    current_weight: currentWeight,
    anomaly_notes: notes
  };
};

export const getBabyLogs = (babyId: string): BabyLog[] => {
  const rows = db.prepare('SELECT * FROM baby_logs WHERE baby_id = ? ORDER BY log_date DESC').all(babyId);
  return rows as BabyLog[];
};

export const getBabiesByMother = (motherId: string): Baby[] => {
  const rows = db.prepare('SELECT * FROM babies WHERE mother_id = ?').all(motherId);
  return rows as Baby[];
};
