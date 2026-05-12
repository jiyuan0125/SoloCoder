import { Subscription, PlanId, PlanChangeRecord } from '../types';
import { store } from '../data/store';
import { planService } from './planService';

export class SubscriptionService {
  createSubscription(customerId: string, planId: PlanId): Subscription | { error: string } {
    const plan = planService.getPlanById(planId);
    if (!plan) {
      return { error: '套餐不存在' };
    }

    const existingSub = store.subscriptions.find(
      (s) => s.customerId === customerId && s.planId === planId && s.status !== 'cancelled'
    );
    if (existingSub) {
      return { error: '该客户已订阅此套餐' };
    }

    const subscription: Subscription = {
      id: store.generateId(),
      customerId,
      planId,
      status: 'active',
      startDate: new Date(),
      planChangeHistory: [
        {
          fromPlanId: null,
          toPlanId: planId,
          effectiveDate: new Date(),
          reason: '初始订阅',
        },
      ],
    };

    store.subscriptions.push(subscription);
    return subscription;
  }

  getSubscriptionsByCustomer(customerId: string): Subscription[] {
    return store.subscriptions.filter((s) => s.customerId === customerId);
  }

  getSubscriptionById(id: string): Subscription | undefined {
    return store.subscriptions.find((s) => s.id === id);
  }

  upgradePlan(
    subscriptionId: string,
    newPlanId: PlanId
  ): { subscription: Subscription } | { error: string; status: number } {
    const subscription = store.subscriptions.find((s) => s.id === subscriptionId);
    if (!subscription) {
      return { error: '订阅不存在', status: 404 };
    }

    if (subscription.status !== 'active') {
      return { error: '仅活跃的订阅可以变更套餐', status: 400 };
    }

    if (subscription.planId === newPlanId) {
      return { error: '重复的套餐变更请求', status: 409 };
    }

    const lastChange = subscription.planChangeHistory[subscription.planChangeHistory.length - 1];
    if (lastChange.toPlanId === newPlanId) {
      return { error: '重复的套餐变更请求', status: 409 };
    }

    const newPlan = planService.getPlanById(newPlanId);
    if (!newPlan) {
      return { error: '套餐不存在', status: 404 };
    }

    const changeRecord: PlanChangeRecord = {
      fromPlanId: subscription.planId,
      toPlanId: newPlanId,
      effectiveDate: new Date(),
      reason: '套餐升级/变更',
    };

    subscription.planId = newPlanId;
    subscription.planChangeHistory.push(changeRecord);

    return { subscription };
  }

  cancelSubscription(subscriptionId: string): { success: boolean; message?: string } {
    const subscription = store.subscriptions.find((s) => s.id === subscriptionId);
    if (!subscription) {
      return { success: false, message: '订阅不存在' };
    }

    subscription.status = 'cancelled';
    return { success: true };
  }

  pauseSubscription(subscriptionId: string): { success: boolean; message?: string } {
    const subscription = store.subscriptions.find((s) => s.id === subscriptionId);
    if (!subscription) {
      return { success: false, message: '订阅不存在' };
    }

    if (subscription.status === 'cancelled') {
      return { success: false, message: '已取消的订阅无法暂停' };
    }

    subscription.status = 'paused';
    return { success: true };
  }

  resumeSubscription(subscriptionId: string): { success: boolean; message?: string } {
    const subscription = store.subscriptions.find((s) => s.id === subscriptionId);
    if (!subscription) {
      return { success: false, message: '订阅不存在' };
    }

    if (subscription.status === 'cancelled') {
      return { success: false, message: '已取消的订阅无法恢复' };
    }

    subscription.status = 'active';
    return { success: true };
  }

  getActiveSubscriptions(): Subscription[] {
    return store.subscriptions.filter((s) => s.status === 'active');
  }
}

export const subscriptionService = new SubscriptionService();
