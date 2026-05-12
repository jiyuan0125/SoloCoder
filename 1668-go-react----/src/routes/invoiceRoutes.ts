import { Router, Request, Response } from 'express';
import {
  createInvoice,
  createRedInvoice,
  verifyInvoice,
  getInvoiceById,
  getInvoiceByCodeAndNumber,
  listAllInvoices,
  toDisplayInvoice,
  toDisplayInvoices,
  InvoiceError
} from '../services/invoiceService';
import { CreateInvoiceRequest, VerifyInvoiceRequest } from '../types';

const router = Router();

router.post('/', async (req: Request, res: Response) => {
  try {
    const body = req.body as CreateInvoiceRequest;
    const invoice = createInvoice(body);
    res.status(201).json(toDisplayInvoice(invoice));
  } catch (error) {
    if (error instanceof InvoiceError) {
      res.status(error.statusCode).json({ error: error.message });
    } else {
      res.status(500).json({ error: (error as Error).message });
    }
  }
});

router.post('/red', async (req: Request, res: Response) => {
  try {
    const body = req.body as { originalInvoiceId: string };
    const invoice = await createRedInvoice(body.originalInvoiceId);
    res.status(201).json(toDisplayInvoice(invoice));
  } catch (error) {
    if (error instanceof InvoiceError) {
      res.status(error.statusCode).json({ error: error.message });
    } else {
      res.status(500).json({ error: (error as Error).message });
    }
  }
});

router.post('/verify', (req: Request, res: Response) => {
  try {
    const body = req.body as VerifyInvoiceRequest;
    const result = verifyInvoice(
      body.code,
      body.number,
      body.amount,
      body.taxAmount
    );
    if (!result.valid) {
      res.status(400).json(result);
    } else {
      res.json({
        ...result,
        invoice: result.invoice ? toDisplayInvoice(result.invoice) : undefined
      });
    }
  } catch (error) {
    if (error instanceof InvoiceError) {
      res.status(error.statusCode).json({ error: error.message });
    } else {
      res.status(500).json({ error: (error as Error).message });
    }
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const invoice = getInvoiceById(req.params.id);
    if (!invoice) {
      res.status(404).json({ error: '发票不存在' });
    } else {
      res.json(toDisplayInvoice(invoice));
    }
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/code/:code/number/:number', (req: Request, res: Response) => {
  try {
    const invoice = getInvoiceByCodeAndNumber(req.params.code, req.params.number);
    if (!invoice) {
      res.status(404).json({ error: '发票不存在' });
    } else {
      res.json(toDisplayInvoice(invoice));
    }
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const invoices = listAllInvoices();
    res.json(toDisplayInvoices(invoices));
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

export default router;
