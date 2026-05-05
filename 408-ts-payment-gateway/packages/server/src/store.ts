import { Order, Payment, Refund, CallbackRecord, PaymentChannel } from "@payment-gateway/shared";

interface OrderStore {
  [orderNo: string]: Order;
}

interface PaymentStore {
  [paymentId: string]: Payment;
}

interface RefundStore {
  [refundNo: string]: Refund;
}

interface CallbackKeyStore {
  [key: string]: CallbackRecord;
}

interface OrderRefundsStore {
  [orderNo: string]: string[];
}

interface OrderPaymentsStore {
  [orderNo: string]: string[];
}

class Store {
  private orders: OrderStore = {};
  private payments: PaymentStore = {};
  private refunds: RefundStore = {};
  private callbackKeys: CallbackKeyStore = {};
  private orderRefunds: OrderRefundsStore = {};
  private orderPayments: OrderPaymentsStore = {};

  public getOrder(orderNo: string): Order | null {
    return this.orders[orderNo] ?? null;
  }

  public saveOrder(order: Order): void {
    this.orders[order.orderNo] = { ...order, updatedAt: Date.now() };
  }

  public getPayment(paymentId: string): Payment | null {
    return this.payments[paymentId] ?? null;
  }

  public savePayment(payment: Payment): void {
    this.payments[payment.paymentId] = payment;
    if (!this.orderPayments[payment.orderNo]) {
      this.orderPayments[payment.orderNo] = [];
    }
    if (!this.orderPayments[payment.orderNo].includes(payment.paymentId)) {
      this.orderPayments[payment.orderNo].push(payment.paymentId);
    }
  }

  public getPaymentsByOrder(orderNo: string): Payment[] {
    const paymentIds = this.orderPayments[orderNo] ?? [];
    return paymentIds.map((id) => this.payments[id]).filter((p): p is Payment => p !== undefined);
  }

  public getRefund(refundNo: string): Refund | null {
    return this.refunds[refundNo] ?? null;
  }

  public saveRefund(refund: Refund): void {
    this.refunds[refund.refundNo] = refund;
    if (!this.orderRefunds[refund.orderNo]) {
      this.orderRefunds[refund.orderNo] = [];
    }
    if (!this.orderRefunds[refund.orderNo].includes(refund.refundNo)) {
      this.orderRefunds[refund.orderNo].push(refund.refundNo);
    }
  }

  public getRefundsByOrder(orderNo: string): Refund[] {
    const refundNos = this.orderRefunds[orderNo] ?? [];
    return refundNos.map((no) => this.refunds[no]).filter((r): r is Refund => r !== undefined);
  }

  public getCallbackRecord(orderNo: string, channel: PaymentChannel): CallbackRecord | null {
    const key = `${channel}:${orderNo}`;
    return this.callbackKeys[key] ?? null;
  }

  public saveCallbackRecord(record: CallbackRecord): void {
    const key = `${record.channel}:${record.orderNo}`;
    this.callbackKeys[key] = record;
  }

  public getPendingConfirmOrders(): Order[] {
    return Object.values(this.orders).filter(
      (order) => order.status === "pending_confirm"
    );
  }

  public getProcessingOrders(): Order[] {
    return Object.values(this.orders).filter(
      (order) => order.status === "processing"
    );
  }

  public getAllOrders(): Order[] {
    return Object.values(this.orders);
  }

  public getAllRefunds(): Refund[] {
    return Object.values(this.refunds);
  }
}

export const store = new Store();
