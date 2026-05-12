import {
  getUser,
  getAccountByType,
  createWithdrawRecord,
  executeInTransaction,
  db,
  createTransaction
} from '../database';
import { AccountType, TransactionType, WithdrawRecord, Transaction } from '../types';
import { AccountNotFoundError, UserNotFoundError, InsufficientBalanceError } from './accountService';

interface WithdrawResult {
  withdrawRecord: WithdrawRecord;
  transaction: Transaction;
}

export function withdraw(
  userId: string,
  bankAccount: string,
  amount: number
): WithdrawResult {
  if (amount <= 0) {
    throw new Error('Amount must be positive');
  }
  
  if (!bankAccount || bankAccount.trim().length === 0) {
    throw new Error('Bank account is required');
  }
  
  if (!getUser(userId)) {
    throw new UserNotFoundError();
  }
  
  return executeInTransaction(() => {
    const rechargeAccount = getAccountByType(userId, AccountType.RECHARGE);
    if (!rechargeAccount) {
      throw new AccountNotFoundError('Recharge account not found');
    }
    
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(rechargeAccount.id) as any;
    
    const sourceBalanceBefore = sourceRow.balance;
    const sourceBalanceAfter = sourceBalanceBefore - amount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError();
    }
    
    const now = Date.now();
    
    const updateResult = db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ? AND balance = ?
    `).run(sourceBalanceAfter, now, rechargeAccount.id, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    const transaction = createTransaction(
      rechargeAccount.id,
      TransactionType.WITHDRAW,
      -amount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      `Withdraw ${amount} cents to bank account ${bankAccount}`,
      undefined
    );
    
    const withdrawRecord = createWithdrawRecord({
      userId,
      accountId: rechargeAccount.id,
      bankAccount: bankAccount.trim(),
      amount
    });
    
    return {
      withdrawRecord,
      transaction
    };
  });
}
