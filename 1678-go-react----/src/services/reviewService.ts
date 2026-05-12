import { db } from '../database';
import { Expert, Review } from '../types';
import {
  generateId,
  getCurrentTime,
  validateScore,
  calculateAverageScore,
  getReviewDecision,
  getRandomItems
} from '../utils';
import { createAuditLog } from './auditService';
import { getProjectById, updateProjectStatus } from './projectService';

interface DBExpert {
  id: string;
  name: string;
  field: string;
  created_at: string;
}

function mapExpert(dbExp: DBExpert): Expert {
  return {
    id: dbExp.id,
    name: dbExp.name,
    field: dbExp.field,
    createdAt: dbExp.created_at
  };
}

export function getAllExperts(): Expert[] {
  const dbExperts = db.prepare('SELECT * FROM experts').all() as DBExpert[];
  return dbExperts.map(mapExpert);
}

export function getExpertById(id: string): Expert | null {
  const dbExp = db.prepare('SELECT * FROM experts WHERE id = ?').get(id) as DBExpert | undefined;
  return dbExp ? mapExpert(dbExp) : null;
}

export function selectRandomExperts(min: number = 3, max: number = 5): Expert[] {
  const allExperts = getAllExperts();
  return getRandomItems(allExperts, min, max);
}

export function submitReview(
  projectId: string,
  expertId: string,
  score: number,
  operator: string
): { success: boolean; message: string; review?: Review } {
  const project = getProjectById(projectId);
  if (!project) {
    return { success: false, message: '项目不存在' };
  }

  if (project.status !== '评审') {
    return { success: false, message: '项目当前不在评审阶段' };
  }

  if (!validateScore(score)) {
    return { success: false, message: '分数必须在1到100之间' };
  }

  const existingReview = db.prepare(`
    SELECT * FROM reviews WHERE project_id = ? AND expert_id = ?
  `).get(projectId, expertId);

  if (existingReview) {
    return { success: false, message: '该专家已对此项目提交过评审' };
  }

  const id = generateId();
  const now = getCurrentTime();

  const stmt = db.prepare(`
    INSERT INTO reviews (id, project_id, expert_id, score, created_at)
    VALUES (?, ?, ?, ?, ?)
  `);

  stmt.run(id, projectId, expertId, score, now);

  const review: Review = {
    id,
    projectId,
    expertId,
    score,
    createdAt: now
  };

  const expert = getExpertById(expertId);

  createAuditLog(
    operator,
    '专家评审',
    `专家"${expert?.name || expertId}"对项目"${project.projectName}"打分：${score}分`,
    null,
    review
  );

  return { success: true, message: '评审提交成功', review };
}

export function getProjectReviews(projectId: string): Review[] {
  return db.prepare(`
    SELECT * FROM reviews WHERE project_id = ? ORDER BY created_at DESC
  `).all(projectId) as Review[];
}

export function finalizeReview(
  projectId: string,
  operator: string
): { success: boolean; message: string; decision?: string; averageScore?: number } {
  const project = getProjectById(projectId);
  if (!project) {
    return { success: false, message: '项目不存在' };
  }

  if (project.status !== '评审') {
    return { success: false, message: '项目当前不在评审阶段' };
  }

  const reviews = getProjectReviews(projectId);
  if (reviews.length === 0) {
    return { success: false, message: '暂无评审数据' };
  }

  const scores = reviews.map(r => r.score);
  const averageScore = calculateAverageScore(scores);
  const decision = getReviewDecision(averageScore);

  const beforeSnapshot = { ...project };

  const updateResult = updateProjectStatus(
    projectId,
    decision,
    operator,
    `专家评审平均分为：${averageScore.toFixed(2)}分，评审结果：${decision}`
  );

  if (!updateResult.success) {
    return { success: false, message: updateResult.message };
  }

  createAuditLog(
    operator,
    '评审结果',
    `项目"${project.projectName}"评审完成：平均分${averageScore.toFixed(2)}，决定：${decision}`,
    beforeSnapshot,
    updateResult.project || null
  );

  return {
    success: true,
    message: '评审完成',
    decision,
    averageScore
  };
}
