import express, { Request, Response } from 'express';
import db from '../db';
import { generateUUID } from '../utils/uuid';
import { calculateP99 } from '../utils/perentile';

const router = express.Router();

router.post('/', (req: Request, res: Response) => {
  const { serviceId, startTime, endTime, description, isAutomatic } = req.body;

  if (!serviceId || !startTime) {
    return res.status(400).json({ error: 'serviceId and startTime are required' });
  }

  const service = db.prepare('SELECT * FROM services WHERE id = ?').get(serviceId);
  if (!service) {
    return res.status(404).json({ error: 'service not found' });
  }

  const existingIncident = db.prepare(`
    SELECT * FROM incidents WHERE service_id = ? AND start_time = ?
  `).get(serviceId, startTime);
  
  if (existingIncident) {
    return res.status(409).json({ error: 'duplicate incident in SLA breach' });
  }

  let durationMinutes = 0;
  if (endTime) {
    const start = new Date(startTime);
    const end = new Date(endTime);
    durationMinutes = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60));
  }

  if (durationMinutes <= 0) {
    return res.status(400).json({ error: 'incident duration must be positive' });
  }

  const now = new Date().toISOString();
  const incidentId = generateUUID();

  const insertStmt = db.prepare(`
    INSERT INTO incidents (id, service_id, start_time, end_time, duration_minutes, description, is_automatic, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);

  insertStmt.run(
    incidentId,
    serviceId,
    startTime,
    endTime || null,
    durationMinutes,
    description || null,
    isAutomatic ? 1 : 0,
    now
  );
  
  const startDate = new Date(startTime);
  const year = startDate.getFullYear();
  const month = startDate.getMonth() + 1;
  
  const slaDefinition = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(serviceId) as any;
  
  if (slaDefinition) {
    let existingReport = db.prepare(`
      SELECT * FROM reports 
      WHERE service_id = ? AND year = ? AND month = ?
    `).get(serviceId, year, month) as any;
    
    if (existingReport) {
      const { start: monthStart, end: monthEnd } = getMonthRange(year, month);
      
      const totalMinutes = getMonthTotalMinutes(year, month);
      const incidents = db.prepare(`
        SELECT * FROM incidents 
        WHERE service_id = ? 
        AND start_time >= ? 
        AND start_time <= ?
      `).all(serviceId, monthStart.toISOString(), monthEnd.toISOString()) as any[];
      
      const totalIncidentMinutes = incidents.reduce((sum, i) => sum + (i.duration_minutes || 0), 0);
      const availabilityAchieved = ((totalMinutes - totalIncidentMinutes) / totalMinutes) * 100;
      
      const metrics = db.prepare(`
        SELECT * FROM metrics 
        WHERE service_id = ? 
        AND timestamp >= ? 
        AND timestamp <= ?
      `).all(serviceId, monthStart.toISOString(), monthEnd.toISOString()) as any[];
      
      const responseTimes = metrics.map(m => m.response_time);
      const responseTimeP99 = Math.round(calculateP99(responseTimes));
      
      const totalRequests = metrics.length;
      const errorCount = metrics.filter(m => m.is_error).length;
      const errorRate = totalRequests > 0 ? (errorCount / totalRequests) * 100 : 0;
      
      const details = db.prepare('SELECT * FROM report_details WHERE report_id = ?').all(existingReport.id) as any[];
      
      const breachTypes: string[] = [];
      
      if (availabilityAchieved < slaDefinition.availability_target) {
        const availabilityDetail = details.find(d => d.metric_type === 'availability');
        if (!availabilityDetail || !availabilityDetail.is_confirmed) {
          breachTypes.push('availability');
        }
      }
      
      if (responseTimeP99 > slaDefinition.response_time_target) {
        const responseTimeDetail = details.find(d => d.metric_type === 'responseTime');
        if (!responseTimeDetail || !responseTimeDetail.is_confirmed) {
          breachTypes.push('responseTime');
        }
      }
      
      if (errorRate > slaDefinition.error_rate_target) {
        const errorRateDetail = details.find(d => d.metric_type === 'errorRate');
        if (!errorRateDetail || !errorRateDetail.is_confirmed) {
          breachTypes.push('errorRate');
        }
      }
      
      for (const metricType of breachTypes) {
        try {
          const breachId = generateUUID();
          db.prepare(`
            INSERT INTO breaches (id, report_id, incident_id, metric_type, created_at)
            VALUES (?, ?, ?, ?, ?)
          `).run(
            breachId,
            existingReport.id,
            incidentId,
            metricType,
            now
          );
        } catch (err: any) {
          if (err.code === 'SQLITE_CONSTRAINT_UNIQUE') {
            return res.status(409).json({ error: 'duplicate incident in SLA breach' });
          }
          throw err;
        }
      }
    }
  }

  const result = db.prepare('SELECT * FROM incidents WHERE id = ?').get(incidentId) as any;
  
  return res.status(201).json({
    id: result.id,
    serviceId: result.service_id,
    startTime: result.start_time,
    endTime: result.end_time,
    durationMinutes: result.duration_minutes,
    description: result.description,
    isAutomatic: !!result.is_automatic,
    createdAt: result.created_at
  });
});

const getMonthTotalMinutes = (year: number, month: number): number => {
  const daysInMonth = new Date(year, month, 0).getDate();
  return daysInMonth * 24 * 60;
};

const getMonthRange = (year: number, month: number): { start: Date; end: Date } => {
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0, 23, 59, 59, 999);
  return { start, end };
};

export default router;
