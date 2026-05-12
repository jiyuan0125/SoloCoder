import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Experiment, ExperimentStatus, ExperimentVariant, TrafficType, Assignment } from '../types';
import { hashToPercentage } from '../utils/hash';
import { getExperimentById } from './experimentService';

interface AssignRequest {
  experimentId: string;
  userId?: string;
  deviceId?: string;
  region?: string;
}

interface AssignResult {
  experimentId: string;
  experimentName: string;
  variantId: string;
  variantName: string;
  isControl: boolean;
}

const getUserKey = (experiment: Experiment, request: AssignRequest): string => {
  switch (experiment.trafficType) {
    case TrafficType.USER_ID:
      if (!request.userId) {
        throw new Error('userId is required for user_id traffic type');
      }
      return request.userId;
    case TrafficType.DEVICE_ID:
      if (!request.deviceId) {
        throw new Error('deviceId is required for device_id traffic type');
      }
      return request.deviceId;
    case TrafficType.REGION:
      if (!request.region) {
        throw new Error('region is required for region traffic type');
      }
      return request.region;
    default:
      throw new Error('Invalid traffic type');
  }
};

const getExistingAssignment = (
  experimentId: string,
  userKey: string
): Assignment | null => {
  const row = db.prepare(`
    SELECT a.id, a.experiment_id as experimentId, a.user_key as userKey, 
           a.variant_id as variantId, v.name as variantName, a.created_at as createdAt
    FROM assignments a
    JOIN variants v ON a.variant_id = v.id
    WHERE a.experiment_id = ? AND a.user_key = ?
  `).get(experimentId, userKey) as any;

  if (!row) return null;

  return {
    ...row,
    createdAt: new Date(row.createdAt)
  };
};

const selectVariantByHash = (
  variants: ExperimentVariant[],
  userKey: string,
  experimentId: string
): ExperimentVariant => {
  const hashValue = hashToPercentage(userKey, experimentId);
  
  let cumulative = 0;
  for (const variant of variants) {
    cumulative += variant.trafficPercentage;
    if (hashValue < cumulative) {
      return variant;
    }
  }

  return variants[variants.length - 1];
};

export const assignUser = (request: AssignRequest): AssignResult | null => {
  const experiment = getExperimentById(request.experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status === ExperimentStatus.ENDED) {
    return null;
  }

  if (experiment.status === ExperimentStatus.CONFIGURING) {
    throw new Error('Experiment is not yet running');
  }

  const userKey = getUserKey(experiment, request);

  const existingAssignment = getExistingAssignment(experiment.id, userKey);
  if (existingAssignment) {
    const variant = experiment.variants.find(v => v.id === existingAssignment.variantId);
    if (variant) {
      return {
        experimentId: experiment.id,
        experimentName: experiment.name,
        variantId: variant.id,
        variantName: variant.name,
        isControl: variant.isControl
      };
    }
  }

  const selectedVariant = selectVariantByHash(experiment.variants, userKey, experiment.id);

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT OR IGNORE INTO assignments (id, experiment_id, user_key, variant_id, created_at)
      VALUES (?, ?, ?, ?, ?)
    `).run(
      uuidv4(),
      experiment.id,
      userKey,
      selectedVariant.id,
      new Date().toISOString()
    );
  });

  transaction();

  return {
    experimentId: experiment.id,
    experimentName: experiment.name,
    variantId: selectedVariant.id,
    variantName: selectedVariant.name,
    isControl: selectedVariant.isControl
  };
};

export const getAssignmentForUser = (
  experimentId: string,
  request: AssignRequest
): Assignment | null => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) return null;

  const userKey = getUserKey(experiment, request);
  return getExistingAssignment(experimentId, userKey);
};
