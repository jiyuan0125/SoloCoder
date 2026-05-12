import { Router, Request as ExpressRequest, Response } from 'express';
import { generateId, getCurrentTime, getYearMonth, isWithinHours } from '../utils/id';
import { runSQL, getOne, getAll, updateUserAvgRating } from '../db';
import { Category, RequestStatus, VALID_CATEGORIES, HelpRequest, Rating } from '../types';

const router = Router();

router.post('/', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const { publisherId, title, category, description, expectedTime, willingToPay } = req.body;

    if (!title || !category) {
      res.status(400).json({ error: '标题或类别不能为空' });
      return;
    }

    if (!VALID_CATEGORIES.includes(category as Category)) {
      res.status(400).json({ error: '类别不在选项中' });
      return;
    }

    if (!publisherId) {
      res.status(400).json({ error: '发布者ID不能为空' });
      return;
    }

    const publisher = await getOne('SELECT id, building FROM users WHERE id = ?', [publisherId]);
    if (!publisher) {
      res.status(404).json({ error: '发布者不存在' });
      return;
    }

    const id = generateId();
    const now = getCurrentTime();
    
    await runSQL(
      `INSERT INTO requests 
       (id, publisher_id, title, category, description, expected_time, willing_to_pay, status, created_at) 
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [id, publisherId, title, category, description, expectedTime, willingToPay ? 1 : 0, RequestStatus.待接单, now]
    );

    const { year, month } = getYearMonth(now);
    await ensureMonthStatExists(year, month);
    await runSQL(
      'UPDATE month_stats SET publish_count = publish_count + 1 WHERE year = ? AND month = ?',
      [year, month]
    );

    const building = (publisher as any).building;
    await ensureBuildingActivityExists(building, year, month);
    await runSQL(
      'UPDATE building_activities SET publish_count = publish_count + 1 WHERE building = ? AND year = ? AND month = ?',
      [building, year, month]
    );
    await updateBuildingActivityScore(building, year, month);

    const request = await getRequestById(id);
    res.status(201).json(request);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/', async (_req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const rows = await getAll<RequestRow>('SELECT * FROM requests ORDER BY created_at DESC');
    res.json(rows.map(mapToRequest));
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/:id', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const request = await getRequestById(req.params.id);
    if (!request) {
      res.status(404).json({ error: '需求不存在' });
      return;
    }
    res.json(request);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.post('/:id/accept', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const { assigneeId } = req.body;
    if (!assigneeId) {
      res.status(400).json({ error: '接单人ID不能为空' });
      return;
    }

    const assignee = await getOne<{ id: string; building: string; cool_down_until?: number }>(
      'SELECT id, building, cool_down_until FROM users WHERE id = ?',
      [assigneeId]
    );
    if (!assignee) {
      res.status(404).json({ error: '接单人不存在' });
      return;
    }

    const now = getCurrentTime();
    if (assignee.cool_down_until && assignee.cool_down_until > now) {
      res.status(403).json({ error: '取消冷却期' });
      return;
    }

    const ongoingRequest = await getOne(
      `SELECT id FROM requests 
       WHERE assignee_id = ? AND status = ?`,
      [assigneeId, RequestStatus.进行中]
    );
    if (ongoingRequest) {
      res.status(409).json({ error: '请先完成当前需求' });
      return;
    }

    const { db } = await import('../db');
    
    await new Promise<void>((resolve, reject) => {
      db.serialize(() => {
        db.run('BEGIN TRANSACTION');
        
        db.get(
          `SELECT r.id, r.status, u.building 
           FROM requests r 
           JOIN users u ON r.publisher_id = u.id 
           WHERE r.id = ?`,
          [req.params.id],
          (err: Error | null, row: any) => {
            if (err) {
              db.run('ROLLBACK');
              reject(err);
              return;
            }
            
            if (!row) {
              db.run('ROLLBACK');
              res.status(404).json({ error: '需求不存在' });
              resolve();
              return;
            }

            if (row.status !== RequestStatus.待接单) {
              db.run('ROLLBACK');
              res.status(409).json({ error: '需求已被接走' });
              resolve();
              return;
            }

            if (row.assignee_id) {
              db.run('ROLLBACK');
              res.status(409).json({ error: '需求已被接走' });
              resolve();
              return;
            }

            db.run(
              `UPDATE requests SET status = ?, assignee_id = ? WHERE id = ? AND status = ?`,
              [RequestStatus.进行中, assigneeId, req.params.id, RequestStatus.待接单],
              function (updateErr: Error | null) {
                if (updateErr) {
                  db.run('ROLLBACK');
                  reject(updateErr);
                  return;
                }

                if (this.changes === 0) {
                  db.run('ROLLBACK');
                  res.status(409).json({ error: '需求已被接走' });
                  resolve();
                  return;
                }

                const { year, month } = getYearMonth(now);
                ensureMonthStatExists(year, month).then(() => {
                  db.run(
                    'UPDATE month_stats SET match_success_count = match_success_count + 1 WHERE year = ? AND month = ?',
                    [year, month]
                  );
                });

                const { building } = assignee;
                ensureBuildingActivityExists(building, year, month).then(() => {
                  db.run(
                    'UPDATE building_activities SET accept_count = accept_count + 1 WHERE building = ? AND year = ? AND month = ?',
                    [building, year, month],
                    () => {
                      updateBuildingActivityScore(building, year, month);
                    }
                  );
                });

                db.run('COMMIT', async () => {
                  const updatedRequest = await getRequestById(req.params.id);
                  res.json(updatedRequest);
                  resolve();
                });
              }
            );
          }
        );
      });
    });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.post('/:id/cancel', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const request = await getRequestById(req.params.id);
    if (!request) {
      res.status(404).json({ error: '需求不存在' });
      return;
    }

    if (request.status !== RequestStatus.进行中) {
      res.status(400).json({ error: '只有进行中的需求可以取消' });
      return;
    }

    const { assigneeId } = request;
    if (!assigneeId) {
      res.status(400).json({ error: '该需求尚未被接单' });
      return;
    }

    const now = getCurrentTime();
    const isWithin12Hours = isWithinHours(request.expectedTime, 12);

    if (isWithin12Hours) {
      const sevenDays = 7 * 24 * 60 * 60 * 1000;
      await runSQL(
        'UPDATE users SET cool_down_until = ? WHERE id = ?',
        [now + sevenDays, assigneeId]
      );
    }

    await runSQL(
      'UPDATE requests SET status = ?, assignee_id = NULL WHERE id = ?',
      [RequestStatus.待接单, req.params.id]
    );

    if (isWithin12Hours) {
      res.status(403).json({ error: '取消冷却期' });
      return;
    }

    const updatedRequest = await getRequestById(req.params.id);
    res.json(updatedRequest);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.post('/:id/complete', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const request = await getRequestById(req.params.id);
    if (!request) {
      res.status(404).json({ error: '需求不存在' });
      return;
    }

    if (request.status !== RequestStatus.进行中) {
      res.status(400).json({ error: '只有进行中的需求可以完成' });
      return;
    }

    const now = getCurrentTime();
    const completeTime = now - request.createdAt;

    await runSQL(
      'UPDATE requests SET status = ?, completed_at = ? WHERE id = ?',
      [RequestStatus.已完成, now, req.params.id]
    );

    const { year, month } = getYearMonth(now);
    await ensureMonthStatExists(year, month);
    await runSQL(
      'UPDATE month_stats SET complete_count = complete_count + 1, total_complete_time = total_complete_time + ? WHERE year = ? AND month = ?',
      [completeTime, year, month]
    );

    const updatedRequest = await getRequestById(req.params.id);
    res.json(updatedRequest);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.post('/:id/rate', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const { fromUserId, score } = req.body;

    if (!fromUserId || score === undefined) {
      res.status(400).json({ error: '缺少必要参数' });
      return;
    }

    if (score < 1 || score > 5) {
      res.status(400).json({ error: '评分超范围' });
      return;
    }

    const request = await getRequestById(req.params.id);
    if (!request) {
      res.status(404).json({ error: '需求不存在' });
      return;
    }

    if (request.status !== RequestStatus.已完成) {
      res.status(400).json({ error: '只能对已完成的需求进行评价' });
      return;
    }

    if (!request.completedAt) {
      res.status(400).json({ error: '需求完成时间异常' });
      return;
    }

    const now = getCurrentTime();
    const seventyTwoHours = 72 * 60 * 60 * 1000;
    if (now - request.completedAt > seventyTwoHours) {
      res.status(400).json({ error: '评价已超时' });
      return;
    }

    let toUserId: string;
    if (fromUserId === request.publisherId) {
      if (!request.assigneeId) {
        res.status(400).json({ error: '该需求没有接单人' });
        return;
      }
      toUserId = request.assigneeId;
    } else if (fromUserId === request.assigneeId) {
      toUserId = request.publisherId;
    } else {
      res.status(403).json({ error: '只有需求相关人员可以评价' });
      return;
    }

    const existingRating = await getOne(
      'SELECT id FROM ratings WHERE request_id = ? AND from_user_id = ?',
      [req.params.id, fromUserId]
    );
    if (existingRating) {
      res.status(400).json({ error: '已经评价过了' });
      return;
    }

    const id = generateId();
    await runSQL(
      'INSERT INTO ratings (id, request_id, from_user_id, to_user_id, score, created_at) VALUES (?, ?, ?, ?, ?, ?)',
      [id, req.params.id, fromUserId, toUserId, score, now]
    );

    await updateUserAvgRating(toUserId);

    const rating = await getOne<RatingRow>(
      'SELECT * FROM ratings WHERE id = ?',
      [id]
    );
    res.status(201).json(mapToRating(rating!));
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.post('/:id/match', async (req: ExpressRequest, res: Response): Promise<void> => {
  try {
    const { userId } = req.body;
    if (!userId) {
      res.status(400).json({ error: '用户ID不能为空' });
      return;
    }

    const user = await getOne<{ building: string }>(
      'SELECT building FROM users WHERE id = ?',
      [userId]
    );
    if (!user) {
      res.status(404).json({ error: '用户不存在' });
      return;
    }

    const request = await getRequestById(req.params.id);
    if (!request) {
      res.status(404).json({ error: '需求不存在' });
      return;
    }

    const sameBuildingUsers = await getAll<{ id: string; building: string; avg_rating: number; rating_count: number }>(
      `SELECT u.id, u.building, u.avg_rating, u.rating_count 
       FROM users u 
       WHERE u.building = ? AND u.id != ?
       AND NOT EXISTS (
         SELECT 1 FROM requests r 
         WHERE r.assignee_id = u.id 
         AND r.status = ?
       )
       AND (u.cool_down_until IS NULL OR u.cool_down_until < ?)
       ORDER BY u.avg_rating DESC, u.rating_count DESC`,
      [user.building, userId, RequestStatus.进行中, getCurrentTime()]
    );

    const otherBuildingUsers = await getAll<{ id: string; building: string; avg_rating: number; rating_count: number }>(
      `SELECT u.id, u.building, u.avg_rating, u.rating_count 
       FROM users u 
       WHERE u.building != ? AND u.id != ?
       AND NOT EXISTS (
         SELECT 1 FROM requests r 
         WHERE r.assignee_id = u.id 
         AND r.status = ?
       )
       AND (u.cool_down_until IS NULL OR u.cool_down_until < ?)
       ORDER BY u.avg_rating DESC, u.rating_count DESC`,
      [user.building, userId, RequestStatus.进行中, getCurrentTime()]
    );

    const allUsers = [
      ...sameBuildingUsers.map(u => ({ ...u, sameBuilding: true })),
      ...otherBuildingUsers.map(u => ({ ...u, sameBuilding: false })),
    ];

    res.json({
      request,
      recommendedUsers: allUsers,
    });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

interface RequestRow {
  id: string;
  publisher_id: string;
  title: string;
  category: string;
  description: string;
  expected_time: number;
  willing_to_pay: number;
  status: string;
  assignee_id?: string;
  completed_at?: number;
  created_at: number;
  expired_at?: number;
  archived_at?: number;
}

interface RatingRow {
  id: string;
  request_id: string;
  from_user_id: string;
  to_user_id: string;
  score: number;
  created_at: number;
}

function mapToRequest(row: RequestRow): HelpRequest {
  return {
    id: row.id,
    publisherId: row.publisher_id,
    title: row.title,
    category: row.category as Category,
    description: row.description,
    expectedTime: row.expected_time,
    willingToPay: row.willing_to_pay === 1,
    status: row.status as RequestStatus,
    assigneeId: row.assignee_id,
    completedAt: row.completed_at,
    createdAt: row.created_at,
    expiredAt: row.expired_at,
    archivedAt: row.archived_at,
  };
}

function mapToRating(row: RatingRow): Rating {
  return {
    id: row.id,
    requestId: row.request_id,
    fromUserId: row.from_user_id,
    toUserId: row.to_user_id,
    score: row.score,
    createdAt: row.created_at,
  };
}

async function getRequestById(id: string): Promise<HelpRequest | undefined> {
  const row = await getOne<RequestRow>('SELECT * FROM requests WHERE id = ?', [id]);
  return row ? mapToRequest(row) : undefined;
}

async function ensureMonthStatExists(year: number, month: number): Promise<void> {
  const existing = await getOne('SELECT id FROM month_stats WHERE year = ? AND month = ?', [year, month]);
  if (!existing) {
    const id = generateId();
    await runSQL(
      'INSERT INTO month_stats (id, year, month) VALUES (?, ?, ?)',
      [id, year, month]
    );
  }
}

async function ensureBuildingActivityExists(building: string, year: number, month: number): Promise<void> {
  const existing = await getOne(
    'SELECT id FROM building_activities WHERE building = ? AND year = ? AND month = ?',
    [building, year, month]
  );
  if (!existing) {
    const id = generateId();
    await runSQL(
      'INSERT INTO building_activities (id, building, year, month) VALUES (?, ?, ?, ?)',
      [id, building, year, month]
    );
  }
}

async function updateBuildingActivityScore(building: string, year: number, month: number): Promise<void> {
  await runSQL(
    `UPDATE building_activities 
     SET score = (publish_count * 0.6) + (accept_count * 0.4) 
     WHERE building = ? AND year = ? AND month = ?`,
    [building, year, month]
  );
}

export { router as requestsRouter };
