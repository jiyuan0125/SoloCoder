import { Request, Response } from 'express';
import { db } from '../database';
import { Category, CategoryType } from '../types';

interface DBCategory {
  id: number;
  name: string;
  created_at: string;
}

function mapCategory(dbCategory: DBCategory): Category {
  return {
    id: dbCategory.id,
    name: dbCategory.name as CategoryType,
    createdAt: dbCategory.created_at
  };
}

export function getAllCategories(_req: Request, res: Response): void {
  try {
    const categories = db
      .prepare('SELECT * FROM categories ORDER BY id')
      .all() as DBCategory[];
    
    res.json(categories.map(mapCategory));
  } catch (error) {
    res.status(500).json({ error: '获取分类列表失败' });
  }
}

export function getCategoryById(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);
    
    if (isNaN(id)) {
      res.status(400).json({ error: '无效的分类ID' });
      return;
    }

    const category = db
      .prepare('SELECT * FROM categories WHERE id = ?')
      .get(id) as DBCategory | undefined;

    if (!category) {
      res.status(404).json({ error: '分类不存在' });
      return;
    }

    res.json(mapCategory(category));
  } catch (error) {
    res.status(500).json({ error: '获取分类详情失败' });
  }
}

export function createCategory(req: Request, res: Response): void {
  try {
    const { name } = req.body as { name: string };

    if (!name || typeof name !== 'string') {
      res.status(400).json({ error: '分类名称不能为空' });
      return;
    }

    const validCategories = Object.values(CategoryType);
    if (!validCategories.includes(name as CategoryType)) {
      res.status(400).json({ 
        error: '无效的分类名称', 
        allowed: validCategories 
      });
      return;
    }

    const existing = db
      .prepare('SELECT id FROM categories WHERE name = ?')
      .get(name);

    if (existing) {
      res.status(409).json({ error: '分类已存在' });
      return;
    }

    const now = new Date().toISOString();
    const result = db
      .prepare('INSERT INTO categories (name, created_at) VALUES (?, ?)')
      .run(name, now);

    const category = db
      .prepare('SELECT * FROM categories WHERE id = ?')
      .get(result.lastInsertRowid) as DBCategory;

    res.status(201).json(mapCategory(category));
  } catch (error) {
    res.status(500).json({ error: '创建分类失败' });
  }
}

export function updateCategory(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);
    const { name } = req.body as { name: string };

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的分类ID' });
      return;
    }

    if (!name || typeof name !== 'string') {
      res.status(400).json({ error: '分类名称不能为空' });
      return;
    }

    const validCategories = Object.values(CategoryType);
    if (!validCategories.includes(name as CategoryType)) {
      res.status(400).json({ 
        error: '无效的分类名称', 
        allowed: validCategories 
      });
      return;
    }

    const existing = db
      .prepare('SELECT id FROM categories WHERE id = ?')
      .get(id);

    if (!existing) {
      res.status(404).json({ error: '分类不存在' });
      return;
    }

    const duplicate = db
      .prepare('SELECT id FROM categories WHERE name = ? AND id != ?')
      .get(name, id);

    if (duplicate) {
      res.status(409).json({ error: '分类名称已存在' });
      return;
    }

    db
      .prepare('UPDATE categories SET name = ? WHERE id = ?')
      .run(name, id);

    const category = db
      .prepare('SELECT * FROM categories WHERE id = ?')
      .get(id) as DBCategory;

    res.json(mapCategory(category));
  } catch (error) {
    res.status(500).json({ error: '更新分类失败' });
  }
}

export function deleteCategory(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的分类ID' });
      return;
    }

    const existing = db
      .prepare('SELECT id FROM categories WHERE id = ?')
      .get(id);

    if (!existing) {
      res.status(404).json({ error: '分类不存在' });
      return;
    }

    const feedbackCount = db
      .prepare('SELECT COUNT(*) as count FROM feedbacks WHERE category_id = ?')
      .get(id) as { count: number };

    if (feedbackCount.count > 0) {
      res.status(409).json({ error: '该分类下存在反馈记录，无法删除' });
      return;
    }

    db
      .prepare('DELETE FROM categories WHERE id = ?')
      .run(id);

    res.status(204).send();
  } catch (error) {
    res.status(500).json({ error: '删除分类失败' });
  }
}
