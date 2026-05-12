import express from 'express';
import db from '../db';

const router = express.Router();

interface FAQ {
  id: number;
  title: string;
  answer: string;
  sort_weight: number;
  is_enabled: number;
  category_id: number;
  created_at: string;
  updated_at: string;
}

interface FAQWithSatisfaction extends FAQ {
  satisfaction: string | number;
}

function escapeRegExp(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function highlight(text: string, keywords: string[]): string {
  if (keywords.length === 0) return text;

  let result = text;
  const uniqueKeywords = [...new Set(keywords.filter((k) => k.length > 0))];

  for (const keyword of uniqueKeywords) {
    const regex = new RegExp(escapeRegExp(keyword), 'gi');
    result = result.replace(regex, '<em>$&</em>');
  }

  return result;
}

function calculateScore(title: string, answer: string, keywords: string[]): number {
  let score = 0;
  const titleLower = title.toLowerCase();
  const answerLower = answer.toLowerCase();

  for (const keyword of keywords) {
    const kwLower = keyword.toLowerCase();
    if (!kwLower) continue;

    if (titleLower === kwLower) {
      score += 10;
    } else if (titleLower.includes(kwLower)) {
      score += 5;
    }

    if (answerLower.includes(kwLower)) {
      score += 2;
    }
  }

  return score;
}

function getSatisfaction(faqId: number): string | number {
  const stats = db.prepare<[number], { total: number; helpful: number }>(`
    SELECT
      COUNT(*) as total,
      SUM(CASE WHEN is_helpful = 1 THEN 1 ELSE 0 END) as helpful
    FROM feedbacks WHERE faq_id = ?
  `).get(faqId);

  if (!stats || stats.total === 0) {
    return 'N/A';
  }

  return (stats.helpful / stats.total) * 100;
}

router.get('/search', (req, res) => {
  const query = (req.query.q as string) || '';

  const trimmed = query.trim();
  const isOnlySpecialOrWhitespace = /^[\s\p{P}]*$/u.test(trimmed);

  db.prepare(`
    INSERT INTO search_stats (query, has_results) VALUES (?, 0)
  `).run(trimmed);

  if (isOnlySpecialOrWhitespace) {
    return res.json({ results: [], total: 0 });
  }

  const keywords = trimmed.split(/\s+/).filter((k) => k.length > 0);

  const faqs = db.prepare<[], FAQ>(`
    SELECT * FROM faqs WHERE is_enabled = 1
  `).all();

  const scored = faqs
    .map((faq) => ({
      ...faq,
      score: calculateScore(faq.title, faq.answer, keywords),
      satisfaction: getSatisfaction(faq.id),
    }))
    .filter((item) => item.score > 0);

  scored.sort((a, b) => {
    if (b.score !== a.score) return b.score - a.score;
    if (b.sort_weight !== a.sort_weight) return b.sort_weight - a.sort_weight;
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });

  const results = scored.map((item) => ({
    ...item,
    title: highlight(item.title, keywords),
    answer: highlight(item.answer, keywords),
  }));

  if (results.length > 0) {
    const searchStatId = db.prepare<[], { id: number }>(`
      SELECT id FROM search_stats ORDER BY id DESC LIMIT 1
    `).get();

    if (searchStatId) {
      db.prepare(`
        UPDATE search_stats SET has_results = 1 WHERE id = ?
      `).run(searchStatId.id);
    }
  }

  res.json({ results, total: results.length });
});

router.get('/', (req, res) => {
  const faqs = db.prepare<[], FAQ>(`
    SELECT * FROM faqs ORDER BY sort_weight DESC, updated_at DESC
  `).all();

  const enriched = faqs.map((faq) => ({
    ...faq,
    satisfaction: getSatisfaction(faq.id),
  }));

  res.json(enriched);
});

router.get('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);

  const faq = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  if (!faq) {
    return res.status(404).json({ error: 'FAQ not found' });
  }

  res.json({
    ...faq,
    satisfaction: getSatisfaction(faq.id),
  });
});

router.post('/', (req, res) => {
  const { title, answer, sort_weight, is_enabled, category_id } = req.body;

  if (!title || typeof title !== 'string') {
    return res.status(400).json({ error: 'Title is required and must be a string' });
  }

  if (!answer || typeof answer !== 'string') {
    return res.status(400).json({ error: 'Answer is required and must be a string' });
  }

  if (!category_id) {
    return res.status(400).json({ error: 'Category ID is required' });
  }

  const category = db.prepare<[number], { id: number }>(`
    SELECT id FROM categories WHERE id = ?
  `).get(category_id);

  if (!category) {
    return res.status(400).json({ error: 'Category does not exist' });
  }

  const stmt = db.prepare(`
    INSERT INTO faqs (title, answer, sort_weight, is_enabled, category_id)
    VALUES (?, ?, ?, ?, ?)
  `);

  const result = stmt.run(
    title,
    answer,
    sort_weight ?? 0,
    is_enabled !== undefined ? (is_enabled ? 1 : 0) : 1,
    category_id
  );

  const id = result.lastInsertRowid as number;
  const faq = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  res.status(201).json({
    ...faq,
    satisfaction: getSatisfaction(id),
  });
});

router.put('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);
  const { title, answer, sort_weight, is_enabled, category_id } = req.body;

  const existing = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'FAQ not found' });
  }

  const updates: string[] = [];
  const values: unknown[] = [];

  if (title !== undefined) {
    updates.push('title = ?');
    values.push(title);
  }

  if (answer !== undefined) {
    updates.push('answer = ?');
    values.push(answer);
  }

  if (sort_weight !== undefined) {
    updates.push('sort_weight = ?');
    values.push(sort_weight);
  }

  if (is_enabled !== undefined) {
    updates.push('is_enabled = ?');
    values.push(is_enabled ? 1 : 0);
  }

  if (category_id !== undefined) {
    const category = db.prepare<[number], { id: number }>(`
      SELECT id FROM categories WHERE id = ?
    `).get(category_id);

    if (!category) {
      return res.status(400).json({ error: 'Category does not exist' });
    }

    updates.push('category_id = ?');
    values.push(category_id);
  }

  if (updates.length > 0) {
    updates.push('updated_at = CURRENT_TIMESTAMP');
    values.push(id);

    const sql = `UPDATE faqs SET ${updates.join(', ')} WHERE id = ?`;
    db.prepare(sql).run(...values);
  }

  const updated = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  res.json({
    ...updated,
    satisfaction: getSatisfaction(id),
  });
});

router.delete('/:id', (req, res) => {
  const id = parseInt(req.params.id, 10);

  const existing = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  if (!existing) {
    return res.status(404).json({ error: 'FAQ not found' });
  }

  db.prepare(`DELETE FROM feedbacks WHERE faq_id = ?`).run(id);
  db.prepare(`DELETE FROM faqs WHERE id = ?`).run(id);

  res.status(204).send();
});

router.post('/:id/feedback', (req, res) => {
  const id = parseInt(req.params.id, 10);
  const { feedback } = req.body;

  const faq = db.prepare<[number], FAQ>(`
    SELECT * FROM faqs WHERE id = ?
  `).get(id);

  if (!faq) {
    return res.status(404).json({ error: 'FAQ not found' });
  }

  if (feedback !== '有帮助' && feedback !== '没帮助') {
    return res.status(400).json({ error: 'Feedback must be "有帮助" or "没帮助"' });
  }

  const isHelpful = feedback === '有帮助' ? 1 : 0;

  db.prepare(`
    INSERT INTO feedbacks (faq_id, is_helpful) VALUES (?, ?)
  `).run(id, isHelpful);

  res.json({
    faq_id: id,
    feedback,
    satisfaction: getSatisfaction(id),
  });
});

export default router;
