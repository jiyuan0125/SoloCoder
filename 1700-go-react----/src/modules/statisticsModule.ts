import { Patient, FollowUpRecord, ChronicDiseaseType, ControlRate, FollowUpCompletionRate } from '../models/types';
import { getAllPatients } from './patientModule';
import { getAllFollowUpRecords, getFollowUpRecordsByPatient } from './followUpModule';
import { isBloodPressureControlled, isBloodSugarControlled } from './riskAssessmentModule';

export function calculateControlRates(period: 'month' | 'quarter' | 'year' = 'month'): ControlRate {
  const allPatients = getAllPatients();
  const totalPatients = allPatients.length;

  let hypertensionControlled = 0;
  let diabetesControlled = 0;
  let overallControlled = 0;

  allPatients.forEach(patient => {
    const followUps = getFollowUpRecordsByPatient(patient.id);
    const latestMetrics = followUps.length > 0 
      ? followUps[followUps.length - 1].metrics 
      : null;

    if (!latestMetrics) return;

    const isHypertensionPatient = patient.chronicDiseases.includes('hypertension');
    const isDiabetesPatient = patient.chronicDiseases.includes('diabetes');

    let bpControlled = true;
    let bsControlled = true;

    if (isHypertensionPatient) {
      bpControlled = isBloodPressureControlled(latestMetrics);
      if (bpControlled) hypertensionControlled++;
    }

    if (isDiabetesPatient) {
      bsControlled = isBloodSugarControlled(latestMetrics);
      if (bsControlled) diabetesControlled++;
    }

    const hasHypertensionOrDiabetes = isHypertensionPatient || isDiabetesPatient;
    if (hasHypertensionOrDiabetes && bpControlled && bsControlled) {
      overallControlled++;
    } else if (!hasHypertensionOrDiabetes) {
      overallControlled++;
    }
  });

  const hypertensionPatients = allPatients.filter(p => p.chronicDiseases.includes('hypertension')).length;
  const diabetesPatients = allPatients.filter(p => p.chronicDiseases.includes('diabetes')).length;

  const overallRate = totalPatients > 0 ? (overallControlled / totalPatients) * 100 : 0;
  const hypertensionRate = hypertensionPatients > 0 ? (hypertensionControlled / hypertensionPatients) * 100 : 0;
  const diabetesRate = diabetesPatients > 0 ? (diabetesControlled / diabetesPatients) * 100 : 0;

  return {
    hypertension: Math.round(hypertensionRate * 100) / 100,
    diabetes: Math.round(diabetesRate * 100) / 100,
    overall: Math.round(overallRate * 100) / 100,
    totalPatients,
    controlledPatients: overallControlled,
    period,
    calculatedAt: new Date(),
  };
}

export function calculateFollowUpCompletionRate(
  period: 'month' | 'quarter' | 'year' = 'month'
): FollowUpCompletionRate {
  const now = new Date();
  let startDate: Date;

  switch (period) {
    case 'month':
      startDate = new Date(now.getFullYear(), now.getMonth(), 1);
      break;
    case 'quarter':
      const quarter = Math.floor(now.getMonth() / 3);
      startDate = new Date(now.getFullYear(), quarter * 3, 1);
      break;
    case 'year':
      startDate = new Date(now.getFullYear(), 0, 1);
      break;
  }

  const allPatients = getAllPatients();
  const allFollowUps = getAllFollowUpRecords();

  let totalExpected = 0;
  let completed = 0;

  allPatients.forEach(patient => {
    const patientFollowUps = allFollowUps.filter(f => f.patientId === patient.id);

    let currentDate = new Date(Math.max(startDate.getTime(), patient.createdAt.getTime()));
    const endDate = new Date(now);

    while (currentDate <= endDate) {
      totalExpected++;

      const nextDate = new Date(currentDate);
      nextDate.setDate(nextDate.getDate() + patient.followUpIntervalDays);

      const hasFollowUp = patientFollowUps.some(f => {
        const followUpDate = new Date(f.date);
        return followUpDate >= currentDate && followUpDate <= nextDate;
      });

      if (hasFollowUp) {
        completed++;
      }

      currentDate = nextDate;
    }
  });

  const rate = totalExpected > 0 ? (completed / totalExpected) * 100 : 0;

  return {
    period,
    totalExpected,
    completed,
    rate: Math.round(rate * 100) / 100,
    calculatedAt: new Date(),
  };
}

export function getPatientStatistics(patientId: string): {
  totalFollowUps: number;
  controlledCount: number;
  uncontrolledCount: number;
  adherenceRate: number;
  averageSystolicBP: number;
  averageDiastolicBP: number;
  averageBloodSugar: number;
  recentTrend: 'improving' | 'stable' | 'worsening' | 'insufficient_data';
} | null {
  const followUps = getFollowUpRecordsByPatient(patientId);
  if (followUps.length === 0) return null;

  const patientFollowUps = followUps;
  const lastThree = patientFollowUps.slice(-3);

  let controlledCount = 0;
  let uncontrolledCount = 0;
  let totalSystolic = 0;
  let totalDiastolic = 0;
  let totalBloodSugar = 0;
  let medicationTakenCount = 0;

  patientFollowUps.forEach(f => {
    const bpControlled = f.metrics.systolicBP <= 140 && f.metrics.diastolicBP <= 90;
    const bsControlled = f.metrics.fastingBloodSugar <= 7.0;

    if (bpControlled && bsControlled) {
      controlledCount++;
    } else {
      uncontrolledCount++;
    }

    totalSystolic += f.metrics.systolicBP;
    totalDiastolic += f.metrics.diastolicBP;
    totalBloodSugar += f.metrics.fastingBloodSugar;

    if (f.medicationTaken) {
      medicationTakenCount++;
    }
  });

  let recentTrend: 'improving' | 'stable' | 'worsening' | 'insufficient_data' = 'insufficient_data';
  if (lastThree.length >= 2) {
    const firstControlled = lastThree[0].metrics.systolicBP <= 140 && 
                            lastThree[0].metrics.diastolicBP <= 90 && 
                            lastThree[0].metrics.fastingBloodSugar <= 7.0;
    const lastControlled = lastThree[lastThree.length - 1].metrics.systolicBP <= 140 && 
                          lastThree[lastThree.length - 1].metrics.diastolicBP <= 90 && 
                          lastThree[lastThree.length - 1].metrics.fastingBloodSugar <= 7.0;

    if (!firstControlled && lastControlled) {
      recentTrend = 'improving';
    } else if (firstControlled && !lastControlled) {
      recentTrend = 'worsening';
    } else {
      recentTrend = 'stable';
    }
  }

  return {
    totalFollowUps: patientFollowUps.length,
    controlledCount,
    uncontrolledCount,
    adherenceRate: Math.round((medicationTakenCount / patientFollowUps.length) * 10000) / 100,
    averageSystolicBP: Math.round(totalSystolic / patientFollowUps.length * 10) / 10,
    averageDiastolicBP: Math.round(totalDiastolic / patientFollowUps.length * 10) / 10,
    averageBloodSugar: Math.round(totalBloodSugar / patientFollowUps.length * 10) / 10,
    recentTrend,
  };
}

export function getDiseaseStatistics(diseaseType: ChronicDiseaseType): {
  totalPatients: number;
  controlledPatients: number;
  uncontrolledPatients: number;
  controlRate: number;
  lowRisk: number;
  mediumRisk: number;
  highRisk: number;
} {
  const allPatients = getAllPatients();
  const diseasePatients = allPatients.filter(p => p.chronicDiseases.includes(diseaseType));

  let controlled = 0;
  let uncontrolled = 0;
  let lowRisk = 0;
  let mediumRisk = 0;
  let highRisk = 0;

  diseasePatients.forEach(patient => {
    const followUps = getFollowUpRecordsByPatient(patient.id);
    const latestMetrics = followUps.length > 0 
      ? followUps[followUps.length - 1].metrics 
      : patient.initialMetrics;

    const isControlled = diseaseType === 'hypertension'
      ? latestMetrics.systolicBP <= 140 && latestMetrics.diastolicBP <= 90
      : latestMetrics.fastingBloodSugar <= 7.0;

    if (isControlled) {
      controlled++;
    } else {
      uncontrolled++;
    }

    switch (patient.riskLevel) {
      case 'low': lowRisk++; break;
      case 'medium': mediumRisk++; break;
      case 'high': highRisk++; break;
    }
  });

  const total = diseasePatients.length;
  const controlRate = total > 0 ? (controlled / total) * 100 : 0;

  return {
    totalPatients: total,
    controlledPatients: controlled,
    uncontrolledPatients: uncontrolled,
    controlRate: Math.round(controlRate * 100) / 100,
    lowRisk,
    mediumRisk,
    highRisk,
  };
}
