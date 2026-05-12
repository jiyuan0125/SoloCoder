import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { ExperimentStatus, GrayscaleConfig } from '../types';
import { getExperimentById, updateExperimentStatus } from './experimentService';
import { checkCoreMetrics } from './metricsService';

export const createGrayscaleConfig = (
  experimentId: string,
  steps: number[]
): GrayscaleConfig => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status !== ExperimentStatus.CONFIGURING) {
    throw new Error('Grayscale can only be configured during configuring state');
  }

  const sortedSteps = [...steps].sort((a, b) => a - b);
  
  if (sortedSteps.length === 0) {
    throw new Error('Steps cannot be empty');
  }

  if (sortedSteps[sortedSteps.length - 1] !== 100) {
    sortedSteps.push(100);
  }

  for (let i = 0; i < sortedSteps.length; i++) {
    if (sortedSteps[i] <= 0 || sortedSteps[i] > 100) {
      throw new Error('Each step must be between 0 and 100');
    }
    if (i > 0 && sortedSteps[i] <= sortedSteps[i - 1]) {
      throw new Error('Steps must be strictly increasing');
    }
  }

  const existing = db.prepare(`
    SELECT id FROM grayscale_configs WHERE experiment_id = ?
  `).get(experimentId);

  if (existing) {
    db.prepare(`
      UPDATE grayscale_configs
      SET steps = ?, current_step_index = 0, is_active = 0
      WHERE experiment_id = ?
    `).run(JSON.stringify(sortedSteps), experimentId);
  } else {
    db.prepare(`
      INSERT INTO grayscale_configs (id, experiment_id, steps, current_step_index, is_active)
      VALUES (?, ?, ?, 0, 0)
    `).run(uuidv4(), experimentId, JSON.stringify(sortedSteps));
  }

  return getGrayscaleConfig(experimentId)!;
};

export const getGrayscaleConfig = (experimentId: string): GrayscaleConfig | null => {
  const row = db.prepare(`
    SELECT id, experiment_id as experimentId, steps, current_step_index, is_active as isActive
    FROM grayscale_configs
    WHERE experiment_id = ?
  `).get(experimentId) as any;

  if (!row) return null;

  return {
    id: row.id,
    experimentId: row.experimentId,
    steps: JSON.parse(row.steps),
    currentStep: row.current_step_index,
    isActive: row.isActive === 1
  };
};

export const startGrayscale = (experimentId: string): GrayscaleConfig => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status !== ExperimentStatus.RUNNING) {
    throw new Error('Experiment must be running to start grayscale');
  }

  const config = getGrayscaleConfig(experimentId);
  if (!config) {
    throw new Error('Grayscale config not found');
  }

  db.prepare(`
    UPDATE grayscale_configs
    SET is_active = 1
    WHERE experiment_id = ?
  `).run(experimentId);

  return getGrayscaleConfig(experimentId)!;
};

export const advanceGrayscaleStep = (experimentId: string): { paused: boolean; reason?: string } => {
  const config = getGrayscaleConfig(experimentId);
  if (!config) {
    throw new Error('Grayscale config not found');
  }

  if (!config.isActive) {
    throw new Error('Grayscale is not active');
  }

  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status !== ExperimentStatus.RUNNING) {
    throw new Error('Experiment must be running to advance grayscale');
  }

  if (config.currentStep >= config.steps.length - 1) {
    db.prepare(`
      UPDATE grayscale_configs
      SET is_active = 0
      WHERE experiment_id = ?
    `).run(experimentId);
    
    return { paused: false, reason: 'Grayscale completed' };
  }

  const coreCheck = checkCoreMetrics(experimentId);
  
  if (coreCheck.shouldPause) {
    updateExperimentStatus(experimentId, ExperimentStatus.PAUSED);
    db.prepare(`
      UPDATE grayscale_configs
      SET is_active = 0
      WHERE experiment_id = ?
    `).run(experimentId);
    
    return { paused: true, reason: coreCheck.reason };
  }

  const nextStep = config.currentStep + 1;
  const targetPercentage = config.steps[nextStep];

  adjustVariantPercentages(experimentId, targetPercentage);

  db.prepare(`
    UPDATE grayscale_configs
    SET current_step_index = ?
    WHERE experiment_id = ?
  `).run(nextStep, experimentId);

  if (nextStep >= config.steps.length - 1) {
    db.prepare(`
      UPDATE grayscale_configs
      SET is_active = 0
      WHERE experiment_id = ?
    `).run(experimentId);
  }

  return { paused: false };
};

const adjustVariantPercentages = (experimentId: string, targetPercentage: number): void => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) return;

  const controlVariant = experiment.variants.find(v => v.isControl);
  if (!controlVariant) return;

  const treatmentVariants = experiment.variants.filter(v => !v.isControl);
  if (treatmentVariants.length === 0) return;

  const totalTreatmentPercentage = targetPercentage;
  const controlPercentage = 100 - totalTreatmentPercentage;
  const treatmentPercentagePerVariant = totalTreatmentPercentage / treatmentVariants.length;

  const transaction = db.transaction(() => {
    db.prepare(`
      UPDATE variants
      SET traffic_percentage = ?
      WHERE experiment_id = ? AND is_control = 1
    `).run(controlPercentage, experimentId);

    for (const variant of treatmentVariants) {
      db.prepare(`
        UPDATE variants
        SET traffic_percentage = ?
        WHERE experiment_id = ? AND id = ?
      `).run(treatmentPercentagePerVariant, experimentId, variant.id);
    }

    db.prepare(`
      UPDATE experiments
      SET updated_at = ?
      WHERE id = ?
    `).run(new Date().toISOString(), experimentId);
  });

  transaction();
};

export const pauseGrayscale = (experimentId: string): void => {
  const config = getGrayscaleConfig(experimentId);
  if (!config) return;

  db.prepare(`
    UPDATE grayscale_configs
    SET is_active = 0
    WHERE experiment_id = ?
  `).run(experimentId);
};

export const resumeGrayscale = (experimentId: string): GrayscaleConfig => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status !== ExperimentStatus.RUNNING) {
    throw new Error('Experiment must be running to resume grayscale');
  }

  const config = getGrayscaleConfig(experimentId);
  if (!config) {
    throw new Error('Grayscale config not found');
  }

  if (config.currentStep >= config.steps.length - 1) {
    throw new Error('Grayscale already completed');
  }

  db.prepare(`
    UPDATE grayscale_configs
    SET is_active = 1
    WHERE experiment_id = ?
  `).run(experimentId);

  return getGrayscaleConfig(experimentId)!;
};
