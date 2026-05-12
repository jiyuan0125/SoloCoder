import express from 'express';
import { db } from './db';
import { ChannelTransaction, InternalOrder, DisputeRequest } from './types';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.post('/api/channel-transactions', (req, res) => {
  const tx: ChannelTransaction = req.body;
  
  try {
    const stmt = db.prepare(`
      INSERT INTO channel_transactions 
      (transaction_id, transaction_time, amount, counterparty, channel_name, status)
      VALUES (?, ?, ?, ?, ?, ?)
    `);
    const result = stmt.run(
      tx.transaction_id,
      tx.transaction_time,
      tx.amount,
      tx.counterparty,
      tx.channel_name,
      tx.status
    );
    
    const queueStmt = db.prepare(`
      INSERT INTO processing_queue (entity_type, entity_id) VALUES ('channel', ?)
    `);
    queueStmt.run(tx.transaction_id);
    
    res.status(201).json({ id: result.lastInsertRowid, ...tx });
  } catch (err: any) {
    if (err.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      res.status(409).json({ error: '流水号已存在' });
    } else {
      res.status(500).json({ error: err.message });
    }
  }
});

app.post('/api/internal-orders', (req, res) => {
  const order: InternalOrder = req.body;
  
  try {
    const stmt = db.prepare(`
      INSERT INTO internal_orders 
      (order_id, order_time, amount, payment_method, status)
      VALUES (?, ?, ?, ?, ?)
    `);
    const result = stmt.run(
      order.order_id,
      order.order_time,
      order.amount,
      order.payment_method,
      order.status
    );
    
    const queueStmt = db.prepare(`
      INSERT INTO processing_queue (entity_type, entity_id) VALUES ('order', ?)
    `);
    queueStmt.run(order.order_id);
    
    res.status(201).json({ id: result.lastInsertRowid, ...order });
  } catch (err: any) {
    if (err.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      res.status(409).json({ error: '订单号已存在' });
    } else {
      res.status(500).json({ error: err.message });
    }
  }
});

app.post('/api/reconcile', (req, res) => {
  const { report_date } = req.body;
  
  const existingReport = db.prepare(`
    SELECT id FROM reconciliation_reports WHERE report_date = ?
  `).get(report_date);
  
  if (existingReport) {
    return res.status(409).json({ error: '该日期已对账' });
  }

  const transaction = db.transaction(() => {
    const startTime = new Date();
    
    const insertReport = db.prepare(`
      INSERT INTO reconciliation_reports (report_date, matched_count, channel_extra_count, order_extra_count, reconciliation_time)
      VALUES (?, 0, 0, 0, ?)
    `);
    const reportResult = insertReport.run(report_date, startTime.toISOString());
    const reportId = reportResult.lastInsertRowid as number;
    
    const transactions = db.prepare(`
      SELECT * FROM channel_transactions 
      WHERE matched_report_id IS NULL
      AND transaction_time >= ? 
      AND transaction_time < ?
    `).all(
      new Date(report_date).toISOString(),
      new Date(new Date(report_date).getTime() + 24 * 60 * 60 * 1000).toISOString()
    ) as any[];
    
    const orders = db.prepare(`
      SELECT * FROM internal_orders 
      WHERE matched_report_id IS NULL
      AND order_time >= ? 
      AND order_time < ?
    `).all(
      new Date(report_date).toISOString(),
      new Date(new Date(report_date).getTime() + 24 * 60 * 60 * 1000).toISOString()
    ) as any[];

    const matchedTransactions = new Set<string>();
    const matchedOrders = new Set<string>();
    let matchCount = 0;
    
    const updateTransaction = db.prepare(`
      UPDATE channel_transactions 
      SET matched_order_id = ?, matched_report_id = ?
      WHERE transaction_id = ?
    `);
    
    const updateOrder = db.prepare(`
      UPDATE internal_orders 
      SET matched_transaction_id = ?, matched_report_id = ?
      WHERE order_id = ?
    `);

    for (const tx of transactions) {
      if (matchedTransactions.has(tx.transaction_id)) continue;
      
      for (const order of orders) {
        if (matchedOrders.has(order.order_id)) continue;
        
        const txTime = new Date(tx.transaction_time).getTime();
        const orderTime = new Date(order.order_time).getTime();
        const timeDiff = Math.abs(txTime - orderTime) / (1000 * 60);
        
        if (tx.amount === order.amount && timeDiff <= 5) {
          updateTransaction.run(order.order_id, reportId, tx.transaction_id);
          updateOrder.run(tx.transaction_id, reportId, order.order_id);
          
          matchedTransactions.add(tx.transaction_id);
          matchedOrders.add(order.order_id);
          matchCount++;
          break;
        }
      }
    }
    
    const channelExtra = transactions.filter(tx => !matchedTransactions.has(tx.transaction_id));
    const orderExtra = orders.filter(order => !matchedOrders.has(order.order_id));
    
    const updateReport = db.prepare(`
      UPDATE reconciliation_reports 
      SET matched_count = ?, channel_extra_count = ?, order_extra_count = ?
      WHERE id = ?
    `);
    updateReport.run(matchCount, channelExtra.length, orderExtra.length, reportId);
    
    return {
      report_id: reportId,
      matched_count: matchCount,
      channel_extra_count: channelExtra.length,
      order_extra_count: orderExtra.length,
      reconciliation_time: startTime.toISOString()
    };
  });
  
  try {
    const result = transaction();
    res.status(201).json(result);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/reports', (req, res) => {
  const reports = db.prepare('SELECT * FROM reconciliation_reports ORDER BY report_date DESC').all();
  res.json(reports);
});

app.get('/api/reports/:id', (req, res) => {
  const report = db.prepare('SELECT * FROM reconciliation_reports WHERE id = ?').get(req.params.id);
  if (!report) return res.status(404).json({ error: '报告不存在' });
  
  const adjustments = db.prepare('SELECT * FROM adjustment_records WHERE report_id = ?').all(req.params.id);
  res.json({ ...report, adjustments });
});

app.post('/api/disputes', (req, res) => {
  const { report_id, entity_type, entity_id, action, supplement_order }: DisputeRequest = req.body;
  
  const report = db.prepare('SELECT * FROM reconciliation_reports WHERE id = ?').get(report_id);
  if (!report) return res.status(404).json({ error: '报告不存在' });
  
  const transaction = db.transaction(() => {
    if (action === 'supplement' && entity_type === 'channel') {
      const tx = db.prepare('SELECT * FROM channel_transactions WHERE transaction_id = ?').get(entity_id) as any;
      if (!tx) return { error: '流水不存在' };
      
      const newOrderId = supplement_order?.order_id || `SUPP-${Date.now()}`;
      const orderAmount = supplement_order?.amount ?? tx.amount;
      
      try {
        const insertOrder = db.prepare(`
          INSERT INTO internal_orders (order_id, order_time, amount, payment_method, status, is_supplemental)
          VALUES (?, ?, ?, ?, ?, 1)
        `);
        insertOrder.run(
          newOrderId,
          supplement_order?.order_time || tx.transaction_time,
          orderAmount,
          supplement_order?.payment_method || 'supplemental',
          supplement_order?.status || 'completed'
        );
        
        const difference = tx.amount - orderAmount;
        const insertAdjustment = db.prepare(`
          INSERT INTO adjustment_records 
          (report_id, type, description, original_amount, adjusted_amount, difference, transaction_id, order_id)
          VALUES (?, 'supplement', '补录缺失订单', ?, ?, ?, ?, ?)
        `);
        insertAdjustment.run(
          report_id,
          tx.amount,
          orderAmount,
          difference,
          entity_id,
          newOrderId
        );
        
        return { success: true, order_id: newOrderId, difference };
      } catch (err: any) {
        if (err.code === 'SQLITE_CONSTRAINT_UNIQUE') {
          return { error: '订单号已存在', code: 409 };
        }
        throw err;
      }
    } else if (action === 'suspend') {
      const insertAdjustment = db.prepare(`
        INSERT INTO adjustment_records (report_id, type, description, transaction_id, order_id)
        VALUES (?, 'suspend', '挂账处理', ?, ?)
      `);
      insertAdjustment.run(
        report_id,
        entity_type === 'channel' ? entity_id : null,
        entity_type === 'order' ? entity_id : null
      );
      return { success: true };
    } else if (action === 'adjust') {
      const insertAdjustment = db.prepare(`
        INSERT INTO adjustment_records (report_id, type, description, transaction_id, order_id)
        VALUES (?, 'adjust', '调平确认差异', ?, ?)
      `);
      insertAdjustment.run(
        report_id,
        entity_type === 'channel' ? entity_id : null,
        entity_type === 'order' ? entity_id : null
      );
      return { success: true };
    }
    
    return { error: '无效的处理方式' };
  });
  
  const result = transaction();
  if ((result as any).error) {
    return res.status((result as any).code || 400).json({ error: (result as any).error });
  }
  
  res.json(result);
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
