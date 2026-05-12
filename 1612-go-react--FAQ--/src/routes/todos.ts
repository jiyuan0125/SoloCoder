import express from 'express';
import db from '../db';

const router = express.Router();

interface Todo {
  id: number;
  title: string;
  description: string | null;
  due_date: string | null;
  is_completed: number;
  created_at: string;
  updated_at: string;
}

interface TodoWithHighlight extends Todo {
  is_overdue: boolean;
}

function isOverdue(dueDate: string | null, isCompleted: number): boolean {
  if (isCompleted === 1 || !dueDate) return false;
  const due = new Date(dueDate);
  const now = new Date();
  return due < now;
}

router.get('/', (req, res) => {
  const todos = db.prepare<[], Todo>(`
    SELECT * FROM todos ORDER BY is_completed ASC, created_at DESC
  `).all();

  const enriched = todos.map((todo) => ({
    ...todo,
    is_overdue: isOverdue(todo.due_date, todo.is_completed),
  }));

  res.json(enriched);
});

router.post('/', (req, res) => {
  const { title, description, due_date } = req.body;

  if (!title || typeof title !== 'string') {
    return res.status(400).json({ error: 'Title is required and must be a string' });
  }

  const stmt = db.prepare(`
    INSERT INTO todos (title, description, due_date) VALUES (?, ?, ?)
  `);

  const result = stmt.run(title, description ?? null, due_date ?? null);
  const id = result.lastInsertRowid as number;

  const todo = db.prepare<[number], Todo>(`
    SELECT * FROM todos WHERE id = ?
  `).get(id);

  res.status(201).json({
    ...todo,
    is_overdue: isOverdue(todo!.due_date, todo!.is_completed),
  });
});

router.put('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);
  const { title, description, due_date, is_completed } = req.body;

  const existing = db.prepare<[number], Todo>(`
    SELECT * FROM todos WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'Todo not found' });
  }

  const updates: string[] = [];
  const values: unknown[] = [];

  if (title !== undefined) {
    updates.push('title = ?');
    values.push(title);
  }

  if (description !== undefined) {
    updates.push('description = ?');
    values.push(description);
  }

  if (due_date !== undefined) {
    updates.push('due_date = ?');
    values.push(due_date);
  }

  if (is_completed !== undefined) {
    updates.push('is_completed = ?');
    values.push(is_completed ? 1 : 0);
  }

  if (updates.length > 0) {
    updates.push('updated_at = CURRENT_TIMESTAMP');
    values.push(id);

    const sql = `UPDATE todos SET ${updates.join(', ')} WHERE id = ?`;
    db.prepare(sql).run(...values);
  }

  const updated = db.prepare<[number], Todo>(`
    SELECT * FROM todos WHERE id = ?
  `).get(id);

  res.json({
    ...updated,
    is_overdue: isOverdue(updated!.due_date, updated!.is_completed),
  });
});

router.delete('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);

  const existing = db.prepare<[number], Todo>(`
    SELECT * FROM todos WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'Todo not found' });
  }

  db.prepare(`DELETE FROM todos WHERE id = ?`).run(id);
  res.status(204).send();
});

export default router;
