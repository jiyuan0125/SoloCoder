import db from '../database';
import { ProjectStatus, InspectionResult } from '../types';
import { AppError } from '../middleware/errorHandler';

const isValidInspectionResult = (result: string): result is InspectionResult => {
  return result === InspectionResult.PASSED ||
         result === InspectionResult.PASSED_AFTER_RECTIFICATION ||
         result === InspectionResult.FAILED;
};

const addDays = (dateStr: string, days: number): string => {
  const date = new Date(dateStr);
  date.setDate(date.getDate() + days);
  return date.toISOString().split('T')[0];
};

export const submitMidTermInspection = (projectId: number) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status !== ProjectStatus.INITIAL_FUND_ALLOCATED) {
    throw new AppError('只有启动资金已拨付的项目才能申请中期验收', 400);
  }

  const now = new Date().toISOString();
  db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
    .run(ProjectStatus.MID_TERM_INSPECTION_PENDING, now, projectId);

  return true;
};

export const submitFinalInspection = (projectId: number) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status !== ProjectStatus.MID_TERM_FUND_ALLOCATED) {
    throw new AppError('只有中期资金已拨付的项目才能申请终期验收', 400);
  }

  const now = new Date().toISOString();
  db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
    .run(ProjectStatus.FINAL_INSPECTION_PENDING, now, projectId);

  return true;
};

export const processMidTermInspection = (
  projectId: number,
  result: string,
  comments: string | null,
  inspector: string
) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status !== ProjectStatus.MID_TERM_INSPECTION_PENDING) {
    throw new AppError('项目不在中期验收状态', 400);
  }

  if (!isValidInspectionResult(result)) {
    throw new AppError('验收结果无效，必须是：通过、整改后通过、不通过', 400);
  }

  const now = new Date().toISOString();
  const inspectionDate = new Date().toISOString().split('T')[0];
  let rectificationDeadline: string | null = null;
  let nextStatus: ProjectStatus;

  if (result === InspectionResult.PASSED) {
    nextStatus = ProjectStatus.MID_TERM_INSPECTION_PASSED;
  } else if (result === InspectionResult.PASSED_AFTER_RECTIFICATION) {
    nextStatus = ProjectStatus.MID_TERM_RECTIFICATION_REQUIRED;
    rectificationDeadline = addDays(inspectionDate, 30);
  } else {
    nextStatus = ProjectStatus.TERMINATED;
  }

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO inspection_records (project_id, stage, result, comments, 
                                       rectification_deadline, inspector, inspected_at, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `).run(projectId, 'mid_term', result, comments, rectificationDeadline, inspector, now, now);

    db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
      .run(nextStatus, now, projectId);
  });

  transaction();

  if (result === InspectionResult.FAILED) {
    return { status: nextStatus, message: '项目验收不通过，已终止并启动审计' };
  }

  return { status: nextStatus, rectificationDeadline };
};

export const processFinalInspection = (
  projectId: number,
  result: string,
  comments: string | null,
  inspector: string
) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status !== ProjectStatus.FINAL_INSPECTION_PENDING) {
    throw new AppError('项目不在终期验收状态', 400);
  }

  if (!isValidInspectionResult(result)) {
    throw new AppError('验收结果无效，必须是：通过、整改后通过、不通过', 400);
  }

  const now = new Date().toISOString();
  const inspectionDate = new Date().toISOString().split('T')[0];
  let rectificationDeadline: string | null = null;
  let nextStatus: ProjectStatus;

  if (result === InspectionResult.PASSED) {
    nextStatus = ProjectStatus.FINAL_INSPECTION_PASSED;
  } else if (result === InspectionResult.PASSED_AFTER_RECTIFICATION) {
    nextStatus = ProjectStatus.FINAL_RECTIFICATION_REQUIRED;
    rectificationDeadline = addDays(inspectionDate, 30);
  } else {
    nextStatus = ProjectStatus.TERMINATED;
  }

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO inspection_records (project_id, stage, result, comments, 
                                       rectification_deadline, inspector, inspected_at, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `).run(projectId, 'final', result, comments, rectificationDeadline, inspector, now, now);

    db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
      .run(nextStatus, now, projectId);
  });

  transaction();

  if (result === InspectionResult.FAILED) {
    return { status: nextStatus, message: '项目验收不通过，已终止并启动审计' };
  }

  return { status: nextStatus, rectificationDeadline };
};

export const submitRectificationCompletion = (projectId: number, stage: 'mid_term' | 'final') => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  const expectedStatus = stage === 'mid_term' 
    ? ProjectStatus.MID_TERM_RECTIFICATION_REQUIRED 
    : ProjectStatus.FINAL_RECTIFICATION_REQUIRED;

  if (project.status !== expectedStatus) {
    throw new AppError('项目不在整改状态', 400);
  }

  const now = new Date().toISOString();
  const nextStatus = stage === 'mid_term' 
    ? ProjectStatus.MID_TERM_INSPECTION_PASSED 
    : ProjectStatus.FINAL_INSPECTION_PASSED;

  db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
    .run(nextStatus, now, projectId);

  return true;
};
