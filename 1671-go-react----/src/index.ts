import express, { Request, Response, NextFunction } from 'express';
import { initializeChannels } from './db';
import { startDegradationChecker } from './healthMonitor';
import {
  createPayment,
  processRefund,
  getPayment,
  MerchantOrderExistsError,
  InvalidAmountError,
  PaymentNotFoundError,
  InvalidStatusError,
  RefundExceedsAmountError,
  AlreadyRefundedError,
  AllChannelsDegradedError,
  ChannelNotFoundError,
} from './paymentService';
import { CreatePaymentRequest, RefundRequest, RoutingStrategy } from './types';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

initializeChannels();
startDegradationChecker();

const validStrategies: RoutingStrategy[] = ['RATE', 'SUCCESS_RATE', 'SPECIFIED'];

app.post('/api/payments', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body as Partial<CreatePaymentRequest>;

    if (!body.merchantOrderNo || typeof body.merchantOrderNo !== 'string') {
      return res.status(400).json({ error: '缺少商户订单号' });
    }

    if (typeof body.amount !== 'number') {
      return res.status(400).json({ error: '金额必须为数字' });
    }

    if (!body.strategy || !validStrategies.includes(body.strategy)) {
      return res.status(400).json({ error: '无效的路由策略，可选值: RATE, SUCCESS_RATE, SPECIFIED' });
    }

    if (body.strategy === 'SPECIFIED' && !body.channelId) {
      return res.status(400).json({ error: '指定渠道策略需要提供 channelId' });
    }

    const result = await createPayment({
      merchantOrderNo: body.merchantOrderNo,
      amount: body.amount,
      strategy: body.strategy,
      channelId: body.channelId,
    });

    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.get('/api/payments/:id', (req: Request, res: Response) => {
  const payment = getPayment(req.params.id);
  if (!payment) {
    return res.status(404).json({ error: '支付记录不存在' });
  }
  res.json(payment);
});

app.post('/api/payments/:id/refunds', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body as Partial<RefundRequest>;

    if (typeof body.refundAmount !== 'number') {
      return res.status(400).json({ error: '退款金额必须为数字' });
    }

    const result = await processRefund({
      paymentId: req.params.id,
      refundAmount: body.refundAmount,
    });

    res.json(result);
  } catch (err) {
    next(err);
  }
});

app.use((err: any, _req: Request, res: Response, _next: NextFunction) => {
  console.error('Error:', err);

  if (err instanceof InvalidAmountError) {
    return res.status(400).json({ error: err.message });
  }

  if (err instanceof MerchantOrderExistsError) {
    return res.status(409).json({ error: err.message });
  }

  if (err instanceof ChannelNotFoundError) {
    return res.status(404).json({ error: err.message });
  }

  if (err instanceof PaymentNotFoundError) {
    return res.status(404).json({ error: err.message });
  }

  if (err instanceof InvalidStatusError) {
    return res.status(400).json({ error: err.message });
  }

  if (err instanceof RefundExceedsAmountError) {
    return res.status(400).json({ error: err.message });
  }

  if (err instanceof AlreadyRefundedError) {
    return res.status(409).json({ error: err.message });
  }

  if (err instanceof AllChannelsDegradedError) {
    return res.status(503).json({ error: err.message });
  }

  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`支付网关服务运行在 http://localhost:${PORT}`);
});
