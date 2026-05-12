import { Router, Request, Response } from 'express';
import { getOne, getAll } from '../db';

const router = Router();

router.get('/monthly', async (req: Request, res: Response): Promise<void> => {
  try {
    const { year, month } = req.query;
    
    let rows;
    if (year && month) {
      rows = await getAll(
        'SELECT * FROM month_stats WHERE year = ? AND month = ?',
        [parseInt(year as string), parseInt(month as string)]
      );
    } else {
      rows = await getAll('SELECT * FROM month_stats ORDER BY year DESC, month DESC');
    }

    const stats = rows.map((row: any) => ({
      year: row.year,
      month: row.month,
      publishCount: row.publish_count,
      matchSuccessRate: row.publish_count > 0 
        ? (row.match_success_count / row.publish_count).toFixed(2)
        : '0.00',
      avgCompleteTime: row.complete_count > 0
        ? Math.round(row.total_complete_time / row.complete_count / 60000)
        : 0,
      matchSuccessCount: row.match_success_count,
      completeCount: row.complete_count,
    }));

    res.json(stats);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/building-activity', async (req: Request, res: Response): Promise<void> => {
  try {
    const { building, year, month } = req.query;
    
    let sql = 'SELECT * FROM building_activities';
    const params: any[] = [];
    const conditions: string[] = [];

    if (building) {
      conditions.push('building = ?');
      params.push(building);
    }
    if (year) {
      conditions.push('year = ?');
      params.push(parseInt(year as string));
    }
    if (month) {
      conditions.push('month = ?');
      params.push(parseInt(month as string));
    }

    if (conditions.length > 0) {
      sql += ' WHERE ' + conditions.join(' AND ');
    }
    sql += ' ORDER BY score DESC, publish_count DESC, accept_count DESC';

    const rows = await getAll(sql, params);

    const activities = rows.map((row: any) => ({
      building: row.building,
      year: row.year,
      month: row.month,
      publishCount: row.publish_count,
      acceptCount: row.accept_count,
      score: row.score,
      formula: '(发布数 * 0.6) + (接单 * 0.4)',
    }));

    res.json(activities);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/warnings', async (_req: Request, res: Response): Promise<void> => {
  try {
    const users = await getAll(
      `SELECT id, name, building, avg_rating, rating_count 
       FROM users 
       WHERE avg_rating < 3.0 AND rating_count >= 10
       ORDER BY avg_rating ASC, rating_count DESC`
    );

    const warnings = users.map((user: any) => ({
      userId: user.id,
      userName: user.name,
      building: user.building,
      avgRating: user.avg_rating,
      ratingCount: user.rating_count,
      warning: '平均评分低于3.0且评价超10条，请注意！',
    }));

    res.json(warnings);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

export { router as statsRouter };
