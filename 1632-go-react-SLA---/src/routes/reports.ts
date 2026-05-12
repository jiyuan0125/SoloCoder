import express, { Request, Response } from 'express';
import db from '../db';
import { generateUUID } from '../utils/uuid';
import { calculateP99 } from '../utils/perentile';

const router = express.Router();

const getMonthTotalMinutes = (year: number, month: number): number => {
  const daysInMonth = new Date(year, month, 0).getDate();
  return daysInMonth * 24 * 60;
};

const getMonthRange = (year: number, month: number): { start: Date; end: Date } => {
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0, 23, 59, 59, 999);
  return { start, end };
};

const calculateReport = (
  serviceId: string,
  slaDefinition: any,
  year: number,
  month: number,
  existingReportDetails?: any[]
) => {
  const { start, end } = getMonthRange(year, month);
  const totalMinutes = getMonthTotalMinutes(year, month);
  
  const incidents = db.prepare(`
    SELECT * FROM incidents 
    WHERE service_id = ? 
    AND start_time >= ? 
    AND start_time <= ?
  `).all(serviceId, start.toISOString(), end.toISOString()) as any[];
  
  const totalIncidentMinutes = incidents.reduce((sum, i) => sum + (i.duration_minutes || 0), 0);
  const availabilityAchieved = ((totalMinutes - totalIncidentMinutes) / totalMinutes) * 100;
  
  const metrics = db.prepare(`
    SELECT * FROM metrics 
    WHERE service_id = ? 
    AND timestamp >= ? 
    AND timestamp <= ?
  `).all(serviceId, start.toISOString(), end.toISOString()) as any[];
  
  const responseTimes = metrics.map(m => m.response_time);
  const responseTimeP99 = Math.round(calculateP99(responseTimes));
  
  const totalRequests = metrics.length;
  const errorCount = metrics.filter(m => m.is_error).length;
  const errorRate = totalRequests > 0 ? (errorCount / totalRequests) * 100 : 0;
  
  const details: Array<{
    type: 'availability' | 'responseTime' | 'errorRate';
    target: number;
    actual: number;
    isAchieved: boolean;
    isConfirmed: boolean;
  }> = [];
  
  const isAvailabilityAchieved = availabilityAchieved >= slaDefinition.availability_target;
  const isResponseTimeAchieved = responseTimeP99 <= slaDefinition.response_time_target;
  const isErrorRateAchieved = errorRate <= slaDefinition.error_rate_target;
  
  const getConfirmed = (type: string, defaultValue: boolean) => {
    if (!existingReportDetails) return defaultValue;
    const existing = existingReportDetails.find(d => d.metric_type === type);
    return existing ? !!existing.is_confirmed : defaultValue;
  };
  
  details.push({
    type: 'availability',
    target: slaDefinition.availability_target,
    actual: availabilityAchieved,
    isAchieved: isAvailabilityAchieved,
    isConfirmed: getConfirmed('availability', false)
  });
  
  details.push({
    type: 'responseTime',
    target: slaDefinition.response_time_target,
    actual: responseTimeP99,
    isAchieved: isResponseTimeAchieved,
    isConfirmed: getConfirmed('responseTime', false)
  });
  
  details.push({
    type: 'errorRate',
    target: slaDefinition.error_rate_target,
    actual: errorRate,
    isAchieved: isErrorRateAchieved,
    isConfirmed: getConfirmed('errorRate', false)
  });
  
  const achievedCount = details.filter(d => d.isAchieved).length;
  const complianceRate = (achievedCount / details.length) * 100;
  
  return {
    availabilityAchieved,
    responseTimeP99,
    errorRate,
    details,
    incidents,
    complianceRate
  };
};

router.get('/:year/:month', (req: Request, res: Response) => {
  const { year, month } = req.params;
  const yearNum = parseInt(year, 10);
  const monthNum = parseInt(month, 10);
  
  if (isNaN(yearNum) || isNaN(monthNum) || monthNum < 1 || monthNum > 12) {
    return res.status(400).json({ error: 'invalid year or month' });
  }
  
  const services = db.prepare('SELECT * FROM services').all() as any[];
  const reports: any[] = [];
  
  for (const service of services) {
    const slaDefinition = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(service.id) as any;
    
    if (!slaDefinition) {
      continue;
    }
    
    let existingReport = db.prepare(`
      SELECT * FROM reports 
      WHERE service_id = ? AND year = ? AND month = ?
    `).get(service.id, yearNum, monthNum) as any;
    
    let existingDetails: any[] = [];
    if (existingReport) {
      existingDetails = db.prepare('SELECT * FROM report_details WHERE report_id = ?').all(existingReport.id) as any[];
    }
    
    const calc = calculateReport(service.id, slaDefinition, yearNum, monthNum, existingDetails);
    
    const confirmedDetails = calc.details.map(d => {
      if (existingReport && d.isConfirmed) {
        const existing = existingDetails.find(ed => ed.metric_type === d.type);
        if (existing) {
          return {
            ...d,
            actual: existing.actual,
            isAchieved: !!existing.is_achieved
          };
        }
      }
      return d;
    });
    
    const breaches: any[] = [];
    if (existingReport) {
      const dbBreaches = db.prepare('SELECT * FROM breaches WHERE report_id = ?').all(existingReport.id) as any[];
      breaches.push(...dbBreaches.map(b => ({
        id: b.id,
        reportId: b.report_id,
        incidentId: b.incident_id,
        metricType: b.metric_type,
        createdAt: b.created_at
      })));
    }
    
    const allMetrics = {
      availability: {
        target: slaDefinition.availability_target,
        actual: confirmedDetails.find(d => d.type === 'availability')!.actual,
        achieved: confirmedDetails.find(d => d.type === 'availability')!.isAchieved
      },
      responseTime: {
        target: slaDefinition.response_time_target,
        actual: confirmedDetails.find(d => d.type === 'responseTime')!.actual,
        achieved: confirmedDetails.find(d => d.type === 'responseTime')!.isAchieved
      },
      errorRate: {
        target: slaDefinition.error_rate_target,
        actual: confirmedDetails.find(d => d.type === 'errorRate')!.actual,
        achieved: confirmedDetails.find(d => d.type === 'errorRate')!.isAchieved
      }
    };
    
    const finalAchievedCount = Object.values(allMetrics).filter(m => m.achieved).length;
    const finalComplianceRate = (finalAchievedCount / 3) * 100;
    
    reports.push({
      serviceId: service.id,
      serviceName: service.name,
      year: yearNum,
      month: monthNum,
      slaDefinition: {
        availabilityTarget: slaDefinition.availability_target,
        responseTimeTarget: slaDefinition.response_time_target,
        errorRateTarget: slaDefinition.error_rate_target
      },
      achieved: {
        availability: allMetrics.availability.actual,
        responseTimeP99: allMetrics.responseTime.actual,
        errorRate: allMetrics.errorRate.actual
      },
      metrics: allMetrics,
      incidents: calc.incidents.map(i => ({
        id: i.id,
        serviceId: i.service_id,
        startTime: i.start_time,
        endTime: i.end_time,
        durationMinutes: i.duration_minutes,
        description: i.description,
        isAutomatic: !!i.is_automatic,
        createdAt: i.created_at
      })),
      complianceRate: finalComplianceRate,
      breaches,
      isConfirmed: existingReport ? !!existingReport.is_confirmed : false
    });
  }
  
  return res.json(reports);
});

router.put('/:year/:month/recalc', (req: Request, res: Response) => {
  const { year, month } = req.params;
  const yearNum = parseInt(year, 10);
  const monthNum = parseInt(month, 10);
  
  if (isNaN(yearNum) || isNaN(monthNum) || monthNum < 1 || monthNum > 12) {
    return res.status(400).json({ error: 'invalid year or month' });
  }
  
  const now = new Date().toISOString();
  const services = db.prepare('SELECT * FROM services').all() as any[];
  const updatedReports: any[] = [];
  
  for (const service of services) {
    const slaDefinition = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(service.id) as any;
    
    if (!slaDefinition) {
      continue;
    }
    
    let existingReport = db.prepare(`
      SELECT * FROM reports 
      WHERE service_id = ? AND year = ? AND month = ?
    `).get(service.id, yearNum, monthNum) as any;
    
    let existingDetails: any[] = [];
    if (existingReport) {
      existingDetails = db.prepare('SELECT * FROM report_details WHERE report_id = ?').all(existingReport.id) as any[];
    }
    
    const calc = calculateReport(service.id, slaDefinition, yearNum, monthNum, existingDetails);
    
    const detailsToUse = calc.details.map(d => {
      if (existingReport && d.isConfirmed) {
        const existing = existingDetails.find(ed => ed.metric_type === d.type);
        if (existing) {
          return {
            ...d,
            actual: existing.actual,
            isAchieved: !!existing.is_achieved
          };
        }
      }
      return d;
    });
    
    let reportId: string;
    
    if (existingReport) {
      reportId = existingReport.id;
      
      db.prepare(`
        UPDATE reports 
        SET availability_achieved = ?, response_time_p99 = ?, error_rate = ?, updated_at = ?
        WHERE id = ?
      `).run(
        detailsToUse.find(d => d.type === 'availability')!.actual,
        detailsToUse.find(d => d.type === 'responseTime')!.actual,
        detailsToUse.find(d => d.type === 'errorRate')!.actual,
        now,
        reportId
      );
      
      for (const detail of detailsToUse) {
        const existing = existingDetails.find(ed => ed.metric_type === detail.type);
        if (existing) {
          db.prepare(`
            UPDATE report_details 
            SET target = ?, actual = ?, is_achieved = ?, is_confirmed = ?
            WHERE id = ?
          `).run(
            detail.target,
            detail.actual,
            detail.isAchieved ? 1 : 0,
            detail.isConfirmed ? 1 : 0,
            existing.id
          );
        } else {
          const detailId = generateUUID();
          db.prepare(`
            INSERT INTO report_details (id, report_id, metric_type, target, actual, is_achieved, is_confirmed)
            VALUES (?, ?, ?, ?, ?, ?, ?)
          `).run(
            detailId,
            reportId,
            detail.type,
            detail.target,
            detail.actual,
            detail.isAchieved ? 1 : 0,
            detail.isConfirmed ? 1 : 0
          );
        }
      }
    } else {
      reportId = generateUUID();
      
      db.prepare(`
        INSERT INTO reports (id, year, month, service_id, sla_definition_id, availability_achieved, response_time_p99, error_rate, is_confirmed, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).run(
        reportId,
        yearNum,
        monthNum,
        service.id,
        slaDefinition.id,
        detailsToUse.find(d => d.type === 'availability')!.actual,
        detailsToUse.find(d => d.type === 'responseTime')!.actual,
        detailsToUse.find(d => d.type === 'errorRate')!.actual,
        0,
        now,
        now
      );
      
      for (const detail of detailsToUse) {
        const detailId = generateUUID();
        db.prepare(`
          INSERT INTO report_details (id, report_id, metric_type, target, actual, is_achieved, is_confirmed)
          VALUES (?, ?, ?, ?, ?, ?, ?)
        `).run(
          detailId,
          reportId,
          detail.type,
          detail.target,
          detail.actual,
          detail.isAchieved ? 1 : 0,
          detail.isConfirmed ? 1 : 0
        );
      }
    }
    
    const breachedDetails = detailsToUse.filter(d => !d.isAchieved);
    for (const breached of breachedDetails) {
      for (const incident of calc.incidents) {
        try {
          const breachId = generateUUID();
          db.prepare(`
            INSERT INTO breaches (id, report_id, incident_id, metric_type, created_at)
            VALUES (?, ?, ?, ?, ?)
          `).run(
            breachId,
            reportId,
            incident.id,
            breached.type,
            now
          );
        } catch (err: any) {
          if (err.code !== 'SQLITE_CONSTRAINT_UNIQUE') {
            throw err;
          }
        }
      }
    }
    
    const breaches = db.prepare('SELECT * FROM breaches WHERE report_id = ?').all(reportId) as any[];
    const allMetrics = {
      availability: {
        target: slaDefinition.availability_target,
        actual: detailsToUse.find(d => d.type === 'availability')!.actual,
        achieved: detailsToUse.find(d => d.type === 'availability')!.isAchieved
      },
      responseTime: {
        target: slaDefinition.response_time_target,
        actual: detailsToUse.find(d => d.type === 'responseTime')!.actual,
        achieved: detailsToUse.find(d => d.type === 'responseTime')!.isAchieved
      },
      errorRate: {
        target: slaDefinition.error_rate_target,
        actual: detailsToUse.find(d => d.type === 'errorRate')!.actual,
        achieved: detailsToUse.find(d => d.type === 'errorRate')!.isAchieved
      }
    };
    
    const finalAchievedCount = Object.values(allMetrics).filter(m => m.achieved).length;
    const finalComplianceRate = (finalAchievedCount / 3) * 100;
    
    updatedReports.push({
      serviceId: service.id,
      serviceName: service.name,
      year: yearNum,
      month: monthNum,
      slaDefinition: {
        availabilityTarget: slaDefinition.availability_target,
        responseTimeTarget: slaDefinition.response_time_target,
        errorRateTarget: slaDefinition.error_rate_target
      },
      achieved: {
        availability: allMetrics.availability.actual,
        responseTimeP99: allMetrics.responseTime.actual,
        errorRate: allMetrics.errorRate.actual
      },
      metrics: allMetrics,
      incidents: calc.incidents.map(i => ({
        id: i.id,
        serviceId: i.service_id,
        startTime: i.start_time,
        endTime: i.end_time,
        durationMinutes: i.duration_minutes,
        description: i.description,
        isAutomatic: !!i.is_automatic,
        createdAt: i.created_at
      })),
      complianceRate: finalComplianceRate,
      breaches: breaches.map(b => ({
        id: b.id,
        reportId: b.report_id,
        incidentId: b.incident_id,
        metricType: b.metric_type,
        createdAt: b.created_at
      })),
      isConfirmed: false
    });
  }
  
  return res.json(updatedReports);
});

export default router;
