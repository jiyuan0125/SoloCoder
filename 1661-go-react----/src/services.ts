import { z } from 'zod';
import { dbMethods } from './db';
import { AdPlan, AdStatus, Targeting, BidType } from './types';

export const createPlanSchema = z.object({
  name: z.string().min(1),
  budget: z.number().int().positive(),
  dailyBudget: z.number().int().positive(),
  targeting: z.object({
    ageRanges: z.array(z.string()).optional(),
    gender: z.enum(['male', 'female', 'all']).optional(),
    regions: z.array(z.string()).optional(),
  }),
  bidType: z.enum(['CPC', 'CPM']),
  bidAmount: z.number().int(),
});

export const statusTransitions: Record<AdStatus, AdStatus> = {
  draft: 'running',
  running: 'paused',
  paused: 'ended',
  ended: 'ended',
};

export function canTransition(from: AdStatus, to: AdStatus): boolean {
  return statusTransitions[from] === to;
}

function generateId(): string {
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

export function createPlan(data: z.infer<typeof createPlanSchema>) {
  if (data.bidAmount <= 0) {
    throw { status: 400, message: '出价金额必须为正整数' };
  }
  const plan = dbMethods.insertPlan({
    id: generateId(),
    name: data.name,
    budget: data.budget,
    dailyBudget: data.dailyBudget,
    targeting: data.targeting,
    bidType: data.bidType,
    bidAmount: data.bidAmount,
    status: 'draft',
    expectedCtr: 0.05,
  });
  return plan;
}

export function getPlan(id: string) {
  return dbMethods.getPlan(id);
}

export function getPlanWithStats(id: string) {
  const plan = dbMethods.getPlan(id);
  if (!plan) return null;
  const stats = dbMethods.getPlanStats(id);
  const ctr = stats.impressions === 0 ? 0 : stats.clicks / stats.impressions;
  return {
    ...plan,
    stats: {
      ...stats,
      ctr,
    },
  };
}

export function updatePlanStatus(id: string, newStatus: AdStatus) {
  const plan = dbMethods.getPlan(id);
  if (!plan) return null;
  if (!canTransition(plan.status, newStatus)) {
    throw { 
      status: 400, 
      message: `非法状态流转：当前状态为 ${plan.status}，无法转换为 ${newStatus}` 
    };
  }
  return dbMethods.updatePlanStatus(id, newStatus);
}

export function checkBudgetAndPause(planId: string): { paused: boolean; notified: boolean } {
  const plan = dbMethods.getPlan(planId);
  if (!plan || plan.status !== 'running') {
    return { paused: false, notified: false };
  }
  const totalSpend = dbMethods.getTotalSpend(planId);
  const remaining = plan.budget - totalSpend;
  if (remaining <= 0) {
    dbMethods.updatePlanStatus(planId, 'paused');
    let notified = true;
    try {
      sendBudgetExhaustedNotification(plan.id, plan.name);
    } catch (e) {
      notified = false;
    }
    return { paused: true, notified };
  }
  return { paused: false, notified: false };
}

function sendBudgetExhaustedNotification(planId: string, planName: string) {
  console.log(`[Notification] 计划 ${planName} (${planId}) 预算已耗尽`);
}

export interface Opportunity {
  age?: string;
  gender?: 'male' | 'female' | 'all';
  region?: string;
}

function matchesTargeting(plan: AdPlan, opp: Opportunity): boolean {
  const t = plan.targeting;
  if (t.ageRanges && opp.age && !t.ageRanges.includes(opp.age)) return false;
  if (t.gender && t.gender !== 'all' && opp.gender && t.gender !== opp.gender) return false;
  if (t.regions && opp.region && !t.regions.includes(opp.region)) return false;
  return true;
}

export function selectBestPlan(opp: Opportunity): AdPlan | null {
  const running = dbMethods.getRunningPlans();
  const eligible = running.filter(plan => {
    if (!matchesTargeting(plan, opp)) return false;
    const totalSpend = dbMethods.getTotalSpend(plan.id);
    if (totalSpend >= plan.budget) return false;
    const todaySpend = dbMethods.getTodaySpend(plan.id);
    if (todaySpend >= plan.dailyBudget) return false;
    return true;
  });
  if (eligible.length === 0) return null;
  
  const scored = eligible.map(plan => {
    let valuePerImpression: number;
    if (plan.bidType === 'CPC') {
      valuePerImpression = plan.bidAmount * plan.expectedCtr;
    } else {
      valuePerImpression = plan.bidAmount / 1000;
    }
    const totalSpend = dbMethods.getTotalSpend(plan.id);
    const remainingRatio = plan.budget > 0 ? (plan.budget - totalSpend) / plan.budget : 0;
    return { plan, valuePerImpression, remainingRatio };
  });
  
  scored.sort((a, b) => {
    if (b.valuePerImpression !== a.valuePerImpression) {
      return b.valuePerImpression - a.valuePerImpression;
    }
    return b.remainingRatio - a.remainingRatio;
  });
  
  return scored[0].plan;
}

export function recordImpression(planId: string): { success: boolean; cost: number; budgetExhausted: boolean; notifyFailed?: boolean } {
  const plan = dbMethods.getPlan(planId);
  if (!plan || plan.status !== 'running') {
    return { success: false, cost: 0, budgetExhausted: false };
  }
  
  const result = dbMethods.transaction(() => {
    const overcharge = dbMethods.getOvercharge(planId);
    let cost: number;
    if (plan.bidType === 'CPC') {
      cost = plan.bidAmount;
    } else {
      cost = Math.round(plan.bidAmount / 1000);
    }
    
    const totalSpend = dbMethods.getTotalSpend(planId);
    const available = plan.budget - totalSpend;
    let actualCost = cost;
    let overchargedAmount = 0;
    
    if (available <= 0) {
      return { success: false, cost: 0, budgetExhausted: true };
    }
    
    if (overcharge > 0) {
      if (cost > overcharge) {
        overchargedAmount = cost - overcharge;
        dbMethods.setOvercharge(planId, 0);
        actualCost = overchargedAmount;
      } else {
        dbMethods.setOvercharge(planId, overcharge - cost);
        actualCost = 0;
      }
    }
    
    if (actualCost > 0) {
      dbMethods.insertLedger({
        planId,
        amount: actualCost,
        type: 'deduction',
      });
    }
    
    const hour = new Date();
    hour.setMinutes(0, 0, 0);
    dbMethods.updateHourlyStats(planId, hour.getTime(), {
      impressions: 1,
      spend: actualCost,
    });
    
    const newTotalSpend = dbMethods.getTotalSpend(planId);
    if (newTotalSpend > plan.budget) {
      const exceed = newTotalSpend - plan.budget;
      dbMethods.setOvercharge(planId, overcharge + exceed);
    }
    
    return { success: true, cost: actualCost, budgetExhausted: false };
  });
  
  if (result.budgetExhausted) {
    const checkResult = checkBudgetAndPause(planId);
    return { ...result, notifyFailed: checkResult.paused && !checkResult.notified };
  }
  
  const checkResult = checkBudgetAndPause(planId);
  if (checkResult.paused) {
    return { ...result, budgetExhausted: true, notifyFailed: !checkResult.notified };
  }
  
  return result;
}

export function recordClick(planId: string) {
  const hour = new Date();
  hour.setMinutes(0, 0, 0);
  dbMethods.updateHourlyStats(planId, hour.getTime(), { clicks: 1 });
  
  const stats = dbMethods.getPlanStats(planId);
  const newCtr = stats.impressions === 0 ? 0.05 : stats.clicks / stats.impressions;
  dbMethods.updatePlanExpectedCtr(planId, newCtr);
}

export function recordConversion(planId: string) {
  const hour = new Date();
  hour.setMinutes(0, 0, 0);
  dbMethods.updateHourlyStats(planId, hour.getTime(), { conversions: 1 });
}
