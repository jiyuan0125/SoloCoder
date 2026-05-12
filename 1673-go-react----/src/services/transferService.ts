import {
  getAccount,
  createTransaction,
  executeInTransaction,
  db
} from '../database';
import { AccountType, TransactionType, Transaction } from '../types';
import { AccountNotFoundError, InsufficientBalanceError } from './accountService';

interface TransferResult {
  sourceTransaction: Transaction;
  targetTransaction: Transaction;
  sourceBalanceAfter: number;
  targetBalanceAfter: number;
}

export function transfer(
  sourceAccountId: string,
  targetAccountId: string,
  amount: number,
  description: string = 'Transfer',
  referenceId?: string
): TransferResult {
  if (amount <= 0) {
    throw new Error('Amount must be positive');
  }
  
  if (sourceAccountId === targetAccountId) {
    throw new Error('Cannot transfer to the same account');
  }
  
  return executeInTransaction(() => {
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit, status FROM accounts WHERE id = ?
    `).get(sourceAccountId) as any;
    
    if (!sourceRow) {
      throw new AccountNotFoundError('Source account not found');
    }
    
    const targetRow = db.prepare(`
      SELECT id, balance, overdraft_limit, status FROM accounts WHERE id = ?
    `).get(targetAccountId) as any;
    
    if (!targetRow) {
      throw new AccountNotFoundError('Target account not found');
    }
    
    const sourceBalanceBefore = sourceRow.balance;
    const sourceBalanceAfter = sourceBalanceBefore - amount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError();
    }
    
    const targetBalanceBefore = targetRow.balance;
    const targetBalanceAfter = targetBalanceBefore + amount;
    
    const now = Date.now();
    
    const updateResult = db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ? AND balance = ?
    `).run(sourceBalanceAfter, now, sourceAccountId, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(targetBalanceAfter, now, targetAccountId);
    
    const sourceTransaction = createTransaction(
      sourceAccountId,
      TransactionType.TRANSFER,
      -amount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      description,
      referenceId
    );
    
    const targetTransaction = createTransaction(
      targetAccountId,
      TransactionType.TRANSFER,
      amount,
      targetBalanceBefore,
      targetBalanceAfter,
      description,
      referenceId
    );
    
    return {
      sourceTransaction,
      targetTransaction,
      sourceBalanceAfter,
      targetBalanceAfter
    };
  });
}

export interface ThreePartyTransferResult {
  sourceTransaction: Transaction;
  targetTransaction: Transaction;
  feeTransaction: Transaction;
  actualTransferAmount: number;
  feeAmount: number;
}

export function threePartyTransfer(
  sourceAccountId: string,
  targetAccountId: string,
  feeAccountId: string,
  totalAmount: number,
  feeRate: number,
  description: string = 'Transfer with fee',
  referenceId?: string
): ThreePartyTransferResult {
  if (totalAmount <= 0) {
    throw new Error('Amount must be positive');
  }
  
  if (feeRate < 0 || feeRate > 1) {
    throw new Error('Fee rate must be between 0 and 1');
  }
  
  return executeInTransaction(() => {
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit, status FROM accounts WHERE id = ?
    `).get(sourceAccountId) as any;
    
    if (!sourceRow) {
      throw new AccountNotFoundError('Source account not found');
    }
    
    const targetRow = db.prepare(`
      SELECT id, balance, overdraft_limit, status FROM accounts WHERE id = ?
    `).get(targetAccountId) as any;
    
    if (!targetRow) {
      throw new AccountNotFoundError('Target account not found');
    }
    
    const feeRow = db.prepare(`
      SELECT id, balance, overdraft_limit, status FROM accounts WHERE id = ?
    `).get(feeAccountId) as any;
    
    if (!feeRow) {
      throw new AccountNotFoundError('Fee account not found');
    }
    
    const calculatedFee = Math.floor(totalAmount * feeRate);
    const sourceBalanceBefore = sourceRow.balance;
    
    const sourceBalanceAfter = sourceBalanceBefore - totalAmount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError();
    }
    
    let actualTransferAmount: number;
    let actualFeeAmount: number;
    
    if (calculatedFee >= totalAmount) {
      actualTransferAmount = 0;
      actualFeeAmount = totalAmount;
    } else {
      actualTransferAmount = totalAmount - calculatedFee;
      actualFeeAmount = calculatedFee;
    }
    
    const targetBalanceBefore = targetRow.balance;
    const targetBalanceAfter = targetBalanceBefore + actualTransferAmount;
    
    const feeBalanceBefore = feeRow.balance;
    const feeBalanceAfter = feeBalanceBefore + actualFeeAmount;
    
    const now = Date.now();
    
    const updateResult = db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ? AND balance = ?
    `).run(sourceBalanceAfter, now, sourceAccountId, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(targetBalanceAfter, now, targetAccountId);
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(feeBalanceAfter, now, feeAccountId);
    
    const sourceTransaction = createTransaction(
      sourceAccountId,
      TransactionType.TRANSFER,
      -totalAmount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      description,
      referenceId
    );
    
    const targetTransaction = createTransaction(
      targetAccountId,
      TransactionType.TRANSFER,
      actualTransferAmount,
      targetBalanceBefore,
      targetBalanceAfter,
      description,
      referenceId
    );
    
    const feeTransaction = createTransaction(
      feeAccountId,
      TransactionType.FEE,
      actualFeeAmount,
      feeBalanceBefore,
      feeBalanceAfter,
      `Platform fee for ${description}`,
      referenceId
    );
    
    return {
      sourceTransaction,
      targetTransaction,
      feeTransaction,
      actualTransferAmount,
      feeAmount: actualFeeAmount
    };
  });
}
