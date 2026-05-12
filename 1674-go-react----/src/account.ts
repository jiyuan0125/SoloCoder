import { db } from './database';
import { Account, Currency } from './types';

const generateId = (): string => {
  return 'acc_' + Math.random().toString(36).substring(2, 15);
};

export const createAccount = async (currency: Currency): Promise<Account> => {
  const id = generateId();
  const now = Date.now();
  
  return new Promise((resolve, reject) => {
    db.run(
      'INSERT INTO accounts (id, currency, balance, created_at) VALUES (?, ?, ?, ?)',
      [id, currency, 0, now],
      (err) => {
        if (err) {
          reject(err);
          return;
        }
        resolve({
          id,
          currency,
          balance: 0,
          created_at: now
        });
      }
    );
  });
};

export const getAccount = async (id: string): Promise<Account | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM accounts WHERE id = ?', [id], (err, row) => {
      if (err) {
        reject(err);
        return;
      }
      if (row) {
        resolve(row as Account);
      } else {
        resolve(null);
      }
    });
  });
};

export const getAccountBalance = async (id: string): Promise<number | null> => {
  const account = await getAccount(id);
  return account ? account.balance : null;
};

export const creditAccount = async (id: string, amount: number): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.run(
      'UPDATE accounts SET balance = balance + ? WHERE id = ?',
      [amount, id],
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

export const debitAccount = async (id: string, amount: number): Promise<boolean> => {
  return new Promise((resolve, reject) => {
    db.run(
      'UPDATE accounts SET balance = balance - ? WHERE id = ? AND balance >= ?',
      [amount, id, amount],
      function(err) {
        if (err) {
          reject(err);
          return;
        }
        resolve(this.changes > 0);
      }
    );
  });
};
