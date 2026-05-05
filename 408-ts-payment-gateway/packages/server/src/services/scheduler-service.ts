import { Order, OrderStatus, CALLBACK_TIMEOUT_SECONDS, SYNC_INTERVAL_HOURS } from "@payment-gateway/shared";
import { getProcessingOrders, getPendingConfirmOrders, saveOrder } from "./order-service";
import { setOrderPendingConfirm, confirmPaymentSuccess, confirmPaymentFailed } from "./payment-service";
import { getPendingRefunds, confirmRefundSuccess, confirmRefundFailed } from "./refund-service";
import { simulateChannelQueryOrderStatus, simulateChannelQueryRefundStatus } from "../channel-mock";
import { CALLBACK_CHECK_INTERVAL_MS, DAILY_SYNC_HOUR } from "../config";

let callbackCheckTimer: NodeJS.Timeout | null = null;
let dailySyncTimer: NodeJS.Timeout | null = null;
let lastSyncDay: number | null = null;

export function startScheduler(): void {
  callbackCheckTimer = setInterval(checkTimeoutPayments, CALLBACK_CHECK_INTERVAL_MS);
  setupDailySync();

  console.log("Scheduler started");
}

export function stopScheduler(): void {
  if (callbackCheckTimer) {
    clearInterval(callbackCheckTimer);
    callbackCheckTimer = null;
  }
  if (dailySyncTimer) {
    clearInterval(dailySyncTimer);
    dailySyncTimer = null;
  }

  console.log("Scheduler stopped");
}

function checkTimeoutPayments(): void {
  const now = Date.now();
  const timeoutThreshold = CALLBACK_TIMEOUT_SECONDS * 1000;

  const processingOrders = getProcessingOrders();
  for (const order of processingOrders) {
    const payments = getOrderPayments(order.orderNo);
    for (const payment of payments) {
      if (payment.status === OrderStatus.PROCESSING) {
        const elapsed = now - payment.createdAt;
        if (elapsed > timeoutThreshold) {
          console.log(`Payment timeout for order ${order.orderNo}, setting to pending_confirm`);
          setOrderPendingConfirm(order);
          break;
        }
      }
    }
  }

  checkDailySync();
}

function setupDailySync(): void {
  dailySyncTimer = setInterval(checkDailySync, 60000);
}

function checkDailySync(): void {
  const now = new Date();
  const currentDay = now.getDate();
  const currentHour = now.getHours();

  if (currentHour >= DAILY_SYNC_HOUR && lastSyncDay !== currentDay) {
    syncPendingConfirmOrders();
    syncPendingRefunds();
    lastSyncDay = currentDay;
    console.log(`Daily sync executed at ${now.toISOString()}`);
  }
}

function syncPendingConfirmOrders(): void {
  const pendingConfirmOrders = getPendingConfirmOrders();
  console.log(`Syncing ${pendingConfirmOrders.length} pending_confirm orders`);

  for (const order of pendingConfirmOrders) {
    try {
      const result = simulateChannelQueryOrderStatus(order.channel, order.orderNo);

      if (result.status === "success" && result.transactionId && result.paidAt) {
        confirmPaymentSuccess(order, result.transactionId, result.paidAt);
        order.lastSyncAt = Date.now();
        saveOrder(order);
        console.log(`Order ${order.orderNo} synced successfully`);
      } else if (result.status === "failed") {
        confirmPaymentFailed(order);
        order.lastSyncAt = Date.now();
        saveOrder(order);
        console.log(`Order ${order.orderNo} synced as failed`);
      } else {
        order.lastSyncAt = Date.now();
        saveOrder(order);
        console.log(`Order ${order.orderNo} still pending`);
      }
    } catch (error) {
      console.error(`Error syncing order ${order.orderNo}:`, error);
    }
  }
}

function syncPendingRefunds(): void {
  const pendingRefunds = getPendingRefunds();
  console.log(`Syncing ${pendingRefunds.length} pending refunds`);

  for (const refund of pendingRefunds) {
    try {
      const result = simulateChannelQueryRefundStatus(refund.channel, refund.refundNo);

      if (result.status === "success" && result.channelRefundId && result.completedAt) {
        const order = getOrderOrNull(refund.orderNo);
        if (order) {
          confirmRefundSuccess(order, refund, result.channelRefundId, result.completedAt);
          console.log(`Refund ${refund.refundNo} synced successfully`);
        }
      } else if (result.status === "failed") {
        confirmRefundFailed(refund);
        console.log(`Refund ${refund.refundNo} synced as failed`);
      }
    } catch (error) {
      console.error(`Error syncing refund ${refund.refundNo}:`, error);
    }
  }
}

function getOrderPayments(orderNo: string): Array<{ orderNo: string; status: OrderStatus; createdAt: number }> {
  const order = getOrderOrNull(orderNo);
  if (!order) return [];

  return [
    {
      orderNo: orderNo,
      status: order.status,
      createdAt: order.createdAt,
    },
  ];
}

function getOrderOrNull(orderNo: string): Order | null {
  const { getOrder } = require("./order-service");
  return getOrder(orderNo);
}
