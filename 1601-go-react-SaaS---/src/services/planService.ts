import { Plan, PlanId } from '../types';
import { store } from '../data/store';

export class PlanService {
  getAllPlans(): Plan[] {
    return store.plans;
  }

  getPlanById(id: PlanId): Plan | undefined {
    return store.plans.find((p) => p.id === id);
  }

  createPlan(plan: Plan): Plan {
    store.plans.push(plan);
    return plan;
  }

  deletePlan(id: PlanId): { success: boolean; message?: string } {
    const hasSubscriptions = store.subscriptions.some((s) => s.planId === id);
    if (hasSubscriptions) {
      return { success: false, message: '该套餐仍有客户订阅，无法删除' };
    }

    const index = store.plans.findIndex((p) => p.id === id);
    if (index === -1) {
      return { success: false, message: '套餐不存在' };
    }

    store.plans.splice(index, 1);
    return { success: true };
  }
}

export const planService = new PlanService();
