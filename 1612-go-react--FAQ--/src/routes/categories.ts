import express from 'express';
import db from '../db';

const router = express.Router();

interface Category {
  id: number;
  name: string;
  parent_id: number | null;
  created_at: string;
  updated_at: string;
}

interface CategoryWithChildren extends Category {
  children?: Category[];
}

router.get('/', (req, res) => {
  const categories = db.prepare<[], Category>(`
    SELECT * FROM categories ORDER BY id ASC
  `).all();

  const categoryMap = new Map<number, CategoryWithChildren>();
  const rootCategories: CategoryWithChildren[] = [];

  categories.forEach((cat) => {
    categoryMap.set(cat.id, { ...cat, children: [] });
  });

  categories.forEach((cat) => {
    const current = categoryMap.get(cat.id)!;
    if (cat.parent_id === null) {
      rootCategories.push(current);
    } else {
      const parent = categoryMap.get(cat.parent_id);
      if (parent) {
        parent.children!.push(current);
      }
    }
  });

  res.json(rootCategories);
});

router.post('/', (req, res) => {
  const { name, parent_id } = req.body;

  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'Name is required and must be a string' });
  }

  if (parent_id !== undefined && parent_id !== null) {
    const parent = db.prepare<[number], Category>(`
      SELECT * FROM categories WHERE id = ?
    `).get(parent_id);

    if (!parent) {
      return res.status(400).json({ error: 'Parent category does not exist' });
    }

    if (parent.parent_id !== null) {
      return res.status(400).json({ error: 'Cannot create category under a second-level category' });
    }
  }

  const stmt = db.prepare(`
    INSERT INTO categories (name, parent_id) VALUES (?, ?)
  `);

  const result = stmt.run(name, parent_id ?? null);
  const id = result.lastInsertRowid as number;

  const category = db.prepare<[number], Category>(`
    SELECT * FROM categories WHERE id = ?
  `).get(id);

  res.status(201).json(category);
});

router.put('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);
  const { name, parent_id } = req.body;

  const existing = db.prepare<[number], Category>(`
    SELECT * FROM categories WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'Category not found' });
  }

  if (parent_id !== undefined) {
    if (parent_id === id) {
      return res.status(400).json({ error: 'Parent cannot be the category itself' });
    }

    if (parent_id !== null) {
      const newParent = db.prepare<[number], Category>(`
        SELECT * FROM categories WHERE id = ?
      `).get(parent_id);

      if (!newParent) {
        return res.status(400).json({ error: 'Parent category does not exist' });
      }

      if (newParent.parent_id !== null) {
        return res.status(400).json({ error: 'Parent must be a first-level category' });
      }
    }

    const isMovingUp = existing.parent_id !== null && parent_id === null;
    const isMovingDown = existing.parent_id === null && parent_id !== null;

    if (isMovingDown) {
      const hasChildren = db.prepare<[number], { count: number }>(`
        SELECT COUNT(*) as count FROM categories WHERE parent_id = ?
      `).get(id);

      if (hasChildren && hasChildren.count > 0) {
        return res.status(400).json({ error: 'First-level category with children cannot become second-level' });
      }
    }

    db.prepare(`
      UPDATE categories SET parent_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
    `).run(parent_id ?? null, id);

    const updateStmt = db.prepare(`
      UPDATE faqs SET category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN (
        SELECT f.id FROM faqs f
        JOIN categories c ON f.category_id = c.id
        WHERE c.parent_id = ?
      )
    `);

    if (isMovingUp) {
      updateStmt.run(id, id);
    }
  }

  if (name && typeof name === 'string') {
    db.prepare(`
      UPDATE categories SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
    `).run(name, id);
  }

  const updated = db.prepare<[number], Category>(`
    SELECT * FROM categories WHERE id = ?
  `).get(id);

  res.json(updated);
});

router.delete('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);

  const existing = db.prepare<[number], Category>(`
    SELECT * FROM categories WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'Category not found' });
  }

  if (existing.parent_id === null) {
    const childCount = db.prepare<[number], { count: number }>(`
      SELECT COUNT(*) as count FROM categories WHERE parent_id = ?
    `).get(id);

    const faqCount = db.prepare<[number, number], { count: number }>(`
      SELECT COUNT(*) as count FROM faqs f
      JOIN categories c ON f.category_id = c.id
      WHERE c.id = ? OR c.parent_id = ?
    `).get(id, id);

    if ((childCount && childCount.count > 0) || (faqCount && faqCount.count > 0)) {
      return res.status(409).json({ error: 'Cannot delete first-level category with children or FAQs' });
    }
  } else {
    const faqCount = db.prepare<[number], { count: number }>(`
      SELECT COUNT(*) as count FROM faqs WHERE category_id = ?
    `).get(id);

    if (faqCount && faqCount.count > 0) {
      return res.status(409).json({ error: 'Cannot delete category with FAQs' });
    }
  }

  db.prepare(`DELETE FROM categories WHERE id = ?`).run(id);
  res.status(204).send();
});

export default router;
