import { db } from '../database';
import { Opinion, Issue, IssueStatus } from '../types';
import { nowISO } from '../utils';

export const addOpinion = (
  issueId: number,
  userId: number,
  content: string
): Opinion => {
  const now = nowISO();
  const result = db.prepare(`
    INSERT INTO opinions (issue_id, user_id, content, created_at)
    VALUES (?, ?, ?, ?)
  `).run(issueId, userId, content, now);

  return getOpinionById(result.lastInsertRowid as number)!;
};

export const getOpinionById = (id: number): Opinion | undefined => {
  const row = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId, 
      content, created_at as createdAt
    FROM opinions WHERE id = ?
  `).get(id) as any;
  return row;
};

export const listOpinionsByIssue = (issueId: number): Opinion[] => {
  const rows = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId, 
      content, created_at as createdAt
    FROM opinions WHERE issue_id = ? ORDER BY created_at DESC
  `).all(issueId) as any[];
  return rows;
};

export const canAddOpinion = (issue: Issue): boolean => {
  return issue.status === IssueStatus.PUBLICITY;
};
