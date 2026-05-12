import express, { Request, Response } from 'express';
import { initDB } from './database';
import { Currency, SUPPORTED_CURRENCIES, TransactionStatus } from './types';
import { createAccount, getAccount, creditAccount } from './account';
import { updateExchangeRate, getAllExchangeRates, simulateRatePush } from './exchangeRate';
import { 
  createTransaction, 
  getTransaction, 
  processTransaction, 
  updateTransactionStatus,
  updateTransactionAmount,
  cancelTransaction
} from './transaction';
import { initWakeUpMechanism } from './wakeUp';

const app = express();
app.use(express.json());

const isValidCurrency = (currency: string): currency is Currency => {
  return SUPPORTED_CURRENCIES.includes(currency as Currency);
};

app.post('/accounts', async (req: Request, res: Response) => {
  try {
    const { currency } = req.body;
    
    if (!currency || !isValidCurrency(currency)) {
      return res.status(400).json({ error: '币种不在支持列表里' });
    }
    
    const account = await createAccount(currency);
    res.status(201).json(account);
  } catch (error) {
    res.status(500).json({ error: '创建账户失败' });
  }
});

app.get('/accounts/:id', async (req: Request, res: Response) => {
  try {
    const account = await getAccount(req.params.id);
    if (!account) {
      return res.status(404).json({ error: '账户不存在' });
    }
    res.json(account);
  } catch (error) {
    res.status(500).json({ error: '获取账户信息失败' });
  }
});

app.post('/accounts/:id/credit', async (req: Request, res: Response) => {
  try {
    const { amount } = req.body;
    
    if (amount <= 0) {
      return res.status(400).json({ error: '金额必须为正数' });
    }
    
    const account = await getAccount(req.params.id);
    if (!account) {
      return res.status(404).json({ error: '账户不存在' });
    }
    
    await creditAccount(req.params.id, amount);
    const updatedAccount = await getAccount(req.params.id);
    res.json(updatedAccount);
  } catch (error) {
    res.status(500).json({ error: '入账失败' });
  }
});

app.post('/exchange-rates', async (req: Request, res: Response) => {
  try {
    const { currency, rate } = req.body;
    
    if (!currency || !isValidCurrency(currency)) {
      return res.status(400).json({ error: '币种不在支持列表里' });
    }
    
    if (typeof rate !== 'number' || rate <= 0) {
      return res.status(400).json({ error: '无效的汇率值' });
    }
    
    const updatedRate = await updateExchangeRate(currency, rate);
    res.status(200).json(updatedRate);
  } catch (error) {
    res.status(500).json({ error: '更新汇率失败' });
  }
});

app.get('/exchange-rates', async (_req: Request, res: Response) => {
  try {
    const rates = await getAllExchangeRates();
    res.json(rates);
  } catch (error) {
    res.status(500).json({ error: '获取汇率失败' });
  }
});

app.post('/transactions', async (req: Request, res: Response) => {
  try {
    const { 
      sourceAccountId, 
      targetAccountId, 
      sourceCurrency, 
      targetCurrency, 
      sourceAmount 
    } = req.body;
    
    if (!sourceAccountId || !targetAccountId) {
      return res.status(400).json({ error: '缺少账户信息' });
    }
    
    if (!sourceCurrency || !isValidCurrency(sourceCurrency)) {
      return res.status(400).json({ error: '源币种不在支持列表里' });
    }
    
    if (!targetCurrency || !isValidCurrency(targetCurrency)) {
      return res.status(400).json({ error: '目标币种不在支持列表里' });
    }
    
    if (typeof sourceAmount !== 'number' || sourceAmount <= 0) {
      return res.status(400).json({ error: '金额必须为正数' });
    }
    
    const transaction = await createTransaction(
      sourceAccountId,
      targetAccountId,
      sourceCurrency,
      targetCurrency,
      sourceAmount
    );
    
    const result = await processTransaction(transaction);
    
    await updateTransactionStatus(
      transaction.id,
      result.status,
      result.targetAmount,
      result.rateInfo
    );
    
    const updatedTransaction = await getTransaction(transaction.id);
    
    if (result.status === TransactionStatus.PENDING_APPROVAL) {
      return res.status(202).json({ 
        ...updatedTransaction,
        message: result.message 
      });
    }
    
    if (result.status === TransactionStatus.AWAITING_RATE) {
      return res.status(503).json({
        ...updatedTransaction,
        message: result.message
      });
    }
    
    if (result.status === TransactionStatus.CANCELLED) {
      return res.status(400).json({
        ...updatedTransaction,
        message: result.message
      });
    }
    
    res.status(200).json(updatedTransaction);
  } catch (error) {
    res.status(500).json({ error: '创建交易失败' });
  }
});

app.get('/transactions/:id', async (req: Request, res: Response) => {
  try {
    const transaction = await getTransaction(req.params.id);
    if (!transaction) {
      return res.status(404).json({ error: '交易不存在' });
    }
    res.json(transaction);
  } catch (error) {
    res.status(500).json({ error: '获取交易信息失败' });
  }
});

app.put('/transactions/:id/amount', async (req: Request, res: Response) => {
  try {
    const { sourceAmount } = req.body;
    
    if (typeof sourceAmount !== 'number' || sourceAmount <= 0) {
      return res.status(400).json({ error: '金额必须为正数' });
    }
    
    const transaction = await getTransaction(req.params.id);
    if (!transaction) {
      return res.status(404).json({ error: '交易不存在' });
    }
    
    if (transaction.status !== TransactionStatus.AWAITING_RATE) {
      return res.status(400).json({ error: '只能修改等待汇率的交易金额' });
    }
    
    await updateTransactionAmount(req.params.id, sourceAmount);
    const updatedTransaction = await getTransaction(req.params.id);
    res.json(updatedTransaction);
  } catch (error) {
    res.status(500).json({ error: '修改交易金额失败' });
  }
});

app.post('/transactions/:id/cancel', async (req: Request, res: Response) => {
  try {
    const transaction = await getTransaction(req.params.id);
    if (!transaction) {
      return res.status(404).json({ error: '交易不存在' });
    }
    
    if (transaction.status === TransactionStatus.COMPLETED) {
      return res.status(400).json({ error: '已完成的交易不能取消' });
    }
    
    await cancelTransaction(req.params.id);
    const updatedTransaction = await getTransaction(req.params.id);
    res.json(updatedTransaction);
  } catch (error) {
    res.status(500).json({ error: '取消交易失败' });
  }
});

const PORT = 8104;

const start = async () => {
  await initDB();
  initWakeUpMechanism();
  simulateRatePush();
  
  app.listen(PORT, () => {
    console.log(`跨境支付系统服务器运行在 http://localhost:${PORT}`);
  });
};

start().catch(console.error);
