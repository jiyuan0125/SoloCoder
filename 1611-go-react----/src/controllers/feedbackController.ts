import { Request, Response } from 'express';
import { db } from '../database';
import { Feedback, FeedbackStatus, Priority, CategoryType, StatusLog } from '../types';
import { validateTitle, validateDescription, determinePriority, isValidStatusTransition, getAllowedTransitions, getPriorityOrder, sendNotification, logInternalMessage, getDaysBetween, getMinutesBetween } from '../utils';

interface DBFeedback {
  id: number;
  user_id: number;
  title: string;
  description: string;
  category_id: number;
  priority: string;
  status: string;
  assignee_id: number | null;
  internal_note: string;
  is_timeout: number;
  resolved_at: string | null;
  created_at: string;
  updated_at: string;
}

interface DBStatusLog {
  id: number;
  feedback_id: number;
  from_status: string;
  to_status: string;
  operator_id: number;
  operator_role: string;
  created_at: string;
}

function mapFeedback(dbFeedback: DBFeedback): Feedback {
  return {
    id: dbFeedback.id,
    userId: dbFeedback.user_id,
    title: dbFeedback.title,
    description: dbFeedback.description,
    categoryId: dbFeedback.category_id,
    priority: dbFeedback.priority as Priority,
    status: dbFeedback.status as FeedbackStatus,
    assigneeId: dbFeedback.assignee_id,
    internalNote: dbFeedback.internal_note,
    isTimeout: dbFeedback.is_timeout === 1,
    resolvedAt: dbFeedback.resolved_at,
    createdAt: dbFeedback.created_at,
    updatedAt: dbFeedback.updated_at
  };
}

function mapStatusLog(dbLog: DBStatusLog): StatusLog {
  return {
    id: dbLog.id,
    feedbackId: dbLog.feedback_id,
    fromStatus: dbLog.from_status as FeedbackStatus,
    toStatus: dbLog.to_status as FeedbackStatus,
    operatorId: dbLog.operator_id,
    operatorRole: dbLog.operator_role as 'admin' | 'user',
    createdAt: dbLog.created_at
  };
}

export function createFeedback(req: Request, res: Response): void {
  try {
    const { userId, title, description, categoryId } = req.body as {
      userId: number;
      title: string;
      description: string;
      categoryId: number;
    };

    if (!userId || typeof userId !== 'number') {
      res.status(400).json({ error: '用户ID不能为空' });
      return;
    }

    if (!validateTitle(title)) {
      res.status(400).json({ error: '标题长度必须在5到100个字符之间' });
      return;
    }

    if (!validateDescription(description)) {
      res.status(400).json({ error: '描述长度必须在20到2000个字符之间' });
      return;
    }

    if (!categoryId || typeof categoryId !== 'number') {
      res.status(400).json({ error: '分类ID不能为空' });
      return;
    }

    const category = db
      .prepare('SELECT name FROM categories WHERE id = ?')
      .get(categoryId) as { name: string } | undefined;

    if (!category) {
      res.status(400).json({ error: '无效的分类ID' });
      return;
    }

    const tenMinutesAgo = new Date(Date.now() - 10 * 60 * 1000).toISOString();
    const duplicate = db
      .prepare('SELECT id FROM feedbacks WHERE user_id = ? AND title = ? AND created_at >= ?')
      .get(userId, title, tenMinutesAgo);

    if (duplicate) {
      res.status(409).json({ error: '10分钟内不能提交相同标题的反馈' });
      return;
    }

    const priority = determinePriority(title, description, category.name as CategoryType);
    const now = new Date().toISOString();

    const result = db.prepare(`
      INSERT INTO feedbacks 
      (user_id, title, description, category_id, priority, status, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      userId,
      title,
      description,
      categoryId,
      priority,
      FeedbackStatus.NEW,
      now,
      now
    );

    const feedbackId = result.lastInsertRowid as number;

    if (req.files && Array.isArray(req.files)) {
      const insertAttachment = db.prepare(`
        INSERT INTO attachments (feedback_id, file_name, file_path, file_size, created_at)
        VALUES (?, ?, ?, ?, ?)
      `);
      for (const file of req.files as Express.Multer.File[]) {
        insertAttachment.run(
          feedbackId,
          file.originalname,
          file.path,
          file.size,
          now
        );
      }
    }

    const feedback = db
      .prepare('SELECT * FROM feedbacks WHERE id = ?')
      .get(feedbackId) as DBFeedback;

    res.status(201).json(mapFeedback(feedback));
  } catch (error) {
    res.status(500).json({ error: '创建反馈失败' });
  }
}

export function getAllFeedbacks(_req: Request, res: Response): void {
  try {
    const feedbacks = db
      .prepare('SELECT * FROM feedbacks')
      .all() as DBFeedback[];

    const mappedFeedbacks = feedbacks.map(mapFeedback);

    mappedFeedbacks.sort((a, b) => {
      const priorityDiff = getPriorityOrder(a.priority) - getPriorityOrder(b.priority);
      if (priorityDiff !== 0) return priorityDiff;
      return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
    });

    res.json(mappedFeedbacks);
  } catch (error) {
    res.status(500).json({ error: '获取反馈列表失败' });
  }
}

export function getFeedbackById(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的反馈ID' });
      return;
    }

    const feedback = db
      .prepare('SELECT * FROM feedbacks WHERE id = ?')
      .get(id) as DBFeedback | undefined;

    if (!feedback) {
      res.status(404).json({ error: '反馈不存在' });
      return;
    }

    res.json(mapFeedback(feedback));
  } catch (error) {
    res.status(500).json({ error: '获取反馈详情失败' });
  }
}

export function updatePriority(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);
    const { priority } = req.body as { priority: Priority };

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的反馈ID' });
      return;
    }

    const validPriorities = Object.values(Priority);
    if (!validPriorities.includes(priority)) {
      res.status(400).json({ 
        error: '无效的优先级', 
        allowed: validPriorities 
      });
      return;
    }

    const existing = db
      .prepare('SELECT id FROM feedbacks WHERE id = ?')
      .get(id);

    if (!existing) {
      res.status(404).json({ error: '反馈不存在' });
      return;
    }

    const now = new Date().toISOString();
    db
      .prepare('UPDATE feedbacks SET priority = ?, updated_at = ? WHERE id = ?')
      .run(priority, now, id);

    const feedback = db
      .prepare('SELECT * FROM feedbacks WHERE id = ?')
      .get(id) as DBFeedback;

    res.json(mapFeedback(feedback));
  } catch (error) {
    res.status(500).json({ error: '更新优先级失败' });
  }
}

export function updateStatus(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);
    const { toStatus, operatorId, operatorRole } = req.body as {
      toStatus: FeedbackStatus;
      operatorId: number;
      operatorRole: 'admin' | 'user';
    };

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的反馈ID' });
      return;
    }

    if (!operatorId || typeof operatorId !== 'number') {
      res.status(400).json({ error: '操作人ID不能为空' });
      return;
    }

    if (operatorRole !== 'admin' && operatorRole !== 'user') {
      res.status(400).json({ error: '无效的操作人角色' });
      return;
    }

    const feedback = db
      .prepare('SELECT * FROM feedbacks WHERE id = ?')
      .get(id) as DBFeedback | undefined;

    if (!feedback) {
      res.status(404).json({ error: '反馈不存在' });
      return;
    }

    const fromStatus = feedback.status as FeedbackStatus;

    if (toStatus === FeedbackStatus.CLOSED && fromStatus === FeedbackStatus.RESOLVED) {
      if (operatorRole !== 'user' || operatorId !== feedback.user_id) {
        res.status(400).json({ error: '已解决状态只能由提交用户确认关闭' });
        return;
      }
    }

    if (!isValidStatusTransition(fromStatus, toStatus)) {
      const allowed = getAllowedTransitions(fromStatus);
      res.status(400).json({ 
        error: `状态流转无效，从"${fromStatus}"只能跳转到`,
        allowedTransitions: allowed
      });
      return;
    }

    const now = new Date().toISOString();
    const transaction = db.transaction(() => {
      db
        .prepare('UPDATE feedbacks SET status = ?, updated_at = ?, resolved_at = ? WHERE id = ?')
        .run(
          toStatus,
          now,
          toStatus === FeedbackStatus.RESOLVED ? now : feedback.resolved_at,
          id
        );

      db
        .prepare(`
          INSERT INTO status_logs 
          (feedback_id, from_status, to_status, operator_id, operator_role, created_at)
          VALUES (?, ?, ?, ?, ?, ?)
        `)
        .run(id, fromStatus, toStatus, operatorId, operatorRole, now);
    });

    transaction();

    const updatedFeedback = db
      .prepare('SELECT * FROM feedbacks WHERE id = ?')
      .get(id) as DBFeedback;

    res.json(mapFeedback(updatedFeedback));
  } catch (error) {
    res.status(500).json({ error: '更新状态失败' });
  }
}

export function getStatusLogs(req: Request, res: Response): void {
  try {
    const feedbackId = parseInt(req.params.id, 10);

    if (isNaN(feedbackId)) {
      res.status(400).json({ error: '无效的反馈ID' });
      return;
    }

    const logs = db
      .prepare('SELECT * FROM status_logs WHERE feedback_id = ? ORDER BY created_at DESC')
      .all(feedbackId) as DBStatusLog[];

    res.json(logs.map(mapStatusLog));
  } catch (error) {
    res.status(500).json({ error: '获取状态历史失败' });
  }
}

export function deleteFeedback(req: Request, res: Response): void {
  try {
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      res.status(400).json({ error: '无效的反馈ID' });
      return;
    }

    const feedback = db
      .prepare('SELECT internal_note FROM feedbacks WHERE id = ?')
      .get(id) as { internal_note: string } | undefined;

    if (!feedback) {
      res.status(404).json({ error: '反馈不存在' });
      return;
    }

    if (feedback.internal_note && feedback.internal_note.trim().length > 0) {
      res.status(409).json({ error: '存在未处理的内部备注，无法删除' });
      return;
    }

    const transaction = db.transaction(() => {
      db
        .prepare('DELETE FROM attachments WHERE feedback_id = ?')
        .run(id);
      db
        .prepare('DELETE FROM status_logs WHERE feedback_id = ?')
        .run(id);
      db
        .prepare('DELETE FROM feedbacks WHERE id = ?')
        .run(id);
    });

    transaction();

    res.status(204).send();
  } catch (error) {
    res.status(500).json({ error: '删除反馈失败' });
  }
}

export function batchClaim(req: Request, res: Response): void {
  try {
    const { feedbackIds, assigneeId } = req.body as {
      feedbackIds: number[];
      assigneeId: number;
    };

    if (!feedbackIds || !Array.isArray(feedbackIds) || feedbackIds.length === 0) {
      res.status(400).json({ error: '反馈ID列表不能为空' });
      return;
    }

    if (!assigneeId || typeof assigneeId !== 'number') {
      res.status(400).json({ error: '处理人ID不能为空' });
      return;
    }

    const results: { id: number; success: boolean; error?: string }[] = [];
    const now = new Date().toISOString();

    for (const id of feedbackIds) {
      try {
        const feedback = db
          .prepare('SELECT id FROM feedbacks WHERE id = ?')
          .get(id);

        if (!feedback) {
          results.push({ id, success: false, error: '反馈不存在' });
          continue;
        }

        db
          .prepare('UPDATE feedbacks SET assignee_id = ?, updated_at = ? WHERE id = ?')
          .run(assigneeId, now, id);

        results.push({ id, success: true });
      } catch (error) {
        results.push({ id, success: false, error: '领取失败' });
      }
    }

    res.json(results);
  } catch (error) {
    res.status(500).json({ error: '批量领取失败' });
  }
}

export function batchResolve(req: Request, res: Response): void {
  try {
    const { feedbackIds, operatorId } = req.body as {
      feedbackIds: number[];
      operatorId: number;
    };

    if (!feedbackIds || !Array.isArray(feedbackIds) || feedbackIds.length === 0) {
      res.status(400).json({ error: '反馈ID列表不能为空' });
      return;
    }

    if (!operatorId || typeof operatorId !== 'number') {
      res.status(400).json({ error: '操作人ID不能为空' });
      return;
    }

    const results: { id: number; success: boolean; error?: string }[] = [];
    const now = new Date().toISOString();

    for (const id of feedbackIds) {
      try {
        const feedback = db
          .prepare('SELECT * FROM feedbacks WHERE id = ?')
          .get(id) as DBFeedback | undefined;

        if (!feedback) {
          results.push({ id, success: false, error: '反馈不存在' });
          continue;
        }

        const fromStatus = feedback.status as FeedbackStatus;

        if (fromStatus !== FeedbackStatus.PROCESSING) {
          results.push({ 
            id, 
            success: false, 
            error: `状态"${fromStatus}"不能直接标记为已解决，当前允许的操作: ${getAllowedTransitions(fromStatus).join('、') || '无'}` 
          });
          continue;
        }

        const transaction = db.transaction(() => {
          db
            .prepare('UPDATE feedbacks SET status = ?, resolved_at = ?, updated_at = ? WHERE id = ?')
            .run(FeedbackStatus.RESOLVED, now, now, id);

          db
            .prepare(`
              INSERT INTO status_logs 
              (feedback_id, from_status, to_status, operator_id, operator_role, created_at)
              VALUES (?, ?, ?, ?, ?, ?)
            `)
            .run(id, fromStatus, FeedbackStatus.RESOLVED, operatorId, 'admin', now);
        });

        transaction();
        results.push({ id, success: true });
      } catch (error) {
        results.push({ id, success: false, error: '标记已解决失败' });
      }
    }

    res.json(results);
  } catch (error) {
    res.status(500).json({ error: '批量标记已解决失败' });
  }
}

export async function runScheduledTasks(): Promise<void> {
  try {
    const now = new Date();
    const nowISO = now.toISOString();

    const sevenDaysAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    const sevenDaysAgoISO = sevenDaysAgo.toISOString();

    const resolvedFeedbacks = db
      .prepare('SELECT * FROM feedbacks WHERE status = ? AND resolved_at < ?')
      .all(FeedbackStatus.RESOLVED, sevenDaysAgoISO) as DBFeedback[];

    for (const feedback of resolvedFeedbacks) {
      const userId = feedback.user_id;
      const notificationResult = await sendNotification(
        userId,
        `您的反馈 #${feedback.id}（${feedback.title}）已自动关闭，因为您超过7天未确认。`
      );

      if (!notificationResult.success) {
        logInternalMessage(`无法向用户 ${userId} 发送自动关闭通知，用户账号可能已注销。反馈ID: ${feedback.id}`);
      }

      const transaction = db.transaction(() => {
        db
          .prepare('UPDATE feedbacks SET status = ?, updated_at = ? WHERE id = ?')
          .run(FeedbackStatus.CLOSED, nowISO, feedback.id);

        db
          .prepare(`
            INSERT INTO status_logs 
            (feedback_id, from_status, to_status, operator_id, operator_role, created_at)
            VALUES (?, ?, ?, ?, ?, ?)
          `)
          .run(feedback.id, FeedbackStatus.RESOLVED, FeedbackStatus.CLOSED, 0, 'admin', nowISO);
      });

      transaction();
    }

    const openFeedbacks = db
      .prepare('SELECT * FROM feedbacks WHERE status NOT IN (?, ?) AND created_at < ? AND is_timeout = 0')
      .all(FeedbackStatus.RESOLVED, FeedbackStatus.CLOSED, sevenDaysAgoISO) as DBFeedback[];

    for (const feedback of openFeedbacks) {
      db
        .prepare('UPDATE feedbacks SET is_timeout = 1, updated_at = ? WHERE id = ?')
        .run(nowISO, feedback.id);
    }

    logInternalMessage(`定时任务执行完成。自动关闭: ${resolvedFeedbacks.length} 条，超时标记: ${openFeedbacks.length} 条`);
  } catch (error) {
    logInternalMessage('定时任务执行失败: ' + (error instanceof Error ? error.message : String(error)));
  }
}
