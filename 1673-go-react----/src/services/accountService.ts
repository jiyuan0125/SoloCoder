import {
  createUser,
  getUser,
  createAccount,
  getAccount,
  getAccountByType,
  getAccountsByUserId,
  accountExistsByType,
  updateAccountBalance,
  getTransactionsByAccount,
  executeInTransaction
} from '../database';
import { AccountType, TransactionType, Account, Transaction, User } from '../types';

export class AccountNotFoundError extends Error {
  constructor(message: string = 'Account not found') {
    super(message);
    this.name = 'AccountNotFoundError';
  }
}

export class UserNotFoundError extends Error {
  constructor(message: string = 'User not found') {
    super(message);
    this.name = 'UserNotFoundError';
  }
}

export class DuplicateAccountError extends Error {
  constructor(message: string = 'Account already exists') {
    super(message);
    this.name = 'DuplicateAccountError';
  }
}

export class InsufficientBalanceError extends Error {
  constructor(message: string = 'Insufficient balance') {
    super(message);
    this.name = 'InsufficientBalanceError';
  }
}

export function createPlatformUser(): User {
  const existing = getUser('platform');
  if (existing) {
    return existing;
  }
  
  const platformUser = createUser({ name: 'Platform' });
  createAccount({
    userId: platformUser.id,
    type: AccountType.PLATFORM_INCOME,
    overdraftLimit: 0
  });
  
  return platformUser;
}

export function createUserWithAccounts(name: string): { user: User; accounts: Account[] } {
  return executeInTransaction(() => {
    const user = createUser({ name });
    
    const mainAccount = createAccount({
      userId: user.id,
      type: AccountType.MAIN,
      overdraftLimit: 0
    });
    
    const rechargeAccount = createAccount({
      userId: user.id,
      type: AccountType.RECHARGE,
      overdraftLimit: 0,
      parentId: mainAccount.id
    });
    
    const escrowAccount = createAccount({
      userId: user.id,
      type: AccountType.ESCROW,
      overdraftLimit: 0,
      parentId: mainAccount.id
    });
    
    const withdrawAccount = createAccount({
      userId: user.id,
      type: AccountType.WITHDRAW,
      overdraftLimit: 0,
      parentId: mainAccount.id
    });
    
    return {
      user,
      accounts: [mainAccount, rechargeAccount, escrowAccount, withdrawAccount]
    };
  });
}

export function getOrCreateSubAccount(userId: string, type: AccountType): Account {
  if (!getUser(userId)) {
    throw new UserNotFoundError();
  }
  
  const existing = getAccountByType(userId, type);
  if (existing) {
    return existing;
  }
  
  let mainAccount = getAccountByType(userId, AccountType.MAIN);
  if (!mainAccount) {
    mainAccount = createAccount({
      userId,
      type: AccountType.MAIN,
      overdraftLimit: 0
    });
  }
  
  return createAccount({
    userId,
    type,
    overdraftLimit: 0,
    parentId: mainAccount.id
  });
}

export function openAccount(userId: string, type: AccountType): Account {
  if (!getUser(userId)) {
    throw new UserNotFoundError();
  }
  
  if (accountExistsByType(userId, type)) {
    throw new DuplicateAccountError();
  }
  
  let mainAccount = getAccountByType(userId, AccountType.MAIN);
  if (!mainAccount && type !== AccountType.MAIN) {
    mainAccount = createAccount({
      userId,
      type: AccountType.MAIN,
      overdraftLimit: 0
    });
  }
  
  return createAccount({
    userId,
    type,
    overdraftLimit: 0,
    parentId: type === AccountType.MAIN ? undefined : mainAccount?.id
  });
}

export function getUserAccounts(userId: string): Account[] {
  if (!getUser(userId)) {
    throw new UserNotFoundError();
  }
  
  return getAccountsByUserId(userId);
}

export function getAccountDetails(accountId: string): Account {
  const account = getAccount(accountId);
  if (!account) {
    throw new AccountNotFoundError();
  }
  return account;
}

export function rechargeAccount(accountId: string, amount: number): { account: Account; transaction: Transaction } {
  if (amount <= 0) {
    throw new Error('Amount must be positive');
  }
  
  const account = getAccount(accountId);
  if (!account) {
    throw new AccountNotFoundError();
  }
  
  return updateAccountBalance({
    accountId,
    amount,
    transactionType: TransactionType.RECHARGE,
    description: `Recharge ${amount} cents`
  });
}

export function getAccountTransactions(accountId: string): Transaction[] {
  const account = getAccount(accountId);
  if (!account) {
    throw new AccountNotFoundError();
  }
  return getTransactionsByAccount(accountId);
}

export function verifyAccountBalance(accountId: string): boolean {
  const account = getAccount(accountId);
  if (!account) {
    throw new AccountNotFoundError();
  }
  
  const transactions = getTransactionsByAccount(accountId);
  let calculatedBalance = 0;
  
  for (let i = transactions.length - 1; i >= 0; i--) {
    calculatedBalance += transactions[i].amount;
  }
  
  return calculatedBalance === account.balance;
}
