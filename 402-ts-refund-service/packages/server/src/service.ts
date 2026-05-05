import {
  Order,
  Refund,
  RefundStatus,
  RefundProduct,
  ProductType,
  createServiceError,
  ErrorCodes,
  ServiceError
} from '@refund/shared';
import { generateId, isRefundPeriodExpired, calculateVirtualRefundAmount } from './utils.js';
import { getConfig, updateConfig, defaultConfig } from './config.js';
import { loadData, saveData } from './storage.js';

let orders: Order[] = [];
let refunds: Refund[] = [];

function initializeData(): void {
  const data = loadData();
  orders = data.orders;
  refunds = data.refunds;
  if (data.config.refundPeriodDays !== defaultConfig.refundPeriodDays ||
      data.config.virtualProductRefundThreshold !== defaultConfig.virtualProductRefundThreshold) {
    updateConfig({
      refundPeriodDays: data.config.refundPeriodDays,
      virtualProductRefundThreshold: data.config.virtualProductRefundThreshold
    });
  }
}

initializeData();

function persistData(): void {
  const config = getConfig();
  saveData({
    orders,
    refunds,
    config
  });
}

function addStatusHistory(refund: Refund, status: RefundStatus, note?: string): void {
  refund.statusHistory.push({
    status,
    time: new Date(),
    note
  });
  refund.updatedAt = new Date();
}

export function addMockOrder(order: Omit<Order, 'orderTime'> & { orderTime?: string }): Order {
  const newOrder: Order = {
    ...order,
    orderTime: order.orderTime ? new Date(order.orderTime) : new Date()
  };
  orders.push(newOrder);
  persistData();
  return newOrder;
}

export function getAllOrders(): Order[] {
  return [...orders];
}

export function getOrderById(orderId: string): Order | undefined {
  return orders.find(o => o.orderId === orderId);
}

function findPendingRefundByOrderId(orderId: string): Refund | undefined {
  return refunds.find(r => 
    r.orderId === orderId && 
    r.status !== 'refunded' && 
    r.status !== 'rejected'
  );
}

export function createRefund(
  orderId: string,
  userId: string,
  refundReason: string,
  productIds: string[]
): Refund | ServiceError {
  try {
    const config = getConfig();
    const order = getOrderById(orderId);
    
    if (!order) {
      return createServiceError(ErrorCodes.ORDER_NOT_FOUND, `Order ${orderId} not found`);
    }

    if (order.userId !== userId) {
      return createServiceError(ErrorCodes.INVALID_REFUND_REQUEST, 'User does not own this order');
    }

    if (isRefundPeriodExpired(order.orderTime, config.refundPeriodDays)) {
      return createServiceError(ErrorCodes.REFUND_PERIOD_EXPIRED, `Refund period expired (${config.refundPeriodDays} days)`);
    }

    const existingRefund = findPendingRefundByOrderId(orderId);
    if (existingRefund) {
      return createServiceError(ErrorCodes.REFUND_ALREADY_EXISTS, `A refund is already in progress for this order: ${existingRefund.refundId}`);
    }

    if (productIds.length === 0) {
      return createServiceError(ErrorCodes.INVALID_REFUND_REQUEST, 'No products specified for refund');
    }

    const uniqueProductIds = [...new Set(productIds)];
    const productsToRefund: RefundProduct[] = [];
    let totalRefundAmount = 0;
    let hasPhysicalProducts = false;
    const isFullRefund = uniqueProductIds.length === order.products.length;

    for (const productId of uniqueProductIds) {
      const orderProduct = order.products.find(p => p.productId === productId);
      if (!orderProduct) {
        return createServiceError(ErrorCodes.PRODUCT_NOT_FOUND, `Product ${productId} not found in order`);
      }

      let refundAmount = orderProduct.paymentAmount;

      if (orderProduct.productType === 'virtual') {
        const progress = orderProduct.consumptionProgress ?? 0;
        if (progress >= config.virtualProductRefundThreshold) {
          refundAmount = calculateVirtualRefundAmount(
            orderProduct.paymentAmount,
            progress,
            config.virtualProductRefundThreshold
          );
        }
      } else {
        hasPhysicalProducts = true;
      }

      if (refundAmount <= 0) {
        return createServiceError(ErrorCodes.PRODUCT_ALREADY_CONSUMED, `Product ${orderProduct.productName} has no refundable amount`);
      }

      productsToRefund.push({
        productId: orderProduct.productId,
        productName: orderProduct.productName,
        productType: orderProduct.productType,
        unitPrice: orderProduct.unitPrice,
        quantity: orderProduct.quantity,
        paymentAmount: orderProduct.paymentAmount,
        refundAmount,
        consumptionProgress: orderProduct.consumptionProgress
      });

      totalRefundAmount += refundAmount;
    }

    let shippingFeeRefund = 0;
    if (isFullRefund) {
      shippingFeeRefund = order.shippingFee;
      totalRefundAmount += shippingFeeRefund;
    }

    const now = new Date();
    const refundId = generateId();
    const initialStatus: RefundStatus = hasPhysicalProducts ? 'pending' : 'processing';

    const refund: Refund = {
      refundId,
      orderId,
      userId,
      refundReason,
      refundAmount: totalRefundAmount,
      status: initialStatus,
      products: productsToRefund,
      isFullRefund,
      shippingFeeRefund,
      createdAt: now,
      updatedAt: now,
      statusHistory: [{ status: initialStatus, time: now, note: 'Refund request created' }]
    };

    refunds.push(refund);
    persistData();

    return refund;
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    return createServiceError(ErrorCodes.INTERNAL_ERROR, message);
  }
}

export function submitLogistics(refundId: string, logisticsNumber: string): Refund | ServiceError {
  try {
    const refund = refunds.find(r => r.refundId === refundId);
    if (!refund) {
      return createServiceError(ErrorCodes.REFUND_NOT_FOUND, `Refund ${refundId} not found`);
    }

    if (refund.status !== 'pending') {
      return createServiceError(
        ErrorCodes.INVALID_STATUS_TRANSITION,
        `Cannot submit logistics in current status: ${refund.status}`
      );
    }

    refund.logisticsNumber = logisticsNumber;
    refund.status = 'waiting_for_warehouse_confirm';
    addStatusHistory(refund, 'waiting_for_warehouse_confirm', `Logistics number submitted: ${logisticsNumber}`);
    
    persistData();
    return refund;
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    return createServiceError(ErrorCodes.INTERNAL_ERROR, message);
  }
}

export function warehouseConfirm(refundId: string, received: boolean): Refund | ServiceError {
  try {
    const refund = refunds.find(r => r.refundId === refundId);
    if (!refund) {
      return createServiceError(ErrorCodes.REFUND_NOT_FOUND, `Refund ${refundId} not found`);
    }

    if (refund.status !== 'waiting_for_warehouse_confirm') {
      return createServiceError(
        ErrorCodes.INVALID_STATUS_TRANSITION,
        `Cannot confirm in current status: ${refund.status}`
      );
    }

    if (received) {
      refund.status = 'processing';
      addStatusHistory(refund, 'processing', 'Warehouse confirmed receipt, processing refund');
      
      refund.status = 'refunded';
      addStatusHistory(refund, 'refunded', 'Refund completed');
    } else {
      refund.status = 'rejected';
      addStatusHistory(refund, 'rejected', 'Warehouse rejected: item not received or damaged');
    }

    persistData();
    return refund;
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    return createServiceError(ErrorCodes.INTERNAL_ERROR, message);
  }
}

export function processVirtualRefund(refundId: string): Refund | ServiceError {
  try {
    const refund = refunds.find(r => r.refundId === refundId);
    if (!refund) {
      return createServiceError(ErrorCodes.REFUND_NOT_FOUND, `Refund ${refundId} not found`);
    }

    if (refund.status !== 'processing') {
      return createServiceError(
        ErrorCodes.INVALID_STATUS_TRANSITION,
        `Cannot process in current status: ${refund.status}`
      );
    }

    const hasPhysicalProducts = refund.products.some(p => p.productType === 'physical');
    if (hasPhysicalProducts) {
      return createServiceError(
        ErrorCodes.INVALID_STATUS_TRANSITION,
        'Cannot auto-process refunds with physical products'
      );
    }

    refund.status = 'refunded';
    addStatusHistory(refund, 'refunded', 'Virtual product refund completed');
    
    persistData();
    return refund;
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    return createServiceError(ErrorCodes.INTERNAL_ERROR, message);
  }
}

export function getRefundById(refundId: string): Refund | ServiceError {
  const refund = refunds.find(r => r.refundId === refundId);
  if (!refund) {
    return createServiceError(ErrorCodes.REFUND_NOT_FOUND, `Refund ${refundId} not found`);
  }
  return refund;
}

export interface QueryOptions {
  page?: number;
  pageSize?: number;
  status?: RefundStatus;
  startTime?: Date;
  endTime?: Date;
  orderId?: string;
}

export function queryRefunds(options: QueryOptions): { refunds: Refund[]; total: number; page: number; pageSize: number } {
  let filtered = [...refunds];

  if (options.status) {
    filtered = filtered.filter(r => r.status === options.status);
  }

  if (options.orderId) {
    filtered = filtered.filter(r => r.orderId === options.orderId);
  }

  if (options.startTime) {
    filtered = filtered.filter(r => r.createdAt >= options.startTime!);
  }

  if (options.endTime) {
    filtered = filtered.filter(r => r.createdAt <= options.endTime!);
  }

  const page = options.page ?? 1;
  const pageSize = options.pageSize ?? 10;
  const startIndex = (page - 1) * pageSize;
  const pagedRefunds = filtered.slice(startIndex, startIndex + pageSize);

  return {
    refunds: pagedRefunds,
    total: filtered.length,
    page,
    pageSize
  };
}

export function updateServiceConfig(updates: { refundPeriodDays?: number; virtualProductRefundThreshold?: number }): { success: boolean } {
  updateConfig(updates);
  persistData();
  return { success: true };
}

export function getServiceConfig() {
  return getConfig();
}
