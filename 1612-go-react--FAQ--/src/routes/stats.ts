import express from 'express';
import db from '../db';

const router = express.Router();

router.get('/search-no-result-rate', (req, res) => {
  const stats = db.prepare<[], { total: number; no_results: number }>(`
    SELECT
      COUNT(*) as total,
      SUM(CASE WHEN has_results = 0 THEN 1 ELSE 0 END) as no_results
    FROM search_stats
  `).get();

  const total = stats?.total || 0;
  const noResults = stats?.no_results || 0;
  const rate = total > 0 ? (noResults / total) * 100 : 0;

  res.json({
    total_searches: total,
    no_result_searches: noResults,
    no_result_rate: rate,
  });
});

export default router;
