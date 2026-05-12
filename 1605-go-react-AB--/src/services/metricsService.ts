import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { ExperimentStatus, MetricComparison } from '../types';
import { getExperimentById, getMetricsForExperiment } from './experimentService';
import { calculateMean, calculateTTest } from '../utils/statistics';

export const recordMetricData = (
  experimentId: string,
  metricId: string,
  variantId: string,
  userKey: string,
  value: number
): void => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status === ExperimentStatus.ENDED) {
    return;
  }

  db.prepare(`
    INSERT INTO metric_data (id, experiment_id, variant_id, metric_id, user_key, value, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(
    uuidv4(),
    experimentId,
    variantId,
    metricId,
    userKey,
    value,
    new Date().toISOString()
  );
};

export const getMetricComparisons = (experimentId: string): MetricComparison[] => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  const metrics = getMetricsForExperiment(experimentId);
  const variants = experiment.variants;
  const controlVariant = variants.find(v => v.isControl);
  
  if (!controlVariant) {
    throw new Error('Control variant not found');
  }

  const treatmentVariants = variants.filter(v => !v.isControl);
  
  const comparisons: MetricComparison[] = [];

  for (const metric of metrics) {
    for (const treatmentVariant of treatmentVariants) {
      const comparison = calculateMetricComparison(
        experimentId,
        metric.id,
        metric.name,
        metric.isCore,
        controlVariant.id,
        treatmentVariant.id,
        treatmentVariant.name
      );
      
      comparisons.push(comparison);
    }
  }

  return comparisons;
};

const calculateMetricComparison = (
  experimentId: string,
  metricId: string,
  metricName: string,
  isCore: boolean,
  controlVariantId: string,
  treatmentVariantId: string,
  treatmentVariantName: string
): MetricComparison => {
  const controlRows = db.prepare(`
    SELECT value, user_key
    FROM metric_data
    WHERE experiment_id = ? AND metric_id = ? AND variant_id = ?
  `).all(experimentId, metricId, controlVariantId) as any[];

  const treatmentRows = db.prepare(`
    SELECT value, user_key
    FROM metric_data
    WHERE experiment_id = ? AND metric_id = ? AND variant_id = ?
  `).all(experimentId, metricId, treatmentVariantId) as any[];

  const controlValues = controlRows.map((r: any) => r.value);
  const treatmentValues = treatmentRows.map((r: any) => r.value);

  const controlMean = calculateMean(controlValues);
  const treatmentMean = calculateMean(treatmentValues);
  const controlSampleSize = controlValues.length;
  const treatmentSampleSize = treatmentValues.length;

  const absoluteDifference = treatmentMean - controlMean;
  const relativeImprovement = controlMean !== 0 ? (absoluteDifference / controlMean) * 100 : null;

  const tTestResult = calculateTTest(controlValues, treatmentValues);
  
  let pValue: number | null = null;
  let sampleSizeWarning: string | undefined;

  if (tTestResult) {
    pValue = tTestResult.pValue;
  } else {
    pValue = null;
    sampleSizeWarning = '样本量不足';
  }

  return {
    metricId,
    metricName,
    control: {
      mean: controlMean,
      sampleSize: controlSampleSize
    },
    treatment: {
      mean: treatmentMean,
      sampleSize: treatmentSampleSize,
      variantId: treatmentVariantId,
      variantName: treatmentVariantName
    },
    absoluteDifference,
    relativeImprovement,
    pValue,
    isCore,
    sampleSizeWarning
  };
};

export const checkCoreMetrics = (experimentId: string): { shouldPause: boolean; reason?: string } => {
  const comparisons = getMetricComparisons(experimentId);
  const coreMetrics = comparisons.filter(c => c.isCore);

  for (const metric of coreMetrics) {
    if (metric.relativeImprovement !== null && metric.relativeImprovement < -5) {
      return {
        shouldPause: true,
        reason: `核心指标 ${metric.metricName} 下降超过 5% (下降 ${Math.abs(metric.relativeImprovement).toFixed(2)}%)`
      };
    }
  }

  return { shouldPause: false };
};
