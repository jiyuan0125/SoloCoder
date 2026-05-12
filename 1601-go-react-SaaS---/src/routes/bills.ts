import { Router, Request, Response } from 'express';
import { billingService } from '../services/billingService';
import { Bill } from '../types';
import { formatMoneyForResponse } from '../utils/currency';

const router = Router();

router.get('/customer/:customerId', (req: Request, res: Response) => {
  const { customerId } = req.params;
  const bills = billingService.getBillsByCustomer(customerId);

  const response = bills.map((bill: Bill) => ({
    ...bill,
    baseFee: formatMoneyForResponse(bill.baseFeeCents),
    usageFee: formatMoneyForResponse(bill.usageFeeCents),
    totalAmount: formatMoneyForResponse(bill.totalAmountCents),
  }));

  res.json(response);
});

router.get('/subscription/:subscriptionId', (req: Request, res: Response) => {
  const { subscriptionId } = req.params;
  const bills = billingService.getBillsBySubscription(subscriptionId);

  const response = bills.map((bill: Bill) => ({
    ...bill,
    baseFee: formatMoneyForResponse(bill.baseFeeCents),
    usageFee: formatMoneyForResponse(bill.usageFeeCents),
    totalAmount: formatMoneyForResponse(bill.totalAmountCents),
  }));

  res.json(response);
});

router.post('/:billId/pay', (req: Request, res: Response) => {
  const { billId } = req.params;
  const result = billingService.payBill(billId);

  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }

  res.json({ success: true });
});

export default router;
