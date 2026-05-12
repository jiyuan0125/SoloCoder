import { db } from '../db';
import { CancelReason } from '../types';

export class ReportService {
  getObservationReport(month: string, equipmentId?: number) {
    let query = `
      SELECT 
        ot.*,
        e.name as equipment_name
      FROM observation_tasks ot
      LEFT JOIN equipment e ON ot.equipment_id = e.id
      WHERE ot.scheduled_month = ?
    `;
    const params: any[] = [month];

    if (equipmentId) {
      query += ' AND ot.equipment_id = ?';
      params.push(equipmentId);
    }

    query += ' ORDER BY ot.start_time';

    const tasks = db.prepare(query).all(...params) as any[];

    const result = tasks.map(task => {
      const isCompleted = task.status === '已完成' || task.status === '已归档';
      const isWeatherCanceled = task.is_canceled && task.cancel_reason === CancelReason.WEATHER;
      const isFaultCanceled = task.is_canceled && task.cancel_reason === CancelReason.EQUIPMENT_FAULT;
      
      return {
        ...task,
        counts_towards_completion: isCompleted || (task.is_canceled && !isWeatherCanceled),
      };
    });

    return { data: result };
  }

  getActivityReport(month: string) {
    const yearMonth = month.substring(0, 7);
    
    const events = db.prepare(`
      SELECT 
        pe.*,
        COUNT(b.id) as booking_count,
        SUM(CASE WHEN b.is_waitlist = 0 THEN b.seats ELSE 0 END) as total_confirmed_seats
      FROM public_events pe
      LEFT JOIN bookings b ON pe.id = b.event_id
      WHERE strftime('%Y-%m', pe.start_time) = ?
      GROUP BY pe.id
      ORDER BY pe.start_time
    `).all(yearMonth) as any[];

    return { data: events };
  }

  getMonthlyStats(month: string) {
    const obsReport = this.getObservationReport(month);
    const tasks = obsReport.data as any[];

    const totalTasks = tasks.length;
    const completedTasks = tasks.filter(t => t.status === '已完成' || t.status === '已归档').length;
    const weatherCanceled = tasks.filter(t => t.cancel_reason === CancelReason.WEATHER).length;
    const faultCanceled = tasks.filter(t => t.cancel_reason === CancelReason.EQUIPMENT_FAULT).length;
    
    const eligibleForRate = tasks.filter(t => !t.cancel_reason || t.cancel_reason !== CancelReason.WEATHER).length;
    const completionRate = eligibleForRate > 0 
      ? Math.round((completedTasks / eligibleForRate) * 100) 
      : 0;

    const actReport = this.getActivityReport(month);
    const events = actReport.data as any[];

    return {
      data: {
        month,
        observation_stats: {
          total_tasks: totalTasks,
          completed_tasks: completedTasks,
          weather_canceled: weatherCanceled,
          fault_canceled: faultCanceled,
          completion_rate: `${completionRate}%`,
        },
        public_activity_stats: {
          total_events: events.length,
          canceled_events: events.filter(e => e.is_canceled).length,
          total_participants: events.reduce((sum, e) => sum + (e.total_confirmed_seats || 0), 0),
        },
      }
    };
  }
}

export const reportService = new ReportService();
