import {
  Salesperson,
  Order,
  Ranking,
  Year,
  Month,
  Quarter,
  RankingFilter,
  AmountInCents,
  SalesContribution,
} from '@commission-tracker/shared';
import { store } from '../store';
import {
  calculateCommissionForSales,
  applyProbationDiscount,
} from './commissionCalculator';
import { allocateTeamOrderAmount } from './teamAllocation';
import { isInProbationPeriod } from './settlementService';

function isOrderInTimeRange(
  order: Order,
  year: Year,
  month?: Month,
  quarter?: Quarter
): boolean {
  const orderDate = new Date(order.date);
  const orderYear = orderDate.getFullYear();
  const orderMonth = orderDate.getMonth() + 1;

  if (orderYear !== year) return false;

  if (month !== undefined) {
    return orderMonth === month;
  }

  if (quarter !== undefined) {
    const orderQuarter = Math.ceil(orderMonth / 3) as Quarter;
    return orderQuarter === quarter;
  }

  return true;
}

function calculateSalespersonStatsForPeriod(
  salesperson: Salesperson,
  orders: Order[],
  year: Year,
  month?: Month,
  quarter?: Quarter
): { totalSales: AmountInCents; totalCommission: AmountInCents } {
  const relevantOrders = orders.filter(o =>
    isOrderInTimeRange(o, year, month, quarter) &&
    o.status === 'active' &&
    o.salesContributions.some((c: SalesContribution) => c.salespersonId === salesperson.id)
  );

  let totalSales: AmountInCents = 0;
  let totalCommission: AmountInCents = 0;

  for (const order of relevantOrders) {
    const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
    const salespersonAllocation = allocations.find(
      a => a.salespersonId === salesperson.id
    );

    if (!salespersonAllocation) continue;

    totalSales += salespersonAllocation.allocatedAmount;

    const { commission } = calculateCommissionForSales(salespersonAllocation.allocatedAmount);
    const inProbation = isInProbationPeriod(salesperson, order.date);
    const finalCommission = inProbation
      ? applyProbationDiscount(commission)
      : commission;

    totalCommission += finalCommission;
  }

  return { totalSales, totalCommission };
}

export function getRankings(filter: RankingFilter): Ranking[] {
  const { dimension, year, month, quarter } = filter;

  const salespeople = store.getSalespeople();
  const orders = store.getOrders();

  const rankings: Ranking[] = [];

  for (const salesperson of salespeople) {
    let targetMonth: Month | undefined;
    let targetQuarter: Quarter | undefined;

    if (dimension === 'monthly') {
      if (month === undefined) {
        const now = new Date();
        targetMonth = (now.getMonth() + 1) as Month;
      } else {
        targetMonth = month;
      }
    } else if (dimension === 'quarterly') {
      if (quarter === undefined) {
        const now = new Date();
        targetQuarter = Math.ceil((now.getMonth() + 1) / 3) as Quarter;
      } else {
        targetQuarter = quarter;
      }
    }

    const stats = calculateSalespersonStatsForPeriod(
      salesperson,
      orders,
      year,
      targetMonth,
      targetQuarter
    );

    if (stats.totalSales > 0) {
      rankings.push({
        salespersonId: salesperson.id,
        salespersonName: salesperson.name,
        rank: 0,
        totalSales: stats.totalSales,
        totalCommission: stats.totalCommission,
      });
    }
  }

  rankings.sort((a, b) => b.totalSales - a.totalSales);

  let currentRank = 1;
  let previousSales: AmountInCents | null = null;

  for (let i = 0; i < rankings.length; i++) {
    if (previousSales !== null && rankings[i].totalSales < previousSales) {
      currentRank = i + 1;
    }
    rankings[i].rank = currentRank;
    previousSales = rankings[i].totalSales;
  }

  return rankings;
}
