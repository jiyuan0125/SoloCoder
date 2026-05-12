import Database from 'better-sqlite3';
import path from 'path';
import { v4 as uuidv4 } from 'uuid';
import {
  AccountType,
  AccountStatus,
  TransactionType,
  EscrowTransactionStatus,
  Account,
  Transaction,
  EscrowTransaction,
  User,
  WithdrawRecord
} from './types';

const DB_PATH = path.join(process.cwd(), 'virtual_account.db');

const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS accounts (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      type TEXT NOT NULL,
      balance INTEGER NOT NULL DEFAULT 0,
      overdraft_limit INTEGER NOT NULL DEFAULT 0,
      status TEXT NOT NULL DEFAULT 'active',
      parent_id TEXT,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      FOREIGN KEY (user_id) REFERENCES users(id),
      FOREIGN KEY (parent_id) REFERENCES accounts(id),
      UNIQUE(user_id, type)
    );

    CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
    CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);

    CREATE TABLE IF NOT EXISTS transactions (
      id TEXT PRIMARY KEY,
      account_id TEXT NOT NULL,
      type TEXT NOT NULL,
      amount INTEGER NOT NULL,
      balance_before INTEGER NOT NULL,
      balance_after INTEGER NOT NULL,
      description TEXT,
      reference_id TEXT,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (account_id) REFERENCES accounts(id)
    );

    CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
    CREATE INDEX IF NOT EXISTS idx_transactions_reference_id ON transactions(reference_id);

    CREATE TABLE IF NOT EXISTS escrow_transactions (
      id TEXT PRIMARY KEY,
      order_id TEXT NOT NULL,
      buyer_user_id TEXT NOT NULL,
      seller_user_id TEXT NOT NULL,
      buyer_escrow_account_id TEXT NOT NULL,
      seller_recharge_account_id TEXT NOT NULL,
      amount INTEGER NOT NULL,
      platform_fee INTEGER NOT NULL DEFAULT 0,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      UNIQUE(buyer_user_id, order_id)
    );

    CREATE INDEX IF NOT EXISTS idx_escrow_transactions_order_id ON escrow_transactions(order_id);
    CREATE INDEX IF NOT EXISTS idx_escrow_transactions_buyer ON escrow_transactions(buyer_user_id);
    CREATE INDEX IF NOT EXISTS idx_escrow_transactions_seller ON escrow_transactions(seller_user_id);

    CREATE TABLE IF NOT EXISTS withdraw_records (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      account_id TEXT NOT NULL,
      bank_account TEXT NOT NULL,
      amount INTEGER NOT NULL,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at INTEGER NOT NULL,
      FOREIGN KEY (user_id) REFERENCES users(id),
      FOREIGN KEY (account_id) REFERENCES accounts(id)
    );

    CREATE INDEX IF NOT EXISTS idx_withdraw_records_user_id ON withdraw_records(user_id);
  `);
}

initDatabase();

interface CreateUserInput {
  name: string;
}

interface CreateAccountInput {
  userId: string;
  type: AccountType;
  overdraftLimit?: number;
  parentId?: string;
}

interface UpdateAccountBalanceInput {
  accountId: string;
  amount: number;
  transactionType: TransactionType;
  description: string;
  referenceId?: string;
}

function createUser(input: CreateUserInput): User {
  const id = uuidv4();
  const now = Date.now();
  
  db.prepare(`
    INSERT INTO users (id, name, created_at)
    VALUES (?, ?, ?)
  `).run(id, input.name, now);

  return {
    id,
    name: input.name,
    created_at: now
  };
}

function getUser(userId: string): User | undefined {
  return db.prepare(`
    SELECT id, name, created_at FROM users WHERE id = ?
  `).get(userId) as User | undefined;
}

function createAccount(input: CreateAccountInput): Account {
  const id = uuidv4();
  const now = Date.now();
  const overdraftLimit = input.overdraftLimit ?? 0;

  db.prepare(`
    INSERT INTO accounts (id, user_id, type, balance, overdraft_limit, status, parent_id, created_at, updated_at)
    VALUES (?, ?, ?, 0, ?, 'active', ?, ?, ?)
  `).run(id, input.userId, input.type, overdraftLimit, input.parentId || null, now, now);

  return getAccount(id)!;
}

function getAccount(accountId: string): Account | undefined {
  return db.prepare(`
    SELECT id, user_id, type, balance, overdraft_limit, status, parent_id, created_at, updated_at
    FROM accounts WHERE id = ?
  `).get(accountId) as Account | undefined;
}

function getAccountByType(userId: string, type: AccountType): Account | undefined {
  return db.prepare(`
    SELECT id, user_id, type, balance, overdraft_limit, status, parent_id, created_at, updated_at
    FROM accounts WHERE user_id = ? AND type = ?
  `).get(userId, type) as Account | undefined;
}

function getAccountsByUserId(userId: string): Account[] {
  return db.prepare(`
    SELECT id, user_id, type, balance, overdraft_limit, status, parent_id, created_at, updated_at
    FROM accounts WHERE user_id = ?
  `).all(userId) as Account[];
}

function accountExistsByType(userId: string, type: AccountType): boolean {
  const result = db.prepare(`
    SELECT 1 FROM accounts WHERE user_id = ? AND type = ?
  `).get(userId, type);
  return !!result;
}

function updateAccountBalance(input: UpdateAccountBalanceInput): { account: Account; transaction: Transaction } {
  const account = getAccount(input.accountId);
  if (!account) {
    throw new Error(`Account not found: ${input.accountId}`);
  }

  const balanceBefore = account.balance;
  const balanceAfter = balanceBefore + input.amount;

  if (balanceAfter < -account.overdraft_limit) {
    throw new Error('Insufficient balance or exceeds overdraft limit');
  }

  const now = Date.now();
  const transactionId = uuidv4();

  db.prepare(`
    UPDATE accounts
    SET balance = ?, updated_at = ?
    WHERE id = ?
  `).run(balanceAfter, now, input.accountId);

  db.prepare(`
    INSERT INTO transactions (id, account_id, type, amount, balance_before, balance_after, description, reference_id, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    transactionId,
    input.accountId,
    input.transactionType,
    input.amount,
    balanceBefore,
    balanceAfter,
    input.description,
    input.referenceId || null,
    now
  );

  const updatedAccount = getAccount(input.accountId)!;
  const transaction = db.prepare(`
    SELECT id, account_id, type, amount, balance_before, balance_after, description, reference_id, created_at
    FROM transactions WHERE id = ?
  `).get(transactionId) as Transaction;

  return { account: updatedAccount, transaction };
}

function createTransaction(
  accountId: string,
  type: TransactionType,
  amount: number,
  balanceBefore: number,
  balanceAfter: number,
  description: string,
  referenceId?: string
): Transaction {
  const id = uuidv4();
  const now = Date.now();

  db.prepare(`
    INSERT INTO transactions (id, account_id, type, amount, balance_before, balance_after, description, reference_id, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, accountId, type, amount, balanceBefore, balanceAfter, description, referenceId || null, now);

  return db.prepare(`
    SELECT id, account_id, type, amount, balance_before, balance_after, description, reference_id, created_at
    FROM transactions WHERE id = ?
  `).get(id) as Transaction;
}

function getTransactionsByAccount(accountId: string): Transaction[] {
  return db.prepare(`
    SELECT id, account_id, type, amount, balance_before, balance_after, description, reference_id, created_at
    FROM transactions WHERE account_id = ?
    ORDER BY created_at DESC
  `).all(accountId) as Transaction[];
}

function createEscrowTransaction(input: {
  orderId: string;
  buyerUserId: string;
  sellerUserId: string;
  buyerEscrowAccountId: string;
  sellerRechargeAccountId: string;
  amount: number;
}): EscrowTransaction {
  const id = uuidv4();
  const now = Date.now();

  db.prepare(`
    INSERT INTO escrow_transactions (id, order_id, buyer_user_id, seller_user_id, buyer_escrow_account_id, seller_recharge_account_id, amount, platform_fee, status, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, 0, 'pending', ?, ?)
  `).run(
    id,
    input.orderId,
    input.buyerUserId,
    input.sellerUserId,
    input.buyerEscrowAccountId,
    input.sellerRechargeAccountId,
    input.amount,
    now,
    now
  );

  return getEscrowTransaction(id)!;
}

function getEscrowTransaction(id: string): EscrowTransaction | undefined {
  return db.prepare(`
    SELECT id, order_id, buyer_user_id, seller_user_id, buyer_escrow_account_id, seller_recharge_account_id, amount, platform_fee, status, created_at, updated_at
    FROM escrow_transactions WHERE id = ?
  `).get(id) as EscrowTransaction | undefined;
}

function getEscrowTransactionByOrder(buyerUserId: string, orderId: string): EscrowTransaction | undefined {
  return db.prepare(`
    SELECT id, order_id, buyer_user_id, seller_user_id, buyer_escrow_account_id, seller_recharge_account_id, amount, platform_fee, status, created_at, updated_at
    FROM escrow_transactions WHERE buyer_user_id = ? AND order_id = ?
  `).get(buyerUserId, orderId) as EscrowTransaction | undefined;
}

function updateEscrowTransactionStatus(
  id: string,
  status: EscrowTransactionStatus,
  platformFee?: number
): void {
  const now = Date.now();
  if (platformFee !== undefined) {
    db.prepare(`
      UPDATE escrow_transactions
      SET status = ?, platform_fee = ?, updated_at = ?
      WHERE id = ?
    `).run(status, platformFee, now, id);
  } else {
    db.prepare(`
      UPDATE escrow_transactions
      SET status = ?, updated_at = ?
      WHERE id = ?
    `).run(status, now, id);
  }
}

function createWithdrawRecord(input: {
  userId: string;
  accountId: string;
  bankAccount: string;
  amount: number;
}): WithdrawRecord {
  const id = uuidv4();
  const now = Date.now();

  db.prepare(`
    INSERT INTO withdraw_records (id, user_id, account_id, bank_account, amount, status, created_at)
    VALUES (?, ?, ?, ?, ?, 'pending', ?)
  `).run(id, input.userId, input.accountId, input.bankAccount, input.amount, now);

  return db.prepare(`
    SELECT id, user_id, account_id, bank_account, amount, status, created_at
    FROM withdraw_records WHERE id = ?
  `).get(id) as WithdrawRecord;
}

function getOrCreatePlatformIncomeAccount(): Account {
  let account = getAccountByType('platform', AccountType.PLATFORM_INCOME);
  if (!account) {
    const platformUser = getUser('platform');
    if (!platformUser) {
      db.prepare(`
        INSERT INTO users (id, name, created_at)
        VALUES (?, ?, ?)
      `).run('platform', 'Platform', Date.now());
    }
    account = createAccount({
      userId: 'platform',
      type: AccountType.PLATFORM_INCOME,
      overdraftLimit: 0
    });
  }
  return account;
}

function executeInTransaction<T>(fn: () => T): T {
  const transaction = db.transaction(fn);
  return transaction();
}

export {
  db,
  createUser,
  getUser,
  createAccount,
  getAccount,
  getAccountByType,
  getAccountsByUserId,
  accountExistsByType,
  updateAccountBalance,
  createTransaction,
  getTransactionsByAccount,
  createEscrowTransaction,
  getEscrowTransaction,
  getEscrowTransactionByOrder,
  updateEscrowTransactionStatus,
  createWithdrawRecord,
  getOrCreatePlatformIncomeAccount,
  executeInTransaction
};
