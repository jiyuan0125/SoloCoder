import { Request, Response } from 'express';
import {
  createUserWithAccounts,
  getUserAccounts,
  openAccount,
  rechargeAccount,
  getAccountDetails,
  getAccountTransactions,
  verifyAccountBalance,
  UserNotFoundError,
  AccountNotFoundError,
  DuplicateAccountError
} from '../services/accountService';
import { transfer } from '../services/transferService';
import { withdraw } from '../services/withdrawService';
import {
  payForOrder,
  confirmShipment,
  confirmReceipt,
  refundEscrow,
  getEscrowTransactionDetails,
  DuplicateEscrowTransactionError,
  InvalidEscrowStatusError
} from '../services/escrowService';
import { AccountType } from '../types';
import { InsufficientBalanceError } from '../services/accountService';

function handleError(res: Response, error: unknown): void {
  if (error instanceof UserNotFoundError) {
    res.status(404).json({ error: error.message });
    return;
  }
  if (error instanceof AccountNotFoundError) {
    res.status(404).json({ error: error.message });
    return;
  }
  if (error instanceof DuplicateAccountError) {
    res.status(409).json({ error: error.message });
    return;
  }
  if (error instanceof DuplicateEscrowTransactionError) {
    res.status(409).json({ error: error.message });
    return;
  }
  if (error instanceof InsufficientBalanceError) {
    res.status(400).json({ error: error.message });
    return;
  }
  if (error instanceof InvalidEscrowStatusError) {
    res.status(400).json({ error: error.message });
    return;
  }
  if (error instanceof Error) {
    res.status(400).json({ error: error.message });
    return;
  }
  res.status(500).json({ error: 'Internal server error' });
}

export function createUser(req: Request, res: Response): void {
  try {
    const { name } = req.body;
    if (!name || typeof name !== 'string') {
      res.status(400).json({ error: 'Name is required' });
      return;
    }
    const result = createUserWithAccounts(name);
    res.status(201).json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function listUserAccounts(req: Request, res: Response): void {
  try {
    const { userId } = req.params;
    const accounts = getUserAccounts(userId);
    res.json(accounts);
  } catch (error) {
    handleError(res, error);
  }
}

export function createSubAccount(req: Request, res: Response): void {
  try {
    const { userId } = req.params;
    const { type } = req.body;
    
    if (!type) {
      res.status(400).json({ error: 'Account type is required' });
      return;
    }
    
    const validTypes = [AccountType.RECHARGE, AccountType.ESCROW, AccountType.WITHDRAW];
    if (!validTypes.includes(type as AccountType)) {
      res.status(400).json({ error: 'Invalid account type' });
      return;
    }
    
    const account = openAccount(userId, type as AccountType);
    res.status(201).json(account);
  } catch (error) {
    handleError(res, error);
  }
}

export function getAccount(req: Request, res: Response): void {
  try {
    const { accountId } = req.params;
    const account = getAccountDetails(accountId);
    res.json(account);
  } catch (error) {
    handleError(res, error);
  }
}

export function recharge(req: Request, res: Response): void {
  try {
    const { accountId } = req.params;
    const { amount } = req.body;
    
    if (!amount || typeof amount !== 'number' || amount <= 0) {
      res.status(400).json({ error: 'Valid amount is required' });
      return;
    }
    
    const result = rechargeAccount(accountId, amount);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function transferAccounts(req: Request, res: Response): void {
  try {
    const { sourceAccountId, targetAccountId, amount, description } = req.body;
    
    if (!sourceAccountId || !targetAccountId) {
      res.status(400).json({ error: 'Source and target account IDs are required' });
      return;
    }
    
    if (!amount || typeof amount !== 'number' || amount <= 0) {
      res.status(400).json({ error: 'Valid amount is required' });
      return;
    }
    
    const result = transfer(
      sourceAccountId,
      targetAccountId,
      amount,
      description || 'Transfer'
    );
    
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function listAccountTransactions(req: Request, res: Response): void {
  try {
    const { accountId } = req.params;
    const transactions = getAccountTransactions(accountId);
    res.json(transactions);
  } catch (error) {
    handleError(res, error);
  }
}

export function verifyBalance(req: Request, res: Response): void {
  try {
    const { accountId } = req.params;
    const isValid = verifyAccountBalance(accountId);
    res.json({ valid: isValid });
  } catch (error) {
    handleError(res, error);
  }
}

export function createWithdraw(req: Request, res: Response): void {
  try {
    const { userId } = req.params;
    const { bankAccount, amount } = req.body;
    
    if (!bankAccount || typeof bankAccount !== 'string') {
      res.status(400).json({ error: 'Bank account is required' });
      return;
    }
    
    if (!amount || typeof amount !== 'number' || amount <= 0) {
      res.status(400).json({ error: 'Valid amount is required' });
      return;
    }
    
    const result = withdraw(userId, bankAccount, amount);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function escrowPay(req: Request, res: Response): void {
  try {
    const { buyerUserId, sellerUserId, orderId, amount } = req.body;
    
    if (!buyerUserId || !sellerUserId || !orderId) {
      res.status(400).json({ error: 'Buyer user ID, seller user ID, and order ID are required' });
      return;
    }
    
    if (!amount || typeof amount !== 'number' || amount <= 0) {
      res.status(400).json({ error: 'Valid amount is required' });
      return;
    }
    
    const result = payForOrder(buyerUserId, sellerUserId, orderId, amount);
    res.status(201).json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function escrowConfirmShipment(req: Request, res: Response): void {
  try {
    const { escrowId } = req.params;
    const result = confirmShipment(escrowId);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function escrowConfirmReceipt(req: Request, res: Response): void {
  try {
    const { escrowId } = req.params;
    const result = confirmReceipt(escrowId);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function escrowRefund(req: Request, res: Response): void {
  try {
    const { escrowId } = req.params;
    const result = refundEscrow(escrowId);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}

export function getEscrowTransaction(req: Request, res: Response): void {
  try {
    const { escrowId } = req.params;
    const result = getEscrowTransactionDetails(escrowId);
    res.json(result);
  } catch (error) {
    handleError(res, error);
  }
}
