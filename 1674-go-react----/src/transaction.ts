import { db } from './database';
import { Currency, Transaction, TransactionStatus, ExchangeRateInfo } from './types';
import { getAccount, creditAccount, debitAccount } from './account';
import { getExchangeRate as getRate, convertCurrency as convert } from './exchangeRate';

const generateId = (): string => {
  return 'tx_' + Math.random().toString(36).substring(2, 15);
};

const SINGLE_TRANSACTION_LIMIT_USD = 50000;
const DAILY_LIMIT_USD = 100000;

const getDailyUsdVolume = async (): Promise<number> => {
  const now = Date.now();
  const oneDayAgo = now - 24 * 60 * 60 * 1000;
  
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT * FROM transactions WHERE status IN (?, ?, ?) AND created_at > ?`,
      [TransactionStatus.COMPLETED, TransactionStatus.PENDING_APPROVAL, TransactionStatus.AWAITING_RATE],
      oneDayAgo,
      async (err: Error | null, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }
        
        let totalUsd = 0;
        
        for (const tx of rows) {
          if (tx.source_currency === Currency.USD) {
            totalUsd += tx.source_amount;
          } else {
            try {
              const rate = await getRate(Currency.USD);
              const sourceRate = await getRate(tx.source_currency as Currency);
              
              if (rate && sourceRate) {
                const cnyAmount = Math.floor((tx.source_amount * sourceRate.rate) / 1000000);
                const usdAmount = Math.floor((cnyAmount * 1000000) / rate.rate);
                totalUsd += usdAmount;
              }
            } catch (e) {
            }
          }
        }
        
        resolve(totalUsd);
      }
    );
  });
};

export const createTransaction = async (
  sourceAccountId: string,
  targetAccountId: string,
  sourceCurrency: Currency,
  targetCurrency: Currency,
  sourceAmount: number
): Promise<Transaction> => {
  const id = generateId();
  const now = Date.now();
  
  return new Promise((resolve, reject) => {
    db.run(
      `INSERT INTO transactions (id, source_account_id, target_account_id, source_currency, target_currency, source_amount, status, created_at, updated_at) 
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [id, sourceAccountId, targetAccountId, sourceCurrency, targetCurrency, sourceAmount, TransactionStatus.PENDING, now, now],
      (err) => {
        if (err) {
          reject(err);
          return;
        }
        resolve({
          id,
          source_account_id: sourceAccountId,
          target_account_id: targetAccountId,
          source_currency: sourceCurrency,
          target_currency: targetCurrency,
          source_amount: sourceAmount,
          status: TransactionStatus.PENDING,
          created_at: now,
          updated_at: now
        });
      }
    );
  });
};

export const getTransaction = async (id: string): Promise<Transaction | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM transactions WHERE id = ?', [id], (err, row) => {
      if (err) {
        reject(err);
        return;
      }
      if (row) {
        resolve(row as Transaction);
      } else {
        resolve(null);
      }
    });
  });
};

export const updateTransactionStatus = async (
  id: string,
  status: TransactionStatus,
  targetAmount?: number,
  exchangeRateInfo?: ExchangeRateInfo
): Promise<void> => {
  const now = Date.now();
  
  return new Promise((resolve, reject) => {
    db.run(
      `UPDATE transactions SET status = ?, target_amount = ?, exchange_rate_info = ?, updated_at = ? WHERE id = ?`,
      [status, targetAmount || null, exchangeRateInfo ? JSON.stringify(exchangeRateInfo) : null, now, id],
      (err) => {
        if (err) {
          reject(err);
          return;
        }
        resolve();
      }
    );
  });
};

export const updateTransactionAmount = async (
  id: string,
  sourceAmount: number
): Promise<void> => {
  const now = Date.now();
  
  return new Promise((resolve, reject) => {
    db.run(
      `UPDATE transactions SET source_amount = ?, updated_at = ? WHERE id = ?`,
      [sourceAmount, now, id],
      (err) => {
        if (err) {
          reject(err);
          return;
        }
        resolve();
      }
    );
  });
};

export const getPendingTransactionsAwaitingRate = async (): Promise<Transaction[]> => {
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT * FROM transactions WHERE status = ? ORDER BY created_at ASC`,
      [TransactionStatus.AWAITING_RATE],
      (err, rows) => {
        if (err) {
          reject(err);
          return;
        }
        resolve(rows as Transaction[]);
      }
    );
  });
};

export const cancelTransaction = async (id: string): Promise<void> => {
  await updateTransactionStatus(id, TransactionStatus.CANCELLED);
};

export const checkDailyLimit = async (sourceAmount: number, sourceCurrency: Currency): Promise<{ willExceed: boolean; currentVolume: number }> => {
  const currentVolume = await getDailyUsdVolume();
  
  let amountInUsd = sourceAmount;
  if (sourceCurrency !== Currency.USD) {
    const usdRate = await getRate(Currency.USD);
    const sourceRate = await getRate(sourceCurrency);
    
    if (usdRate && sourceRate) {
      const cnyAmount = Math.floor((sourceAmount * sourceRate.rate) / 1000000);
      amountInUsd = Math.floor((cnyAmount * 1000000) / usdRate.rate);
    }
  }
  
  const willExceed = (currentVolume + amountInUsd) > DAILY_LIMIT_USD;
  
  return { willExceed, currentVolume };
};

export const processTransaction = async (transaction: Transaction): Promise<{
  status: TransactionStatus;
  message?: string;
  targetAmount?: number;
  rateInfo?: ExchangeRateInfo;
}> => {
  const sourceAccount = await getAccount(transaction.source_account_id);
  const targetAccount = await getAccount(transaction.target_account_id);
  
  if (!sourceAccount || !targetAccount) {
    return { status: TransactionStatus.CANCELLED, message: '账户不存在' };
  }
  
  if (sourceAccount.balance < transaction.source_amount) {
    return { status: TransactionStatus.CANCELLED, message: '余额不足' };
  }
  
  try {
    const { amount: targetAmount, rateInfo } = await convert(
      transaction.source_amount,
      transaction.source_currency,
      transaction.target_currency
    );
    
    let amountInUsd = transaction.source_amount;
    if (transaction.source_currency !== Currency.USD) {
      const usdRate = await getRate(Currency.USD);
      const sourceRate = await getRate(transaction.source_currency);
      
      if (usdRate && sourceRate) {
        const cnyAmount = Math.floor((transaction.source_amount * sourceRate.rate) / 1000000);
        amountInUsd = Math.floor((cnyAmount * 1000000) / usdRate.rate);
      }
    }
    
    const { willExceed } = await checkDailyLimit(transaction.source_amount, transaction.source_currency);
    
    if (amountInUsd > SINGLE_TRANSACTION_LIMIT_USD) {
      return {
        status: TransactionStatus.PENDING_APPROVAL,
        message: '待合规审批'
      };
    }
    
    const success = await debitAccount(transaction.source_account_id, transaction.source_amount);
    if (!success) {
      return { status: TransactionStatus.CANCELLED, message: '余额不足' };
    }
    
    await creditAccount(transaction.target_account_id, targetAmount);
    
    if (willExceed) {
      console.warn(`警告: 日累计交易金额即将超过 10 万美元阈值`);
    }
    
    return {
      status: TransactionStatus.COMPLETED,
      targetAmount,
      rateInfo
    };
    
  } catch (error) {
    if (error instanceof Error && error.message === '汇率过期') {
      return {
        status: TransactionStatus.AWAITING_RATE,
        message: '汇率过期等待更新'
      };
    }
    return { status: TransactionStatus.CANCELLED, message: '交易失败' };
  }
};
