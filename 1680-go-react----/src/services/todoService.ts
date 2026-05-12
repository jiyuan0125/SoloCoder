import { db } from '../database';
import { Todo, IssueStatus } from '../types';
import { nowISO, addDays } from '../utils';

export const createTodo = (
  issueId: number,
  userId: number,
  type: string,
  deadlineDays: number = 3
): Todo => {
  const now = nowISO();
  const deadline = addDays(new Date(), deadlineDays);

  const result = db.prepare(`
    INSERT INTO todos (
      issue_id, user_id, type, deadline_at, 
      is_completed, is_reminded, created_at
    ) VALUES (?, ?, ?, ?, 0, 0, ?)
  `).run(issueId, userId, type, deadline.toISOString(), now);

  return getTodoById(result.lastInsertRowid as number)!;
};

export const getTodoById = (id: number): Todo | undefined => {
  const row = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId, type,
      deadline_at as deadlineAt, is_completed as isCompleted,
      is_reminded as isReminded, created_at as createdAt
    FROM todos WHERE id = ?
  `).get(id) as any;
  return row;
};

export const listTodosByUser = (userId: number): Todo[] => {
  const rows = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId, type,
      deadline_at as deadlineAt, is_completed as isCompleted,
      is_reminded as isReminded, created_at as createdAt
    FROM todos WHERE user_id = ? ORDER BY deadline_at ASC
  `).all(userId) as any[];
  return rows;
};

export const listOverdueTodos = (): Todo[] => {
  const now = nowISO();
  const rows = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId, type,
      deadline_at as deadlineAt, is_completed as isCompleted,
      is_reminded as isReminded, created_at as createdAt
    FROM todos 
    WHERE is_completed = 0 AND deadline_at < ?
  `).all(now) as any[];
  return rows;
};

export const markTodoCompleted = (todoId: number): Todo => {
  db.prepare(`
    UPDATE todos SET is_completed = 1 WHERE id = ?
  `).run(todoId);

  return getTodoById(todoId)!;
};

export const markTodoReminded = (todoId: number): Todo => {
  db.prepare(`
    UPDATE todos SET is_reminded = 1 WHERE id = ?
  `).run(todoId);

  return getTodoById(todoId)!;
};

export const getTodoTypeForStatus = (status: IssueStatus): string | null => {
  const typeMap: Record<IssueStatus, string> = {
    [IssueStatus.WAITING_VOTE_CONFIRM]: 'committee_vote_confirm',
    [IssueStatus.WAITING_EXEC_CONFIRM]: 'executor_confirm',
    [IssueStatus.EXECUTING]: 'execute_task',
    [IssueStatus.DRAFT]: '',
    [IssueStatus.PUBLICITY]: '',
    [IssueStatus.VOTING]: '',
    [IssueStatus.VOTED]: '',
    [IssueStatus.COMPLETED]: '',
    [IssueStatus.REJECTED]: '',
    [IssueStatus.REPEAL]: ''
  };

  return typeMap[status] || null;
};
