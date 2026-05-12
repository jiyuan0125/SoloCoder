import { Patient, BasicInfo, InitialMetrics, ChronicDiseaseType, RiskLevel } from '../models/types';
import { riskLevelConfig, determineInitialRisk, getFollowUpInterval } from './riskAssessmentModule';

const patients: Map<string, Patient> = new Map();

export class PatientValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'PatientValidationError';
  }
}

export class DuplicateFollowUpError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'DuplicateFollowUpError';
  }
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
}

function validatePatientData(
  chronicDiseases: ChronicDiseaseType[],
  diagnosisDate: Date
): void {
  if (!chronicDiseases || chronicDiseases.length === 0) {
    throw new PatientValidationError('慢病类型不能为空');
  }

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  if (diagnosisDate > today) {
    throw new PatientValidationError('确诊日期不能晚于当前日期');
  }
}

export function createPatient(
  basicInfo: BasicInfo,
  chronicDiseases: ChronicDiseaseType[],
  diagnosisDate: Date,
  medicalHistory: string[],
  initialMetrics: InitialMetrics
): Patient {
  validatePatientData(chronicDiseases, diagnosisDate);

  const id = generateId();
  const riskLevel = determineInitialRisk(initialMetrics);
  const followUpIntervalDays = getFollowUpInterval(riskLevel);

  const now = new Date();
  const nextFollowUpDate = new Date(now);
  nextFollowUpDate.setDate(nextFollowUpDate.getDate() + followUpIntervalDays);

  const patient: Patient = {
    id,
    basicInfo,
    chronicDiseases,
    diagnosisDate,
    medicalHistory,
    initialMetrics,
    riskLevel,
    followUpIntervalDays,
    nextFollowUpDate,
    createdAt: now,
    updatedAt: now,
    medicationAdherence: 'good',
    nonCompliantCount: 0,
  };

  patients.set(id, patient);
  return patient;
}

export function getPatient(id: string): Patient | undefined {
  return patients.get(id);
}

export function getAllPatients(): Patient[] {
  return Array.from(patients.values());
}

export function updatePatientRisk(
  patientId: string,
  newRiskLevel: RiskLevel,
  isManualDowngrade: boolean = false
): Patient | undefined {
  const patient = patients.get(patientId);
  if (!patient) return undefined;

  const riskOrder: RiskLevel[] = ['low', 'medium', 'high'];
  const currentIndex = riskOrder.indexOf(patient.riskLevel);
  const newIndex = riskOrder.indexOf(newRiskLevel);

  if (!isManualDowngrade && newIndex < currentIndex) {
    throw new PatientValidationError('风险等级只能从低到高升级，人工降级需要明确指定');
  }

  patient.riskLevel = newRiskLevel;
  patient.followUpIntervalDays = getFollowUpInterval(newRiskLevel);

  if (patient.lastFollowUpDate) {
    const nextDate = new Date(patient.lastFollowUpDate);
    nextDate.setDate(nextDate.getDate() + patient.followUpIntervalDays);
    patient.nextFollowUpDate = nextDate;
  }

  patient.updatedAt = new Date();
  patients.set(patientId, patient);
  return patient;
}

export function updatePatient(
  id: string,
  updates: Partial<Pick<Patient, 'basicInfo' | 'chronicDiseases' | 'medicalHistory'>>
): Patient | undefined {
  const patient = patients.get(id);
  if (!patient) return undefined;

  if (updates.basicInfo) {
    patient.basicInfo = { ...patient.basicInfo, ...updates.basicInfo };
  }
  if (updates.chronicDiseases) {
    if (updates.chronicDiseases.length === 0) {
      throw new PatientValidationError('慢病类型不能为空');
    }
    patient.chronicDiseases = updates.chronicDiseases;
  }
  if (updates.medicalHistory) {
    patient.medicalHistory = updates.medicalHistory;
  }

  patient.updatedAt = new Date();
  patients.set(id, patient);
  return patient;
}

export function manuallyDowngradeRisk(patientId: string, newRiskLevel: RiskLevel): Patient | undefined {
  const patient = patients.get(patientId);
  if (!patient) return undefined;

  const riskOrder: RiskLevel[] = ['low', 'medium', 'high'];
  const currentIndex = riskOrder.indexOf(patient.riskLevel);
  const newIndex = riskOrder.indexOf(newRiskLevel);

  if (newIndex >= currentIndex) {
    throw new PatientValidationError('人工降级只能降低风险等级，使用 updatePatientRisk 进行升级');
  }

  return updatePatientRisk(patientId, newRiskLevel, true);
}

export function updateMedicationAdherence(patientId: string, medicationTaken: boolean): Patient | undefined {
  const patient = patients.get(patientId);
  if (!patient) return undefined;

  if (!medicationTaken) {
    patient.nonCompliantCount += 1;
  } else {
    patient.nonCompliantCount = 0;
  }

  if (patient.nonCompliantCount >= 2) {
    patient.medicationAdherence = 'poor';
  } else {
    patient.medicationAdherence = 'good';
  }

  patient.updatedAt = new Date();
  patients.set(patientId, patient);
  return patient;
}

export function updateFollowUpDates(patientId: string, followUpDate: Date): Patient | undefined {
  const patient = patients.get(patientId);
  if (!patient) return undefined;

  patient.lastFollowUpDate = followUpDate;
  const nextDate = new Date(followUpDate);
  nextDate.setDate(nextDate.getDate() + patient.followUpIntervalDays);
  patient.nextFollowUpDate = nextDate;
  patient.updatedAt = new Date();
  patients.set(patientId, patient);
  return patient;
}

export { riskLevelConfig };
