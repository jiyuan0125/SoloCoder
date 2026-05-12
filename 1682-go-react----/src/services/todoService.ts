import db from '../database';
import { Todo, TodoType, TodoStatus } from '../types';
import { generateId, now } from '../utils';
import { STAGE_DURATION, SECRETARY_USER, VOTING_ADMIN_USER, DIRECTOR_USER } from '../constants';

const mapRowToTodo = (row: any): Todo => ({
  id: row.id,
  proposalId: row.proposal_id,
  type: row.type as TodoType,
  assignee: row.assignee,
  status: row.status as TodoStatus,
  createdAt: row.created_at,
  handledAt: row.handled_at,
  escalatedAt: row.escalated_at
});

export const todoService = {
  createTodo: (
    proposalId: string,
    type: TodoType
  ): Todo => {
    const id = generateId();
    const currentTime = now();
    
    let assignee = SECRETARY_USER;
    if (type === TodoType.VOTING_ADMIN) {
      assignee = VOTING_ADMIN_USER;
    } else if (type === TodoType.NOTIFY_DIRECTOR) {
      assignee = DIRECTOR_USER;
    }
    
    db.prepare(`
      INSERT INTO todos (
        id, proposal_id, type, assignee, status, created_at
      ) VALUES (?, ?, ?, ?, ?, ?)
    `).run(
      id,
      proposalId,
      type,
      assignee,
      TodoStatus.PENDING,
      currentTime
    );
    
    return todoService.getTodo(id)!;
  },

  getTodo: (id: string): Todo | null => {
    const row = db.prepare('SELECT * FROM todos WHERE id = ?').get(id);
    return row ? mapRowToTodo(row) : null;
  },

  getTodosByProposal: (proposalId: string): Todo[] => {
    const rows = db.prepare(`
      SELECT * FROM todos WHERE proposal_id = ? ORDER BY created_at DESC
    `).all(proposalId) as any[];
    return rows.map(mapRowToTodo);
  },

  getTodosByAssignee: (assignee: string, status?: TodoStatus): Todo[] => {
    let query = 'SELECT * FROM todos WHERE assignee = ?';
    const params: any[] = [assignee];
    
    if (status) {
      query += ' AND status = ?';
      params.push(status);
    }
    
    query += ' ORDER BY created_at DESC';
    
    const rows = db.prepare(query).all(...params) as any[];
    return rows.map(mapRowToTodo);
  },

  getPendingTodos: (): Todo[] => {
    const rows = db.prepare(`
      SELECT * FROM todos WHERE status = ? ORDER BY created_at ASC
    `).all(TodoStatus.PENDING) as any[];
    return rows.map(mapRowToTodo);
  },

  completeTodo: (todoId: string): Todo | null => {
    const todo = todoService.getTodo(todoId);
    if (!todo) return null;
    
    if (todo.status !== TodoStatus.PENDING) {
      return null;
    }
    
    db.prepare(`
      UPDATE todos SET status = ?, handled_at = ? WHERE id = ?
    `).run(TodoStatus.COMPLETED, now(), todoId);
    
    return todoService.getTodo(todoId);
  },

  escalateTodo: (todoId: string): Todo | null => {
    const todo = todoService.getTodo(todoId);
    if (!todo) return null;
    
    if (todo.status !== TodoStatus.PENDING) {
      return null;
    }
    
    db.prepare(`
      UPDATE todos SET status = ?, escalated_at = ?, assignee = ? WHERE id = ?
    `).run(TodoStatus.ESCALATED, now(), DIRECTOR_USER, todoId);
    
    return todoService.getTodo(todoId);
  },

  checkAndEscalateOverdueTodos: (): Todo[] => {
    const currentTime = now();
    const overdueThreshold = currentTime - STAGE_DURATION.TODO_TIMEOUT;
    
    const overdueTodos = db.prepare(`
      SELECT * FROM todos 
      WHERE status = ? AND created_at <= ?
    `).all(TodoStatus.PENDING, overdueThreshold) as any[];
    
    const escalatedTodos: Todo[] = [];
    
    for (const row of overdueTodos) {
      const escalated = todoService.createTodo(
        row.proposal_id,
        TodoType.NOTIFY_DIRECTOR
      );
      todoService.escalateTodo(row.id);
      escalatedTodos.push(escalated);
    }
    
    return escalatedTodos;
  },

  getTodos: (): Todo[] => {
    const rows = db.prepare('SELECT * FROM todos ORDER BY created_at DESC').all() as any[];
    return rows.map(mapRowToTodo);
  }
};
