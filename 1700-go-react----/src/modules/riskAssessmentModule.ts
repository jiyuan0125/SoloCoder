import { RiskLevel, InitialMetrics, FollowUpMetrics } from '../models/types';

export const riskLevelConfig = {
  low: {
    intervalDays: 90,
    description: '低风险：每3个月随访一次',
  },
  medium: {
    intervalDays: 30,
    description: '中风险：每月随访一次',
  },
  high: {
    intervalDays: 14,
    description: '高风险：每两周随访一次',
  },
};

export const controlThresholds = {
  systolicBPMax: 140,
  diastolicBPMax: 90,
  systolicBPMin: 60,
  systolicBPAbsoluteMax: 250,
  fastingBloodSugarMax: 7.0,
};

export function getFollowUpInterval(riskLevel: RiskLevel): number {
  return riskLevelConfig[riskLevel].intervalDays;
}

export function determineInitialRisk(metrics: InitialMetrics): RiskLevel {
  const { systolicBP, diastolicBP, fastingBloodSugar } = metrics;

  let highRiskIndicators = 0;
  let mediumRiskIndicators = 0;

  if (systolicBP > controlThresholds.systolicBPMax) {
    highRiskIndicators++;
  } else if (systolicBP > 130) {
    mediumRiskIndicators++;
  }

  if (diastolicBP > controlThresholds.diastolicBPMax) {
    highRiskIndicators++;
  } else if (diastolicBP > 85) {
    mediumRiskIndicators++;
  }

  if (fastingBloodSugar > controlThresholds.fastingBloodSugarMax) {
    highRiskIndicators++;
  } else if (fastingBloodSugar > 6.1) {
    mediumRiskIndicators++;
  }

  if (highRiskIndicators >= 2) {
    return 'high';
  } else if (highRiskIndicators === 1 || mediumRiskIndicators >= 2) {
    return 'medium';
  }

  return 'low';
}

export function isBloodPressureControlled(metrics: FollowUpMetrics): boolean {
  return (
    metrics.systolicBP <= controlThresholds.systolicBPMax &&
    metrics.diastolicBP <= controlThresholds.diastolicBPMax
  );
}

export function isBloodSugarControlled(metrics: FollowUpMetrics): boolean {
  return metrics.fastingBloodSugar <= controlThresholds.fastingBloodSugarMax;
}

export function assessControlStatus(
  currentMetrics: FollowUpMetrics,
  previousMetrics: FollowUpMetrics | null
): {
  bloodPressureControlled: boolean;
  bloodSugarControlled: boolean;
  overallControlled: boolean;
  needsRiskUpgrade: boolean;
  consecutiveAbnormalBP: boolean;
  consecutiveAbnormalBS: boolean;
} {
  const currentBPControlled = isBloodPressureControlled(currentMetrics);
  const currentBSControlled = isBloodSugarControlled(currentMetrics);

  let consecutiveAbnormalBP = false;
  let consecutiveAbnormalBS = false;

  if (previousMetrics) {
    const previousBPControlled = isBloodPressureControlled(previousMetrics);
    const previousBSControlled = isBloodSugarControlled(previousMetrics);

    consecutiveAbnormalBP = !currentBPControlled && !previousBPControlled;
    consecutiveAbnormalBS = !currentBSControlled && !previousBSControlled;
  }

  const needsRiskUpgrade = consecutiveAbnormalBP || consecutiveAbnormalBS;

  return {
    bloodPressureControlled: currentBPControlled,
    bloodSugarControlled: currentBSControlled,
    overallControlled: currentBPControlled && currentBSControlled,
    needsRiskUpgrade,
    consecutiveAbnormalBP,
    consecutiveAbnormalBS,
  };
}

export function calculateNextRiskLevel(
  currentRisk: RiskLevel,
  needsUpgrade: boolean
): RiskLevel {
  if (!needsUpgrade) {
    return currentRisk;
  }

  const riskOrder: RiskLevel[] = ['low', 'medium', 'high'];
  const currentIndex = riskOrder.indexOf(currentRisk);

  if (currentIndex < riskOrder.length - 1) {
    return riskOrder[currentIndex + 1];
  }

  return currentRisk;
}

export function validateFollowUpMetrics(metrics: FollowUpMetrics): {
  valid: boolean;
  errors: string[];
} {
  const errors: string[] = [];

  if (metrics.systolicBP < controlThresholds.systolicBPMin || 
      metrics.systolicBP > controlThresholds.systolicBPAbsoluteMax) {
    errors.push(`收缩压必须在 ${controlThresholds.systolicBPMin}-${controlThresholds.systolicBPAbsoluteMax} mmHg 范围内`);
  }

  if (metrics.fastingBloodSugar < 0) {
    errors.push('空腹血糖不能为负数');
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}
