import { db } from '../database';
import { Issue, IssueStatus } from '../types';
import { nowISO, addDays, parseISO, isAfter } from '../utils';

const DEFAULT_PUBLICITY_DAYS = 7;

export const createIssue = (
  title: string,
  content: string,
  category: string,
  attachments: string[],
  initiatorId: number
): Issue => {
  const now = nowISO();
  const publicityStart = parseISO(now);
  const publicityEnd = addDays(publicityStart, DEFAULT_PUBLICITY_DAYS);

  const result = db.prepare(`
    INSERT INTO issues (
      title, content, category, attachments, initiator_id,
      status, publicity_start_at, publicity_end_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    title,
    content,
    category,
    JSON.stringify(attachments),
    initiatorId,
    IssueStatus.PUBLICITY,
    publicityStart.toISOString(),
    publicityEnd.toISOString(),
    now,
    now
  );

  return getIssueById(result.lastInsertRowid as number)!;
};

export const getIssueById = (id: number): Issue | undefined => {
  const row = db.prepare(`
    SELECT 
      id, title, content, category, attachments,
      initiator_id as initiatorId, status,
      publicity_start_at as publicityStartAt,
      publicity_end_at as publicityEndAt,
      created_at as createdAt, updated_at as updatedAt
    FROM issues WHERE id = ?
  `).get(id) as any;
  return row;
};

export const listIssues = (): Issue[] => {
  const rows = db.prepare(`
    SELECT 
      id, title, content, category, attachments,
      initiator_id as initiatorId, status,
      publicity_start_at as publicityStartAt,
      publicity_end_at as publicityEndAt,
      created_at as createdAt, updated_at as updatedAt
    FROM issues ORDER BY created_at DESC
  `).all() as any[];
  return rows;
};

export const isPublicityPeriodEnded = (issue: Issue): boolean => {
  const endAt = parseISO(issue.publicityEndAt);
  return isAfter(new Date(), endAt);
};

export const updateIssueStatus = (issueId: number, newStatus: IssueStatus): Issue => {
  const now = nowISO();
  db.prepare(`
    UPDATE issues SET status = ?, updated_at = ? WHERE id = ?
  `).run(newStatus, now, issueId);

  return getIssueById(issueId)!;
};

export const checkStatusTransition = (currentStatus: IssueStatus, nextStatus: IssueStatus): boolean => {
  const validTransitions: Record<IssueStatus, IssueStatus[]> = {
    [IssueStatus.DRAFT]: [IssueStatus.PUBLICITY],
    [IssueStatus.PUBLICITY]: [IssueStatus.WAITING_VOTE_CONFIRM, IssueStatus.REJECTED],
    [IssueStatus.WAITING_VOTE_CONFIRM]: [IssueStatus.VOTING, IssueStatus.REJECTED],
    [IssueStatus.VOTING]: [IssueStatus.VOTED],
    [IssueStatus.VOTED]: [IssueStatus.WAITING_EXEC_CONFIRM, IssueStatus.REPEAL],
    [IssueStatus.WAITING_EXEC_CONFIRM]: [IssueStatus.EXECUTING],
    [IssueStatus.EXECUTING]: [IssueStatus.COMPLETED],
    [IssueStatus.COMPLETED]: [],
    [IssueStatus.REJECTED]: [],
    [IssueStatus.REPEAL]: []
  };

  return validTransitions[currentStatus]?.includes(nextStatus) ?? false;
};
