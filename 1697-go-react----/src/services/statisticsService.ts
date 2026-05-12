import db from '../database';
import { StatisticsReport } from '../types';
import { toYuan, parseNumeric } from '../utils';

function getDateRange(type: 'daily' | 'weekly' | 'monthly'): { start: Date; end: Date; label: string } {
  const now = new Date();
  let start: Date;
  let end: Date;
  let label: string;

  switch (type) {
    case 'daily':
      start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
      label = start.toISOString().split('T')[0];
      break;
    case 'weekly':
      const dayOfWeek = now.getDay();
      const diffToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
      start = new Date(now);
      start.setDate(now.getDate() - diffToMonday);
      start.setHours(0, 0, 0, 0);
      end = new Date(start);
      end.setDate(start.getDate() + 6);
      end.setHours(23, 59, 59, 999);
      label = `${start.toISOString().split('T')[0]} ~ ${end.toISOString().split('T')[0]}`;
      break;
    case 'monthly':
      start = new Date(now.getFullYear(), now.getMonth(), 1);
      end = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59, 999);
      label = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
      break;
    default:
      start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
      label = start.toISOString().split('T')[0];
  }

  return { start, end, label };
}

function sumNumeric(values: any[]): number {
  return values.reduce((sum, val) => {
    const num = parseNumeric(val);
    if (typeof num === 'number') return sum + num;
    return sum;
  }, 0);
}

function sumEconomicLoss(values: any[]): number {
  const totalFen = values.reduce((sum, val) => {
    const num = parseNumeric(val);
    if (typeof num === 'number') return sum + Math.round(num * 100);
    return sum;
  }, 0);
  return toYuan(totalFen);
}

export function generateStatistics(type: 'daily' | 'weekly' | 'monthly'): StatisticsReport {
  const { start, end, label } = getDateRange(type);

  const reports = db.prepare(`
    SELECT * FROM disaster_reports
    WHERE occurrence_time >= ? AND occurrence_time <= ?
  `).all(start.toISOString(), end.toISOString()) as any[];

  const verifiedReports = reports.filter(r => r.status === '已核查' || r.status === '已发布');

  const breakdownMap = new Map<string, {
    count: number;
    affectedPopulation: number[];
    economicLoss: any[];
  }>();

  reports.forEach(report => {
    const type = report.disaster_type;
    if (!breakdownMap.has(type)) {
      breakdownMap.set(type, {
        count: 0,
        affectedPopulation: [],
        economicLoss: []
      });
    }
    const data = breakdownMap.get(type)!;
    data.count++;
    data.affectedPopulation.push(report.affected_population);
    data.economicLoss.push(report.direct_economic_loss);
  });

  const breakdown = Array.from(breakdownMap.entries()).map(([disasterType, data]) => ({
    disasterType,
    count: data.count,
    affectedPopulation: sumNumeric(data.affectedPopulation),
    economicLoss: sumEconomicLoss(data.economicLoss)
  }));

  return {
    type,
    period: label,
    totalReports: reports.length,
    verifiedReports: verifiedReports.length,
    totalAffectedPopulation: sumNumeric(reports.map(r => r.affected_population)),
    totalEvacuated: sumNumeric(reports.map(r => r.evacuated_population)),
    totalDeaths: sumNumeric(reports.map(r => r.death_missing_count)),
    totalEconomicLoss: sumEconomicLoss(reports.map(r => r.direct_economic_loss)),
    breakdown
  };
}
