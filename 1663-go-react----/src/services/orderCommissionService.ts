import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Order, Commission, CreateOrderRequest, SettlementStats, Channel } from '../types';
import { attributionService } from './attributionService';

export class OrderCommissionService {
  createOrder(data: CreateOrderRequest): { order: Order; commission: Commission | null } {
    // 验证订单号
    if (!data.order_id) {
      throw new Error('订单号不能为空');
    }

    // 验证用户ID
    if (!data.user_id) {
      throw new Error('用户ID不能为空');
    }

    // 验证订单金额
    if (data.amount <= 0) {
      throw new Error('订单金额必须大于0');
    }

    // 检查订单号重复
    const existingOrder = db.prepare('SELECT id FROM orders WHERE external_order_id = ?').get(data.order_id);
    if (existingOrder) {
      throw new Error('订单号已存在');
    }

    // 创建订单
    const orderId = uuidv4();
    const order: Order = {
      id: orderId,
      external_order_id: data.order_id,
      user_id: data.user_id,
      amount: data.amount,
      status: 'completed',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    const orderStmt = db.prepare(`
      INSERT INTO orders (id, external_order_id, user_id, amount, status, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);

    orderStmt.run(
      order.id,
      order.external_order_id,
      order.user_id,
      order.amount,
      order.status,
      order.created_at,
      order.updated_at
    );

    // 检查是否有佣金记录
    const existingCommission = db.prepare('SELECT * FROM commissions WHERE order_id = ?').get(orderId) as Commission | undefined;
    if (existingCommission) {
      throw new Error('订单已归因');
    }

    // 查找有效的归因
    const attribution = attributionService.getActiveAttribution(data.user_id);
    
    let commission: Commission | null = null;
    if (attribution) {
      // 检查渠道是否有效
      const channel = db.prepare('SELECT * FROM channels WHERE id = ?').get(attribution.channel_id) as Channel | undefined;
      
      if (channel && channel.status === 'active') {
        // 计算佣金（精确到分，整数运算）
        const commissionAmount = Math.round(data.amount * channel.commission_rate);
        
        const commissionId = uuidv4();
        commission = {
          id: commissionId,
          order_id: orderId,
          channel_id: attribution.channel_id,
          amount: commissionAmount,
          status: 'pending',
          settlement_month: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };

        const commissionStmt = db.prepare(`
          INSERT INTO commissions (id, order_id, channel_id, amount, status, settlement_month, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `);

        commissionStmt.run(
          commission.id,
          commission.order_id,
          commission.channel_id,
          commission.amount,
          commission.status,
          commission.settlement_month,
          commission.created_at,
          commission.updated_at
        );
      } else if (channel && channel.status === 'inactive') {
        // 无效渠道的订单佣金自动标为已取消
        const commissionId = uuidv4();
        commission = {
          id: commissionId,
          order_id: orderId,
          channel_id: attribution.channel_id,
          amount: 0,
          status: 'cancelled',
          settlement_month: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };

        const commissionStmt = db.prepare(`
          INSERT INTO commissions (id, order_id, channel_id, amount, status, settlement_month, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `);

        commissionStmt.run(
          commission.id,
          commission.order_id,
          commission.channel_id,
          commission.amount,
          commission.status,
          commission.settlement_month,
          commission.created_at,
          commission.updated_at
        );
      }
    }

    return { order, commission };
  }

  getOrderByExternalId(externalOrderId: string): Order | undefined {
    return db.prepare('SELECT * FROM orders WHERE external_order_id = ?').get(externalOrderId) as Order | undefined;
  }

  refundOrder(orderIdentifier: string): void {
    // 先尝试通过内部 ID 查找，再尝试通过外部订单号查找
    let order = db.prepare('SELECT * FROM orders WHERE id = ?').get(orderIdentifier) as Order | undefined;
    if (!order) {
      order = db.prepare('SELECT * FROM orders WHERE external_order_id = ?').get(orderIdentifier) as Order | undefined;
    }
    
    if (!order) {
      throw new Error('订单不存在');
    }

    // 更新订单状态
    db.prepare(`
      UPDATE orders 
      SET status = ?, updated_at = ? 
      WHERE id = ?
    `).run('refunded', new Date().toISOString(), order.id);

    // 检查佣金记录
    const commission = db.prepare('SELECT * FROM commissions WHERE order_id = ?').get(order.id) as Commission | undefined;
    if (commission) {
      if (commission.status === 'settled') {
        // 已结算的佣金不能直接取消
        throw new Error('已结算的佣金需要先走退款流程');
      } else if (commission.status === 'pending') {
        // 待结算的佣金标为已取消
        db.prepare(`
          UPDATE commissions 
          SET status = ?, updated_at = ? 
          WHERE order_id = ?
        `).run('cancelled', new Date().toISOString(), order.id);
      }
    }
  }

  getOrderById(id: string): Order | undefined {
    return db.prepare('SELECT * FROM orders WHERE id = ?').get(id) as Order | undefined;
  }

  getCommissionById(id: string): Commission | undefined {
    return db.prepare('SELECT * FROM commissions WHERE id = ?').get(id) as Commission | undefined;
  }

  getCommissionsByChannelId(channelId: string): Commission[] {
    return db.prepare(`
      SELECT * FROM commissions 
      WHERE channel_id = ? 
      ORDER BY created_at DESC
    `).all(channelId) as Commission[];
  }

  getSettlementStats(): SettlementStats[] {
    return db.prepare(`
      SELECT 
        c.channel_id,
        ch.name as channel_name,
        c.settlement_month,
        COUNT(c.id) as total_orders,
        SUM(c.amount) as total_commission,
        c.status
      FROM commissions c
      JOIN channels ch ON c.channel_id = ch.id
      GROUP BY c.channel_id, c.settlement_month, c.status
      ORDER BY c.settlement_month DESC, ch.name
    `).all() as SettlementStats[];
  }

  settleCommissions(channelId: string, settlementMonth: string): void {
    const now = new Date().toISOString();
    
    // 更新待结算的佣金为已结算
    db.prepare(`
      UPDATE commissions 
      SET status = ?, settlement_month = ?, updated_at = ? 
      WHERE channel_id = ? AND status = 'pending'
    `).run('settled', settlementMonth, now, channelId);
  }
}

export const orderCommissionService = new OrderCommissionService();
