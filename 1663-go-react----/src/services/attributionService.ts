import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Attribution, TrackAttributionRequest, Link } from '../types';

const ATTRIBUTION_WINDOW_DAYS = 7;

export class AttributionService {
  trackAttribution(data: TrackAttributionRequest): Attribution {
    // 验证用户ID和链接ID
    if (!data.user_id || !data.link_id) {
      throw new Error('用户ID和链接ID不能为空');
    }

    // 检查链接是否存在
    const link = db.prepare('SELECT * FROM links WHERE id = ?').get(data.link_id) as Link | undefined;
    if (!link) {
      throw new Error('推广链接不存在');
    }

    // 计算归因窗口结束时间
    const clickTime = new Date();
    const attributionWindowEnd = new Date(clickTime);
    attributionWindowEnd.setDate(attributionWindowEnd.getDate() + ATTRIBUTION_WINDOW_DAYS);

    const id = uuidv4();
    const attribution: Attribution = {
      id,
      user_id: data.user_id,
      channel_id: link.channel_id,
      link_id: data.link_id,
      click_time: clickTime.toISOString(),
      attribution_window_end: attributionWindowEnd.toISOString(),
      created_at: new Date().toISOString(),
    };

    const stmt = db.prepare(`
      INSERT INTO attributions (id, user_id, channel_id, link_id, click_time, attribution_window_end, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      attribution.id,
      attribution.user_id,
      attribution.channel_id,
      attribution.link_id,
      attribution.click_time,
      attribution.attribution_window_end,
      attribution.created_at
    );

    return attribution;
  }

  getActiveAttribution(userId: string): Attribution | undefined {
    const now = new Date().toISOString();
    
    // 获取用户的最后一次有效归因（在归因窗口内）
    return db.prepare(`
      SELECT * FROM attributions 
      WHERE user_id = ? AND attribution_window_end >= ?
      ORDER BY click_time DESC
      LIMIT 1
    `).get(userId, now) as Attribution | undefined;
  }

  getAttributionsByUserId(userId: string): Attribution[] {
    return db.prepare(`
      SELECT * FROM attributions 
      WHERE user_id = ? 
      ORDER BY click_time DESC
    `).all(userId) as Attribution[];
  }
}

export const attributionService = new AttributionService();
