import { FollowUpRecord, FollowUpMetrics, LifestyleAssessment, Medication, RiskLevel } from '../models/types';
import {
  getPatient,
  updatePatientRisk,
  updateFollowUpDates,
  updateMedicationAdherence,
  DuplicateFollowUpError,
  PatientValidationError,
} from './patientModule';

export { DuplicateFollowUpError };
import {
  assessControlStatus,
  calculateNextRiskLevel,
  validateFollowUpMetrics,
} from './riskAssessmentModule';

const followUpRecords: Map<string, FollowUpRecord> = new Map();
const patientFollowUpDates: Map<string, Set<string>> = new Map();

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
}

function formatDateKey(date: Date): string {
  return date.toISOString().split('T')[0];
}

function checkDuplicateFollowUp(patientId: string, date: Date): boolean {
  const dates = patientFollowUpDates.get(patientId);
  if (!dates) return false;
  return dates.has(formatDateKey(date));
}

function registerFollowUpDate(patientId: string, date: Date): void {
  if (!patientFollowUpDates.has(patientId)) {
    patientFollowUpDates.set(patientId, new Set());
  }
  patientFollowUpDates.get(patientId)!.add(formatDateKey(date));
}

export interface CreateFollowUpInput {
  patientId: string;
  date: Date;
  symptoms: string[];
  medications: Medication[];
  metrics: FollowUpMetrics;
  lifestyleAssessment: LifestyleAssessment;
  medicationTaken: boolean;
  notes?: string;
}

export interface FollowUpResult {
  record: FollowUpRecord;
  riskLevelChanged: boolean;
  oldRiskLevel: RiskLevel | null;
  newRiskLevel: RiskLevel | null;
  controlStatus: {
    bloodPressureControlled: boolean;
    bloodSugarControlled: boolean;
    overallControlled: boolean;
  };
  adherenceStatusChanged: boolean;
  notificationRequired: boolean;
}

export function createFollowUpRecord(input: CreateFollowUpInput): FollowUpResult {
  const patient = getPatient(input.patientId);
  if (!patient) {
    throw new PatientValidationError('患者不存在');
  }

  const metricsValidation = validateFollowUpMetrics(input.metrics);
  if (!metricsValidation.valid) {
    throw new PatientValidationError(metricsValidation.errors.join('; '));
  }

  if (checkDuplicateFollowUp(input.patientId, input.date)) {
    throw new DuplicateFollowUpError('同一患者同一天不能重复随访');
  }

  const previousFollowUps = getFollowUpRecordsByPatient(input.patientId);
  const previousMetrics = previousFollowUps.length > 0 
    ? previousFollowUps[previousFollowUps.length - 1].metrics 
    : null;

  const controlStatus = assessControlStatus(input.metrics, previousMetrics);

  const oldRiskLevel = patient.riskLevel;
  const nextRiskLevel = calculateNextRiskLevel(oldRiskLevel, controlStatus.needsRiskUpgrade);
  const riskLevelChanged = nextRiskLevel !== oldRiskLevel;

  const id = generateId();
  const record: FollowUpRecord = {
    id,
    patientId: input.patientId,
    date: input.date,
    symptoms: input.symptoms,
    medications: input.medications,
    metrics: input.metrics,
    lifestyleAssessment: input.lifestyleAssessment,
    medicationTaken: input.medicationTaken,
    notes: input.notes,
    createdAt: new Date(),
  };

  followUpRecords.set(id, record);
  registerFollowUpDate(input.patientId, input.date);

  updateFollowUpDates(input.patientId, input.date);

  const updatedPatient = updateMedicationAdherence(input.patientId, input.medicationTaken);
  const adherenceStatusChanged = updatedPatient 
    ? updatedPatient.medicationAdherence !== patient.medicationAdherence 
    : false;
  const notificationRequired = updatedPatient?.medicationAdherence === 'poor' && adherenceStatusChanged;

  if (riskLevelChanged) {
    updatePatientRisk(input.patientId, nextRiskLevel);
  }

  return {
    record,
    riskLevelChanged,
    oldRiskLevel: riskLevelChanged ? oldRiskLevel : null,
    newRiskLevel: riskLevelChanged ? nextRiskLevel : null,
    controlStatus: {
      bloodPressureControlled: controlStatus.bloodPressureControlled,
      bloodSugarControlled: controlStatus.bloodSugarControlled,
      overallControlled: controlStatus.overallControlled,
    },
    adherenceStatusChanged,
    notificationRequired,
  };
}

export function getFollowUpRecord(id: string): FollowUpRecord | undefined {
  return followUpRecords.get(id);
}

export function getFollowUpRecordsByPatient(patientId: string): FollowUpRecord[] {
  const records = Array.from(followUpRecords.values())
    .filter(r => r.patientId === patientId)
    .sort((a, b) => a.date.getTime() - b.date.getTime());
  return records;
}

export function getAllFollowUpRecords(): FollowUpRecord[] {
  return Array.from(followUpRecords.values());
}

export function getLatestFollowUpMetrics(patientId: string): FollowUpMetrics | null {
  const records = getFollowUpRecordsByPatient(patientId);
  if (records.length === 0) return null;
  return records[records.length - 1].metrics;
}

export function getFollowUpRecordsByDateRange(startDate: Date, endDate: Date): FollowUpRecord[] {
  return Array.from(followUpRecords.values())
    .filter(r => r.date >= startDate && r.date <= endDate)
    .sort((a, b) => a.date.getTime() - b.date.getTime());
}
