import { db } from '../database';
import { Todo, TodoStatus } from '../types';
import { generateId, getCurrentTime } from '../utils';
import { createAuditLog } from './auditService';
import { getProjectById } from './projectService';

interface DBTodo {
  id: string;
  project_id: string;
  title: string;
  description: string;
  deadline: string;
  status: string;
  created_at: string;
}

function mapTodo(dbTodo: DBTodo): Todo {
  return {
    id: dbTodo.id,
    projectId: dbTodo.project_id,
    title: dbTodo.title,
    description: dbTodo.description,
    deadline: dbTodo.deadline,
    status: dbTodo.status as TodoStatus,
    createdAt: dbTodo.created_at
  };
}

export function createTodo(
  projectId: string,
  title: string,
  description: string,
  deadline: string,
  operator: string
): string {
  const id = generateId();
  const now = getCurrentTime();

  const stmt = db.prepare(`
    INSERT INTO todos (
      id, project_id, title, description, deadline, status, created_at
    ) VALUES (?, ?, ?, ?, ?, '待办', ?)
  `);

  stmt.run(id, projectId, title, description, deadline, now);

  const todo = getTodoById(id)!;
  const project = getProjectById(projectId);

  createAuditLog(
    operator,
    '创建待办',
    `为项目"${project?.projectName || projectId}"创建待办：${title}`,
    null,
    todo
  );

  return id;
}

export function getTodoById(id: string): Todo | null {
  const dbTodo = db.prepare('SELECT * FROM todos WHERE id = ?').get(id) as DBTodo | undefined;
  return dbTodo ? mapTodo(dbTodo) : null;
}

export function getTodosByProject(projectId: string): Todo[] {
  const dbTodos = db.prepare(`
    SELECT * FROM todos WHERE project_id = ? ORDER BY deadline ASC
  `).all(projectId) as DBTodo[];
  return dbTodos.map(mapTodo);
}

export function getAllTodos(): Todo[] {
  const dbTodos = db.prepare(`
    SELECT * FROM todos ORDER BY deadline ASC
  `).all() as DBTodo[];
  return dbTodos.map(mapTodo);
}

export function completeTodo(todoId: string, operator: string): { success: boolean; message: string } {
  const todo = getTodoById(todoId);
  if (!todo) {
    return { success: false, message: '待办不存在' };
  }

  if (todo.status === '已完成') {
    return { success: false, message: '待办已完成' };
  }

  const beforeSnapshot = { ...todo };
  const now = getCurrentTime();

  const stmt = db.prepare(`
    UPDATE todos SET status = '已完成' WHERE id = ?
  `);

  stmt.run(todoId);

  const updatedTodo = getTodoById(todoId)!;

  createAuditLog(
    operator,
    '完成待办',
    `完成待办：${todo.title}`,
    beforeSnapshot,
    updatedTodo
  );

  return { success: true, message: '待办已完成' };
}

export function checkOverdueTodos(): number {
  const now = getCurrentTime();
  const overdueTodos = db.prepare(`
    SELECT * FROM todos 
    WHERE status = '待办' AND deadline < ?
  `).all(now) as DBTodo[];

  let count = 0;
  for (const dbTodo of overdueTodos) {
    const beforeSnapshot = mapTodo(dbTodo);
    
    db.prepare(`
      UPDATE todos SET status = '逾期' WHERE id = ?
    `).run(dbTodo.id);

    const updatedTodo = getTodoById(dbTodo.id)!;
    const project = getProjectById(dbTodo.project_id);

    createAuditLog(
      'system',
      '待办逾期',
      `待办"${dbTodo.title}"已逾期，所属项目：${project?.projectName || dbTodo.project_id}`,
      beforeSnapshot,
      updatedTodo
    );

    count++;
  }

  return count;
}
