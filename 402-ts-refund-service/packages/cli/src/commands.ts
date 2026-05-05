import { Command, getHelpText } from './args.js';
import { request } from './api.js';
import { Refund, Order, Config } from '@refund/shared';
import {
  formatRefund,
  formatRefundList,
  formatOrder,
  formatConfig,
  formatSuccess,
  formatError
} from './formatter.js';

export async function runCommand(command: Command): Promise<void> {
  switch (command.type) {
    case 'help':
      console.log(getHelpText());
      break;

    case 'create-refund': {
      const response = await request<Refund>({
        method: 'POST',
        path: '/api/refunds',
        body: {
          orderId: command.orderId,
          userId: command.userId,
          refundReason: command.reason,
          productIds: command.productIds
        }
      });

      if (response.success && response.data) {
        console.log(formatRefund(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'get-refund': {
      const response = await request<Refund>({
        method: 'GET',
        path: `/api/refunds/${command.refundId}`
      });

      if (response.success && response.data) {
        console.log(formatRefund(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'list-refunds': {
      const params = new URLSearchParams();
      if (command.page !== undefined) params.append('page', command.page.toString());
      if (command.pageSize !== undefined) params.append('pageSize', command.pageSize.toString());
      if (command.status) params.append('status', command.status);
      if (command.startTime) params.append('startTime', command.startTime);
      if (command.endTime) params.append('endTime', command.endTime);
      if (command.orderId) params.append('orderId', command.orderId);

      const queryString = params.toString();
      const path = queryString ? `/api/refunds?${queryString}` : '/api/refunds';

      const response = await request<{ refunds: Refund[]; total: number; page: number; pageSize: number }>({
        method: 'GET',
        path
      });

      if (response.success && response.data) {
        console.log(formatRefundList(
          response.data.refunds,
          response.data.total,
          response.data.page,
          response.data.pageSize
        ));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'submit-logistics': {
      const response = await request<Refund>({
        method: 'POST',
        path: '/api/refunds/logistics',
        body: {
          refundId: command.refundId,
          logisticsNumber: command.logisticsNumber
        }
      });

      if (response.success && response.data) {
        console.log(formatSuccess('物流单号已提交'));
        console.log(formatRefund(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'warehouse-confirm': {
      const response = await request<Refund>({
        method: 'POST',
        path: '/api/refunds/warehouse-confirm',
        body: {
          refundId: command.refundId,
          received: command.received
        }
      });

      if (response.success && response.data) {
        console.log(formatSuccess(command.received ? '仓库已确认收货，退款已完成' : '仓库拒绝收货'));
        console.log(formatRefund(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'get-config': {
      const response = await request<Config>({
        method: 'GET',
        path: '/api/config'
      });

      if (response.success && response.data) {
        console.log(formatConfig(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'update-config': {
      const body: Record<string, number> = {};
      if (command.refundPeriodDays !== undefined) {
        body.refundPeriodDays = command.refundPeriodDays;
      }
      if (command.virtualThreshold !== undefined) {
        body.virtualProductRefundThreshold = command.virtualThreshold;
      }

      const response = await request<{ success: boolean }>({
        method: 'PUT',
        path: '/api/config',
        body
      });

      if (response.success) {
        console.log(formatSuccess('配置已更新'));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'add-order': {
      let order: unknown;
      try {
        order = JSON.parse(command.orderJson);
      } catch {
        console.log(formatError('INVALID_JSON', '无效的 JSON 格式'));
        process.exit(1);
      }

      const response = await request<Order>({
        method: 'POST',
        path: '/api/orders',
        body: { order }
      });

      if (response.success && response.data) {
        console.log(formatSuccess('订单已添加'));
        console.log(formatOrder(response.data));
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }

    case 'list-orders': {
      const response = await request<Order[]>({
        method: 'GET',
        path: '/api/orders'
      });

      if (response.success && response.data) {
        if (response.data.length === 0) {
          console.log('暂无订单');
        } else {
          response.data.forEach((order, index) => {
            if (index > 0) console.log('');
            console.log(formatOrder(order));
          });
        }
      } else if (response.error) {
        console.log(formatError(response.error.code, response.error.message));
        process.exit(1);
      }
      break;
    }
  }
}
