import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Experiment, ExperimentStatus, TrafficType, ExperimentVariant, Metric } from '../types';

const getVariantsByExperimentId = (experimentId: string): ExperimentVariant[] => {
  const rows = db.prepare(`
    SELECT id, name, traffic_percentage as trafficPercentage, is_control as isControl
    FROM variants
    WHERE experiment_id = ?
  `).all(experimentId) as any[];

  return rows.map(row => ({
    ...row,
    isControl: row.isControl === 1
  }));
};

const getMetricsByExperimentId = (experimentId: string): Metric[] => {
  const rows = db.prepare(`
    SELECT id, experiment_id as experimentId, name, is_core as isCore
    FROM metrics
    WHERE experiment_id = ?
  `).all(experimentId) as any[];

  return rows.map(row => ({
    ...row,
    isCore: row.isCore === 1
  }));
};

const rowToExperiment = (row: any): Experiment => {
  return {
    id: row.id,
    name: row.name,
    status: row.status as ExperimentStatus,
    trafficType: row.traffic_type as TrafficType,
    variants: getVariantsByExperimentId(row.id),
    createdAt: new Date(row.created_at),
    updatedAt: new Date(row.updated_at)
  };
};

export const createExperiment = (
  name: string,
  trafficType: TrafficType,
  variants: Omit<ExperimentVariant, 'id'>[],
  metrics: Omit<Metric, 'id' | 'experimentId'>[]
): Experiment => {
  const totalPercentage = variants.reduce((sum, v) => sum + v.trafficPercentage, 0);
  if (Math.abs(totalPercentage - 100) > 0.001) {
    throw new Error('Traffic percentages must sum to 100%');
  }

  const controlVariant = variants.find(v => v.isControl);
  if (!controlVariant) {
    throw new Error('At least one variant must be the control group');
  }

  const transaction = db.transaction(() => {
    const experimentId = uuidv4();
    const now = new Date().toISOString();

    db.prepare(`
      INSERT INTO experiments (id, name, status, traffic_type, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(experimentId, name, ExperimentStatus.CONFIGURING, trafficType, now, now);

    const insertVariant = db.prepare(`
      INSERT INTO variants (id, experiment_id, name, traffic_percentage, is_control)
      VALUES (?, ?, ?, ?, ?)
    `);

    for (const variant of variants) {
      insertVariant.run(
        uuidv4(),
        experimentId,
        variant.name,
        variant.trafficPercentage,
        variant.isControl ? 1 : 0
      );
    }

    const insertMetric = db.prepare(`
      INSERT INTO metrics (id, experiment_id, name, is_core)
      VALUES (?, ?, ?, ?)
    `);

    for (const metric of metrics) {
      insertMetric.run(
        uuidv4(),
        experimentId,
        metric.name,
        metric.isCore ? 1 : 0
      );
    }

    return experimentId;
  });

  const experimentId = transaction();
  return getExperimentById(experimentId)!;
};

export const getExperimentById = (id: string): Experiment | null => {
  const row = db.prepare(`
    SELECT id, name, status, traffic_type, created_at, updated_at
    FROM experiments
    WHERE id = ?
  `).get(id) as any;

  if (!row) return null;
  return rowToExperiment(row);
};

export const getAllExperiments = (): Experiment[] => {
  const rows = db.prepare(`
    SELECT id, name, status, traffic_type, created_at, updated_at
    FROM experiments
    ORDER BY created_at DESC
  `).all() as any[];

  return rows.map(rowToExperiment);
};

const isValidStatusTransition = (from: ExperimentStatus, to: ExperimentStatus): boolean => {
  const validTransitions: Record<ExperimentStatus, ExperimentStatus[]> = {
    [ExperimentStatus.CONFIGURING]: [ExperimentStatus.RUNNING],
    [ExperimentStatus.RUNNING]: [ExperimentStatus.PAUSED, ExperimentStatus.ENDED],
    [ExperimentStatus.PAUSED]: [ExperimentStatus.RUNNING, ExperimentStatus.ENDED],
    [ExperimentStatus.ENDED]: []
  };

  return validTransitions[from]?.includes(to) ?? false;
};

export const updateExperimentStatus = (
  experimentId: string,
  newStatus: ExperimentStatus
): Experiment => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (!isValidStatusTransition(experiment.status, newStatus)) {
    throw new Error(`Invalid status transition from ${experiment.status} to ${newStatus}`);
  }

  db.prepare(`
    UPDATE experiments
    SET status = ?, updated_at = ?
    WHERE id = ?
  `).run(newStatus, new Date().toISOString(), experimentId);

  return getExperimentById(experimentId)!;
};

export const updateVariants = (
  experimentId: string,
  variants: Omit<ExperimentVariant, 'id'>[]
): Experiment => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status !== ExperimentStatus.PAUSED) {
    throw new Error('Variants can only be updated when experiment is paused');
  }

  const totalPercentage = variants.reduce((sum, v) => sum + v.trafficPercentage, 0);
  if (Math.abs(totalPercentage - 100) > 0.001) {
    throw new Error('Traffic percentages must sum to 100%');
  }

  const controlVariant = variants.find(v => v.isControl);
  if (!controlVariant) {
    throw new Error('At least one variant must be the control group');
  }

  const existingVariants = experiment.variants;
  const newVariantKeys = new Set(variants.map(v => `${v.name}|${v.isControl ? '1' : '0'}`));
  const existingVariantKeys = new Set(existingVariants.map(v => `${v.name}|${v.isControl ? '1' : '0'}`));

  const keysToRemove = [...existingVariantKeys].filter(k => !newVariantKeys.has(k));
  if (keysToRemove.length > 0) {
    throw new Error('Cannot remove existing variants; only traffic percentage can be modified');
  }

  const transaction = db.transaction(() => {
    const updateVariant = db.prepare(`
      UPDATE variants
      SET traffic_percentage = ?
      WHERE experiment_id = ? AND name = ? AND is_control = ?
    `);

    for (const variant of variants) {
      const result = updateVariant.run(
        variant.trafficPercentage,
        experimentId,
        variant.name,
        variant.isControl ? 1 : 0
      );

      if (result.changes === 0) {
        db.prepare(`
          INSERT INTO variants (id, experiment_id, name, traffic_percentage, is_control)
          VALUES (?, ?, ?, ?, ?)
        `).run(
          uuidv4(),
          experimentId,
          variant.name,
          variant.trafficPercentage,
          variant.isControl ? 1 : 0
        );
      }
    }

    db.prepare(`
      UPDATE experiments SET updated_at = ? WHERE id = ?
    `).run(new Date().toISOString(), experimentId);
  });

  transaction();
  return getExperimentById(experimentId)!;
};

export const deleteExperiment = (experimentId: string): void => {
  const experiment = getExperimentById(experimentId);
  if (!experiment) {
    throw new Error('Experiment not found');
  }

  if (experiment.status === ExperimentStatus.RUNNING || 
      experiment.status === ExperimentStatus.PAUSED) {
    throw new Error('Cannot delete running or paused experiment');
  }

  const transaction = db.transaction(() => {
    db.prepare(`DELETE FROM grayscale_configs WHERE experiment_id = ?`).run(experimentId);
    db.prepare(`DELETE FROM metric_data WHERE experiment_id = ?`).run(experimentId);
    db.prepare(`DELETE FROM metrics WHERE experiment_id = ?`).run(experimentId);
    db.prepare(`DELETE FROM assignments WHERE experiment_id = ?`).run(experimentId);
    db.prepare(`DELETE FROM variants WHERE experiment_id = ?`).run(experimentId);
    db.prepare(`DELETE FROM experiments WHERE id = ?`).run(experimentId);
  });

  transaction();
};

export const getMetricsForExperiment = (experimentId: string): Metric[] => {
  return getMetricsByExperimentId(experimentId);
};
