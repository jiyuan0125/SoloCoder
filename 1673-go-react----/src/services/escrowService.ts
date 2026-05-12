import {
  getUser,
  getAccountByType,
  createEscrowTransaction,
  getEscrowTransaction,
  getEscrowTransactionByOrder,
  updateEscrowTransactionStatus,
  getOrCreatePlatformIncomeAccount,
  executeInTransaction,
  db,
  createTransaction
} from '../database';
import { AccountType, TransactionType, EscrowTransactionStatus, EscrowTransaction, Transaction } from '../types';
import { AccountNotFoundError, UserNotFoundError, InsufficientBalanceError } from './accountService';

export class DuplicateEscrowTransactionError extends Error {
  constructor(message: string = 'Escrow transaction already exists for this order') {
    super(message);
    this.name = 'DuplicateEscrowTransactionError';
  }
}

export class InvalidEscrowStatusError extends Error {
  constructor(message: string = 'Invalid escrow transaction status') {
    super(message);
    this.name = 'InvalidEscrowStatusError';
  }
}

const PLATFORM_FEE_RATE = 0.01;

interface EscrowPaymentResult {
  escrowTransaction: EscrowTransaction;
  sourceTransaction: Transaction;
  targetTransaction: Transaction;
}

export function payForOrder(
  buyerUserId: string,
  sellerUserId: string,
  orderId: string,
  amount: number
): EscrowPaymentResult {
  if (amount <= 0) {
    throw new Error('Amount must be positive');
  }
  
  if (!getUser(buyerUserId)) {
    throw new UserNotFoundError('Buyer user not found');
  }
  
  if (!getUser(sellerUserId)) {
    throw new UserNotFoundError('Seller user not found');
  }
  
  return executeInTransaction(() => {
    const existing = getEscrowTransactionByOrder(buyerUserId, orderId);
    if (existing) {
      throw new DuplicateEscrowTransactionError();
    }
    
    const buyerRechargeAccount = getAccountByType(buyerUserId, AccountType.RECHARGE);
    if (!buyerRechargeAccount) {
      throw new AccountNotFoundError('Buyer recharge account not found');
    }
    
    const buyerEscrowAccount = getAccountByType(buyerUserId, AccountType.ESCROW);
    if (!buyerEscrowAccount) {
      throw new AccountNotFoundError('Buyer escrow account not found');
    }
    
    const sellerRechargeAccount = getAccountByType(sellerUserId, AccountType.RECHARGE);
    if (!sellerRechargeAccount) {
      throw new AccountNotFoundError('Seller recharge account not found');
    }
    
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(buyerRechargeAccount.id) as any;
    
    const sourceBalanceBefore = sourceRow.balance;
    const sourceBalanceAfter = sourceBalanceBefore - amount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError();
    }
    
    const targetRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(buyerEscrowAccount.id) as any;
    
    const targetBalanceBefore = targetRow.balance;
    const targetBalanceAfter = targetBalanceBefore + amount;
    
    const now = Date.now();
    
    const updateResult = db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ? AND balance = ?
    `).run(sourceBalanceAfter, now, buyerRechargeAccount.id, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(targetBalanceAfter, now, buyerEscrowAccount.id);
    
    const sourceTransaction = createTransaction(
      buyerRechargeAccount.id,
      TransactionType.CONSUME,
      -amount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      `Payment for order ${orderId} to escrow`,
      orderId
    );
    
    const targetTransaction = createTransaction(
      buyerEscrowAccount.id,
      TransactionType.TRANSFER,
      amount,
      targetBalanceBefore,
      targetBalanceAfter,
      `Escrow hold for order ${orderId}`,
      orderId
    );
    
    const escrowTransaction = createEscrowTransaction({
      orderId,
      buyerUserId,
      sellerUserId,
      buyerEscrowAccountId: buyerEscrowAccount.id,
      sellerRechargeAccountId: sellerRechargeAccount.id,
      amount
    });
    
    updateEscrowTransactionStatus(escrowTransaction.id, EscrowTransactionStatus.PAID);
    
    const updatedEscrow = getEscrowTransaction(escrowTransaction.id)!;
    
    return {
      escrowTransaction: updatedEscrow,
      sourceTransaction,
      targetTransaction
    };
  });
}

export function confirmShipment(escrowTransactionId: string): EscrowTransaction {
  const escrow = getEscrowTransaction(escrowTransactionId);
  if (!escrow) {
    throw new Error('Escrow transaction not found');
  }
  
  if (escrow.status !== EscrowTransactionStatus.PAID) {
    throw new InvalidEscrowStatusError('Can only confirm shipment from paid status');
  }
  
  updateEscrowTransactionStatus(escrowTransactionId, EscrowTransactionStatus.SHIPPED);
  
  return getEscrowTransaction(escrowTransactionId)!;
}

interface ConfirmReceiptResult {
  escrowTransaction: EscrowTransaction;
  sourceTransaction: Transaction;
  targetTransaction: Transaction;
  feeTransaction: Transaction;
  feeAmount: number;
}

export function confirmReceipt(escrowTransactionId: string): ConfirmReceiptResult {
  const escrow = getEscrowTransaction(escrowTransactionId);
  if (!escrow) {
    throw new Error('Escrow transaction not found');
  }
  
  if (escrow.status !== EscrowTransactionStatus.SHIPPED) {
    throw new InvalidEscrowStatusError('Can only confirm receipt from shipped status');
  }
  
  return executeInTransaction(() => {
    const platformAccount = getOrCreatePlatformIncomeAccount();
    
    const totalAmount = escrow.amount;
    const feeRate = PLATFORM_FEE_RATE;
    const calculatedFee = Math.floor(totalAmount * feeRate);
    
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(escrow.buyer_escrow_account_id) as any;
    
    const sourceBalanceBefore = sourceRow.balance;
    const sourceBalanceAfter = sourceBalanceBefore - totalAmount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError();
    }
    
    const targetRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(escrow.seller_recharge_account_id) as any;
    
    const feeRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(platformAccount.id) as any;
    
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
    `).run(sourceBalanceAfter, now, escrow.buyer_escrow_account_id, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(targetBalanceAfter, now, escrow.seller_recharge_account_id);
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(feeBalanceAfter, now, platformAccount.id);
    
    const sourceTransaction = createTransaction(
      escrow.buyer_escrow_account_id,
      TransactionType.TRANSFER,
      -totalAmount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      `Release escrow for order ${escrow.order_id} to seller`,
      escrow.order_id
    );
    
    const targetTransaction = createTransaction(
      escrow.seller_recharge_account_id,
      TransactionType.TRANSFER,
      actualTransferAmount,
      targetBalanceBefore,
      targetBalanceAfter,
      `Receive payment for order ${escrow.order_id}`,
      escrow.order_id
    );
    
    const feeTransaction = createTransaction(
      platformAccount.id,
      TransactionType.FEE,
      actualFeeAmount,
      feeBalanceBefore,
      feeBalanceAfter,
      `Platform fee for order ${escrow.order_id}`,
      escrow.order_id
    );
    
    updateEscrowTransactionStatus(escrowTransactionId, EscrowTransactionStatus.COMPLETED, actualFeeAmount);
    
    const updatedEscrow = getEscrowTransaction(escrowTransactionId)!;
    
    return {
      escrowTransaction: updatedEscrow,
      sourceTransaction,
      targetTransaction,
      feeTransaction,
      feeAmount: actualFeeAmount
    };
  });
}

interface RefundResult {
  escrowTransaction: EscrowTransaction;
  sourceTransaction: Transaction;
  targetTransaction: Transaction;
}

export function refundEscrow(escrowTransactionId: string): RefundResult {
  const escrow = getEscrowTransaction(escrowTransactionId);
  if (!escrow) {
    throw new Error('Escrow transaction not found');
  }
  
  if (escrow.status !== EscrowTransactionStatus.PAID && escrow.status !== EscrowTransactionStatus.SHIPPED) {
    throw new InvalidEscrowStatusError('Can only refund from paid or shipped status');
  }
  
  return executeInTransaction(() => {
    const buyerRechargeAccount = getAccountByType(escrow.buyer_user_id, AccountType.RECHARGE);
    if (!buyerRechargeAccount) {
      throw new AccountNotFoundError('Buyer recharge account not found');
    }
    
    const amount = escrow.amount;
    
    const sourceRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(escrow.buyer_escrow_account_id) as any;
    
    const sourceBalanceBefore = sourceRow.balance;
    const sourceBalanceAfter = sourceBalanceBefore - amount;
    
    if (sourceBalanceAfter < -sourceRow.overdraft_limit) {
      throw new InsufficientBalanceError('Insufficient funds in escrow account for refund');
    }
    
    const targetRow = db.prepare(`
      SELECT id, balance, overdraft_limit FROM accounts WHERE id = ?
    `).get(buyerRechargeAccount.id) as any;
    
    const targetBalanceBefore = targetRow.balance;
    const targetBalanceAfter = targetBalanceBefore + amount;
    
    const now = Date.now();
    
    const updateResult = db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ? AND balance = ?
    `).run(sourceBalanceAfter, now, escrow.buyer_escrow_account_id, sourceBalanceBefore);
    
    if (updateResult.changes === 0) {
      throw new Error('Concurrent update detected, please retry');
    }
    
    db.prepare(`
      UPDATE accounts
      SET balance = ?, updated_at = ?
      WHERE id = ?
    `).run(targetBalanceAfter, now, buyerRechargeAccount.id);
    
    const sourceTransaction = createTransaction(
      escrow.buyer_escrow_account_id,
      TransactionType.REFUND,
      -amount,
      sourceBalanceBefore,
      sourceBalanceAfter,
      `Refund escrow for order ${escrow.order_id}`,
      escrow.order_id
    );
    
    const targetTransaction = createTransaction(
      buyerRechargeAccount.id,
      TransactionType.REFUND,
      amount,
      targetBalanceBefore,
      targetBalanceAfter,
      `Receive refund for order ${escrow.order_id}`,
      escrow.order_id
    );
    
    updateEscrowTransactionStatus(escrowTransactionId, EscrowTransactionStatus.REFUNDED);
    
    const updatedEscrow = getEscrowTransaction(escrowTransactionId)!;
    
    return {
      escrowTransaction: updatedEscrow,
      sourceTransaction,
      targetTransaction
    };
  });
}

export function getEscrowTransactionDetails(id: string): EscrowTransaction {
  const escrow = getEscrowTransaction(id);
  if (!escrow) {
    throw new Error('Escrow transaction not found');
  }
  return escrow;
}
