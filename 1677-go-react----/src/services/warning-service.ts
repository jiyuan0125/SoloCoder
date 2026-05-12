import { db } from '../database';
import { v4 as uuidv4 } from 'uuid';
import dayjs from 'dayjs';
import { Project, ProjectWarning } from '../types';

export function checkAndGenerateProjectWarnings() {
  const now = dayjs();
  const sevenDaysLater = now.add(7, 'day');
  const sevenDaysLaterStr = sevenDaysLater.format('YYYY-MM-DD');
  const nowStr = now.format('YYYY-MM-DD');
  
  const projects = db.prepare(
    `SELECT * FROM projects 
     WHERE end_date >= ? AND end_date <= ?`
  ).all(nowStr, sevenDaysLaterStr) as Project[];
  
  for (const project of projects) {
    const donated = db.prepare(
      'SELECT COALESCE(SUM(amount), 0) as total FROM donations WHERE project_id = ?'
    ).get(project.id) as { total: number };
    
    const target60 = Math.floor(project.target_amount * 0.6);
    if (donated.total < target60) {
      const existingWarning = db.prepare(
        `SELECT * FROM project_warnings 
         WHERE project_id = ? AND warning_type = 'funding_shortage'`
      ).get(project.id);
      
      if (!existingWarning) {
        const id = uuidv4();
        const createdAt = now.format('YYYY-MM-DD HH:mm:ss');
        const message = `项目"${project.name}"距离截止日期不足7天，当前捐赠金额${donated.total}分，未达目标金额的60%（${target60}分）`;
        
        db.prepare(
          `INSERT INTO project_warnings (id, project_id, warning_type, message, created_at)
           VALUES (?, ?, 'funding_shortage', ?, ?)`
        ).run(id, project.id, message, createdAt);
      }
    }
  }
}

export function getWarnings(projectId?: string) {
  if (projectId) {
    return db.prepare(
      'SELECT * FROM project_warnings WHERE project_id = ? ORDER BY created_at DESC'
    ).all(projectId) as ProjectWarning[];
  }
  return db.prepare(
    'SELECT * FROM project_warnings ORDER BY created_at DESC'
  ).all() as ProjectWarning[];
}
