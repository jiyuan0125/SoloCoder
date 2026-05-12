import { Router, Response } from 'express';
import db from '../database';
import { AuthRequest, authenticateToken, requireSupervisor } from '../middleware/auth';

const router = Router();
router.use(authenticateToken);
router.use(requireSupervisor);

function getCurrentMonthRange(): { start: string; end: string } {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const daysInMonth = new Date(year, now.getMonth() + 1, 0).getDate();
  
  return {
    start: `${year}-${month}-01`,
    end: `${year}-${month}-${daysInMonth}`,
  };
}

router.get('/', (req: AuthRequest, res: Response) => {
  const range = getCurrentMonthRange();

  const activeCases = db.prepare(`
    SELECT COUNT(*) as count FROM cases 
    WHERE status IN ('open', 'in_progress', 'extended')
  `).get() as { count: number };

  const newCasesThisMonth = db.prepare(`
    SELECT COUNT(*) as count FROM cases 
    WHERE DATE(start_date) BETWEEN ? AND ?
  `).get(range.start, range.end) as { count: number };

  const closedCasesThisMonth = db.prepare(`
    SELECT COUNT(*) as count FROM cases 
    WHERE status IN ('closed', 'archived') 
    AND DATE(actual_end_date) BETWEEN ? AND ?
  `).get(range.start, range.end) as { count: number };

  const avgDuration = db.prepare(`
    SELECT AVG(
      julianday(actual_end_date) - julianday(start_date)
    ) as avg_days
    FROM cases 
    WHERE status IN ('closed', 'archived')
    AND actual_end_date IS NOT NULL
  `).get() as { avg_days: number | null };

  const groupStats = db.prepare(`
    SELECT 
      SUM(completed_count) as total_completed,
      SUM(activity_count) as total_planned
    FROM group_activities
  `).get() as { total_completed: number; total_planned: number };

  const completionRate = groupStats.total_planned > 0 
    ? (groupStats.total_completed / groupStats.total_planned) * 100 
    : 0;

  const communityCoverage = db.prepare(`
    SELECT SUM(participants) as total_coverage
    FROM community_services
    WHERE DATE(service_date) BETWEEN ? AND ?
  `).get(range.start, range.end) as { total_coverage: number | null };

  res.json({
    current_active_cases: activeCases.count,
    new_cases_this_month: newCasesThisMonth.count,
    closed_cases_this_month: closedCasesThisMonth.count,
    average_case_duration_days: avgDuration.avg_days !== null ? Math.round(avgDuration.avg_days) : 0,
    group_activity_completion_rate: Math.round(completionRate * 100) / 100,
    community_service_coverage: communityCoverage.total_coverage || 0,
    period: {
      start: range.start,
      end: range.end,
    },
  });
});

router.get('/cases-breakdown', (req: AuthRequest, res: Response) => {
  const byStatus = db.prepare(`
    SELECT status, COUNT(*) as count
    FROM cases
    GROUP BY status
  `).all();

  const bySocialWorker = db.prepare(`
    SELECT u.id, u.name, COUNT(c.id) as case_count
    FROM users u
    LEFT JOIN cases c ON u.id = c.social_worker_id AND c.status IN ('open', 'in_progress', 'extended')
    WHERE u.role = 'social_worker'
    GROUP BY u.id, u.name
  `).all();

  res.json({
    by_status: byStatus,
    by_social_worker: bySocialWorker,
  });
});

router.get('/community-stats', (req: AuthRequest, res: Response) => {
  const range = getCurrentMonthRange();

  const monthlyServices = db.prepare(`
    SELECT 
      DATE(service_date) as date,
      COUNT(*) as service_count,
      SUM(participants) as participants
    FROM community_services
    WHERE DATE(service_date) BETWEEN ? AND ?
    GROUP BY DATE(service_date)
    ORDER BY date
  `).all(range.start, range.end);

  const totalServices = db.prepare(`
    SELECT 
      COUNT(*) as total_count,
      SUM(participants) as total_participants,
      AVG(participants) as avg_participants
    FROM community_services
    WHERE DATE(service_date) BETWEEN ? AND ?
  `).get(range.start, range.end);

  res.json({
    monthly_breakdown: monthlyServices,
    summary: totalServices,
  });
});

export default router;
