import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Link, CreateLinkRequest } from '../types';

export class LinkService {
  createLink(data: CreateLinkRequest): Link {
    // 检查渠道是否存在
    const channel = db.prepare('SELECT id FROM channels WHERE id = ?').get(data.channel_id);
    if (!channel) {
      throw new Error('渠道不存在');
    }

    // 生成唯一标识码
    const uniqueCode = uuidv4().replace(/-/g, '').substring(0, 12).toUpperCase();
    
    const id = uuidv4();
    const link: Link = {
      id,
      channel_id: data.channel_id,
      material_type: data.material_type,
      unique_code: uniqueCode,
      created_at: new Date().toISOString(),
    };

    const stmt = db.prepare(`
      INSERT INTO links (id, channel_id, material_type, unique_code, created_at)
      VALUES (?, ?, ?, ?, ?)
    `);

    stmt.run(
      link.id,
      link.channel_id,
      link.material_type,
      link.unique_code,
      link.created_at
    );

    return link;
  }

  getLinkById(id: string): Link | undefined {
    return db.prepare('SELECT * FROM links WHERE id = ?').get(id) as Link | undefined;
  }

  getLinkByUniqueCode(uniqueCode: string): Link | undefined {
    return db.prepare('SELECT * FROM links WHERE unique_code = ?').get(uniqueCode) as Link | undefined;
  }

  getLinksByChannelId(channelId: string): Link[] {
    return db.prepare('SELECT * FROM links WHERE channel_id = ? ORDER BY created_at DESC').all(channelId) as Link[];
  }

  getAllLinks(): Link[] {
    return db.prepare('SELECT * FROM links ORDER BY created_at DESC').all() as Link[];
  }
}

export const linkService = new LinkService();
