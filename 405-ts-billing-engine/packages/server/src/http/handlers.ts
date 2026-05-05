import type { IncomingMessage, ServerResponse } from 'http';
import { parseRequestBody, sendResponse, sendError, parseQueryParams } from './utils';
import {
  PLANS,
  ErrorCodes,
  BillingError,
  type CreateCustomerRequest,
  type CreateSubscriptionRequest,
  type UpgradeSubscriptionRequest,
  type RecordUsageRequest,
  type PayBillRequest,
  type RequestReviewRequest,
  type RequestRefundRequest,
  type GenerateBillRequest,
} from '@billing/shared';
import {
  createCustomer,
  checkAndUpdateCustomerStatus,
} from '../services';
import {
  createSubscription,
  getActiveSubscription,
  upgradeSubscription,
  requestRefund,
} from '../services';
import {
  recordUsage,
  getCurrentMonthUsage,
  calculateMonthlyUsage,
} from '../services';
import {
  generateBill,
  getBill,
  getCustomerBills,
  getBillItems,
  requestBillReview,
} from '../services';
import {
  payBill,
} from '../services';
import {
  findOverdueBillsByCustomerId,
} from '../store';

export async function handleCustomerRoutes(
  req: IncomingMessage,
  res: ServerResponse,
  segments: string[]
): Promise<void> {
  const method = req.method || 'GET';

  if (method === 'POST' && segments.length === 1) {
    try {
      const body = await parseRequestBody<CreateCustomerRequest>(req);
      if (!body.name || !body.email) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: name, email');
      }
      const customer = createCustomer(body);
      sendResponse(res, {
        id: customer.id,
        name: customer.name,
        email: customer.email,
        status: customer.status,
        trialEndAt: customer.trialEndAt,
        createdAt: customer.createdAt,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'GET' && segments.length === 2) {
    try {
      const customerId = segments[1];
      if (!customerId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少客户ID');
      }
      const customer = checkAndUpdateCustomerStatus(customerId);
      const subscription = getActiveSubscription(customerId);

      sendResponse(res, {
        id: customer.id,
        name: customer.name,
        email: customer.email,
        status: customer.status,
        trialEndAt: customer.trialEndAt,
        subscription: subscription
          ? {
              id: subscription.id,
              planType: subscription.planType,
              billingCycle: subscription.billingCycle,
              startsAt: subscription.startsAt,
              isActive: subscription.isActive,
            }
          : undefined,
        createdAt: customer.createdAt,
        updatedAt: customer.updatedAt,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'GET' && segments.length === 3 && segments[2] === 'status') {
    try {
      const customerId = segments[1];
      if (!customerId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少客户ID');
      }
      const customer = checkAndUpdateCustomerStatus(customerId);
      const overdueBills = findOverdueBillsByCustomerId(customerId);

      sendResponse(res, {
        customerId: customer.id,
        status: customer.status,
        isFrozen: customer.status === 'frozen',
        freezeReason: customer.status === 'frozen' ? '连续欠费未付' : undefined,
        overdueBills: overdueBills.map((bill) => ({
          billId: bill.id,
          amount: bill.totalAmount,
          dueAt: bill.dueAt,
        })),
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  sendError(res, new BillingError(ErrorCodes.BAD_REQUEST, '无效的请求路径'), 404);
}

export async function handleSubscriptionRoutes(
  req: IncomingMessage,
  res: ServerResponse,
  segments: string[]
): Promise<void> {
  const method = req.method || 'GET';

  if (method === 'POST' && segments.length === 1) {
    try {
      const body = await parseRequestBody<CreateSubscriptionRequest>(req);
      if (!body.customerId || !body.planType) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: customerId, planType');
      }
      if (!PLANS[body.planType]) {
        throw new BillingError(ErrorCodes.INVALID_PLAN_TYPE, '无效的套餐类型');
      }
      const result = createSubscription({
        customerId: body.customerId,
        planType: body.planType,
        billingCycle: body.billingCycle || 'monthly',
      });
      sendResponse(res, {
        id: result.subscription.id,
        customerId: result.subscription.customerId,
        planType: result.subscription.planType,
        billingCycle: result.subscription.billingCycle,
        startsAt: result.subscription.startsAt,
        isActive: result.subscription.isActive,
        yearlyDiscountApplied: result.yearlyDiscountApplied,
        createdAt: result.subscription.createdAt,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'POST' && segments.length === 2 && segments[1] === 'upgrade') {
    try {
      const body = await parseRequestBody<UpgradeSubscriptionRequest>(req);
      if (!body.customerId || !body.subscriptionId || !body.toPlanType) {
        throw new BillingError(
          ErrorCodes.BAD_REQUEST,
          '缺少必要参数: customerId, subscriptionId, toPlanType'
        );
      }
      const result = upgradeSubscription(body);
      sendResponse(res, {
        subscriptionId: result.subscription.id,
        customerId: result.subscription.customerId,
        fromPlan: result.proration.fromPlan,
        toPlan: result.proration.toPlan,
        upgradeDate: result.proration.upgradeDate,
        proration: {
          daysOnOldPlan: result.proration.daysOnOldPlan,
          daysOnNewPlan: result.proration.daysOnNewPlan,
          oldPlanProratedFee: result.proration.oldPlanProratedFee,
          newPlanProratedFee: result.proration.newPlanProratedFee,
          totalBaseFee: result.proration.totalBaseFee,
        },
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'POST' && segments.length === 2 && segments[1] === 'refund') {
    try {
      const body = await parseRequestBody<RequestRefundRequest>(req);
      if (!body.customerId || !body.subscriptionId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: customerId, subscriptionId');
      }
      const result = requestRefund(body);
      sendResponse(res, {
        refundId: result.refundId,
        customerId: body.customerId,
        subscriptionId: body.subscriptionId,
        amount: result.amount,
        refundedMonths: result.refundedMonths,
        status: 'pending',
        createdAt: new Date().toISOString(),
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  sendError(res, new BillingError(ErrorCodes.BAD_REQUEST, '无效的请求路径'), 404);
}

export async function handleUsageRoutes(
  req: IncomingMessage,
  res: ServerResponse,
  segments: string[]
): Promise<void> {
  const method = req.method || 'GET';

  if (method === 'POST' && segments.length === 1) {
    try {
      const body = await parseRequestBody<RecordUsageRequest>(req);
      if (!body.customerId || body.units === undefined) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: customerId, units');
      }
      const result = recordUsage(body);
      sendResponse(res, {
        id: result.record.id,
        customerId: result.record.customerId,
        units: result.record.units,
        recordedAt: result.record.recordedAt,
        isLimited: result.isLimited,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'GET' && segments.length === 2) {
    try {
      const customerId = segments[1];
      if (!customerId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少客户ID');
      }
      const query = parseQueryParams(req);
      const periodStart = query.periodStart;
      const periodEnd = query.periodEnd;

      let usage;
      if (periodStart && periodEnd) {
        const subscription = getActiveSubscription(customerId);
        usage = calculateMonthlyUsage(customerId, subscription.id, periodStart, periodEnd);
      } else {
        usage = getCurrentMonthUsage(customerId);
      }

      const subscription = getActiveSubscription(customerId);
      const plan = PLANS[subscription.planType];

      sendResponse(res, {
        customerId: usage.customerId,
        subscriptionId: usage.subscriptionId,
        planType: subscription.planType,
        monthlyQuota: plan.monthlyQuota,
        periodStart: usage.periodStart,
        periodEnd: usage.periodEnd,
        totalUnits: usage.totalUnits,
        quotaUsed: usage.quotaUsed,
        overageUnits: usage.overageUnits,
        overageFee: usage.overageUnits * plan.overageFeePerUnit,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  sendError(res, new BillingError(ErrorCodes.BAD_REQUEST, '无效的请求路径'), 404);
}

export async function handleBillingRoutes(
  req: IncomingMessage,
  res: ServerResponse,
  segments: string[]
): Promise<void> {
  const method = req.method || 'GET';

  if (method === 'POST' && segments.length === 1) {
    try {
      const body = await parseRequestBody<GenerateBillRequest>(req);
      if (!body.customerId || !body.periodStart || !body.periodEnd) {
        throw new BillingError(
          ErrorCodes.BAD_REQUEST,
          '缺少必要参数: customerId, periodStart, periodEnd'
        );
      }
      const bill = generateBill(body);
      const items = getBillItems(bill.id);

      sendResponse(res, {
        id: bill.id,
        customerId: bill.customerId,
        subscriptionId: bill.subscriptionId,
        periodStart: bill.periodStart,
        periodEnd: bill.periodEnd,
        baseFee: bill.baseFee,
        overageFee: bill.overageFee,
        discount: bill.discount,
        totalAmount: bill.totalAmount,
        status: bill.status,
        dueAt: bill.dueAt,
        reviewDeadline: bill.reviewDeadline,
        items: items.map((item) => ({
          type: item.type,
          description: item.description,
          amount: item.amount,
        })),
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'GET' && segments.length === 1) {
    try {
      const query = parseQueryParams(req);
      const customerId = query.customerId;
      const status = query.status as string | undefined;

      if (!customerId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: customerId');
      }

      const bills = getCustomerBills(customerId, status);
      const limit = query.limit ? parseInt(query.limit, 10) : 100;
      const offset = query.offset ? parseInt(query.offset, 10) : 0;

      const paginatedBills = bills.slice(offset, offset + limit);

      sendResponse(res, {
        bills: paginatedBills.map((bill) => ({
          id: bill.id,
          customerId: bill.customerId,
          periodStart: bill.periodStart,
          periodEnd: bill.periodEnd,
          totalAmount: bill.totalAmount,
          status: bill.status,
          dueAt: bill.dueAt,
          createdAt: bill.createdAt,
        })),
        total: bills.length,
        limit,
        offset,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'GET' && segments.length === 2) {
    try {
      const billId = segments[1];
      if (!billId) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少账单ID');
      }
      const bill = getBill(billId);
      const items = getBillItems(billId);

      sendResponse(res, {
        id: bill.id,
        customerId: bill.customerId,
        subscriptionId: bill.subscriptionId,
        periodStart: bill.periodStart,
        periodEnd: bill.periodEnd,
        baseFee: bill.baseFee,
        overageFee: bill.overageFee,
        discount: bill.discount,
        totalAmount: bill.totalAmount,
        status: bill.status,
        dueAt: bill.dueAt,
        paidAt: bill.paidAt,
        reviewRequestedAt: bill.reviewRequestedAt,
        reviewDeadline: bill.reviewDeadline,
        notes: bill.notes,
        items: items.map((item) => ({
          id: item.id,
          type: item.type,
          description: item.description,
          units: item.units,
          unitPrice: item.unitPrice,
          amount: item.amount,
        })),
        createdAt: bill.createdAt,
        updatedAt: bill.updatedAt,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'POST' && segments.length === 2 && segments[1] === 'pay') {
    try {
      const body = await parseRequestBody<PayBillRequest>(req);
      if (!body.billId || body.amount === undefined) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: billId, amount');
      }
      const payment = payBill(body);
      sendResponse(res, {
        paymentId: payment.id,
        billId: payment.billId,
        amount: payment.amount,
        status: payment.status,
        paidAt: payment.paidAt,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  if (method === 'POST' && segments.length === 2 && segments[1] === 'review') {
    try {
      const body = await parseRequestBody<RequestReviewRequest>(req);
      if (!body.billId || !body.reason) {
        throw new BillingError(ErrorCodes.BAD_REQUEST, '缺少必要参数: billId, reason');
      }
      const bill = requestBillReview(body.billId, body.reason);
      sendResponse(res, {
        billId: bill.id,
        status: bill.status,
        reviewRequestedAt: bill.reviewRequestedAt,
        reviewDeadline: bill.reviewDeadline,
      });
    } catch (error) {
      sendError(res, error);
    }
    return;
  }

  sendError(res, new BillingError(ErrorCodes.BAD_REQUEST, '无效的请求路径'), 404);
}
