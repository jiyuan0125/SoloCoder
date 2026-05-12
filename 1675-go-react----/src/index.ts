import express, { Request, Response } from 'express';
import {
  registerReceivable,
  listReceivables,
  getReceivableById,
  submitFinancingApplication,
  listFinancings,
  getFinancingById,
  approveAndFund,
  settleFinancing,
} from './factoring';
import {
  registerWarehouseReceipt,
  listReceipts,
  getReceiptById,
  updateReceiptMarketValue,
  createPledge,
  listPledges,
  getPledgeById,
  closePledge,
  marginCheck,
} from './pledge';

const app = express();
const PORT = process.env.PORT || 8105;

app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.post('/api/receivables', (req: Request, res: Response) => {
  const result = registerReceivable(req.body);
  res.status(result.code).json(result);
});

app.get('/api/receivables', (req: Request, res: Response) => {
  const supplierId = req.query.supplier_id as string | undefined;
  const data = listReceivables(supplierId);
  res.json({ code: 200, message: 'success', data });
});

app.get('/api/receivables/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const data = getReceivableById(id);
  if (!data) {
    res.status(404).json({ code: 404, message: '应收账款不存在' });
    return;
  }
  res.json({ code: 200, message: 'success', data });
});

app.post('/api/financing/factoring', (req: Request, res: Response) => {
  const result = submitFinancingApplication(req.body);
  res.status(result.code).json(result);
});

app.get('/api/financing/factoring', (req: Request, res: Response) => {
  const status = req.query.status as string | undefined;
  const data = listFinancings(status);
  res.json({ code: 200, message: 'success', data });
});

app.get('/api/financing/factoring/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const data = getFinancingById(id);
  if (!data) {
    res.status(404).json({ code: 404, message: '融资申请不存在' });
    return;
  }
  res.json({ code: 200, message: 'success', data });
});

app.post('/api/financing/factoring/:id/fund', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const result = approveAndFund(id);
  res.status(result.code).json(result);
});

app.post('/api/financing/factoring/:id/settle', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const paymentDate = req.body.payment_date as string | undefined;
  const result = settleFinancing(id, paymentDate);
  res.status(result.code).json(result);
});

app.post('/api/warehouse-receipts', (req: Request, res: Response) => {
  const result = registerWarehouseReceipt(req.body);
  res.status(result.code).json(result);
});

app.get('/api/warehouse-receipts', (req: Request, res: Response) => {
  const supplierId = req.query.supplier_id as string | undefined;
  const data = listReceipts(supplierId);
  res.json({ code: 200, message: 'success', data });
});

app.get('/api/warehouse-receipts/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const data = getReceiptById(id);
  if (!data) {
    res.status(404).json({ code: 404, message: '仓单不存在' });
    return;
  }
  res.json({ code: 200, message: 'success', data });
});

app.put('/api/warehouse-receipts/:id/price', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const newUnitPrice = req.body.unit_price as number;
  const result = updateReceiptMarketValue(id, newUnitPrice);
  res.status(result.code).json(result);
});

app.post('/api/financing/pledge', (req: Request, res: Response) => {
  const result = createPledge(req.body);
  res.status(result.code).json(result);
});

app.get('/api/financing/pledge', (req: Request, res: Response) => {
  const status = req.query.status as string | undefined;
  const data = listPledges(status);
  res.json({ code: 200, message: 'success', data });
});

app.get('/api/financing/pledge/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const data = getPledgeById(id);
  if (!data) {
    res.status(404).json({ code: 404, message: '质押融资不存在' });
    return;
  }
  res.json({ code: 200, message: 'success', data });
});

app.post('/api/financing/pledge/:id/margin-check', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const marketValue = req.body.market_value as number | undefined;
  const result = marginCheck(id, marketValue);
  res.status(result.code).json(result);
});

app.post('/api/financing/pledge/:id/close', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  const result = closePledge(id);
  res.status(result.code).json(result);
});

app.listen(PORT, () => {
  console.log(`供应链金融平台后端服务已启动，监听端口 ${PORT}`);
});
