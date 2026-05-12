import { db, POVERTY_CAUSES } from '../database';
import type { DashboardStats, PovertyCause, Household } from '../types';
import { getAllHouseholds } from './householdService';

function aggregateStats(
  households: Household[],
  level: 'county' | 'township' | 'village',
  name: string
): DashboardStats {
  const totalHouseholds = households.length;
  const outOfPovertyCount = households.filter(h => h.status === 'out_of_poverty').length;
  const returnMonitoringCount = households.filter(h => h.status === 'return_monitoring').length;

  const povertyCauseDistribution: Record<string, number> = {};
  for (const cause of POVERTY_CAUSES) {
    povertyCauseDistribution[cause] = households.filter(h => h.mainPovertyCause === cause).length;
  }

  return {
    level,
    name,
    totalHouseholds,
    outOfPovertyCount,
    returnMonitoringCount,
    povertyCauseDistribution: povertyCauseDistribution as Record<PovertyCause, number>
  };
}

export function getCountyDashboard(county: string): DashboardStats {
  const households = getAllHouseholds().filter(h => h.county === county);
  return aggregateStats(households, 'county', county);
}

export function getTownshipDashboard(county: string, township: string): DashboardStats {
  const households = getAllHouseholds().filter(h => 
    h.county === county && h.township === township
  );
  return aggregateStats(households, 'township', township);
}

export function getVillageDashboard(
  county: string,
  township: string,
  village: string
): DashboardStats {
  const households = getAllHouseholds().filter(h => 
    h.county === county && h.township === township && h.village === village
  );
  return aggregateStats(households, 'village', village);
}

export function getSummary(): {
  totalCount: number;
  outOfPovertyCount: number;
  returnMonitoringCount: number;
  inPovertyCount: number;
  topCauses: { cause: PovertyCause; count: number }[];
} {
  const all = getAllHouseholds();
  
  const totalCount = all.length;
  const outOfPovertyCount = all.filter(h => h.status === 'out_of_poverty').length;
  const returnMonitoringCount = all.filter(h => h.status === 'return_monitoring').length;
  const inPovertyCount = all.filter(h => h.status === 'poverty').length;

  const causeCounts: Record<string, number> = {};
  for (const h of all) {
    causeCounts[h.mainPovertyCause] = (causeCounts[h.mainPovertyCause] || 0) + 1;
  }

  const topCauses = Object.entries(causeCounts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([cause, count]) => ({ cause: cause as PovertyCause, count }));

  return {
    totalCount,
    outOfPovertyCount,
    returnMonitoringCount,
    inPovertyCount,
    topCauses
  };
}
