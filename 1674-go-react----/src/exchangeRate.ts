import { db } from './database';
import { Currency, ExchangeRate, SUPPORTED_CURRENCIES, ExchangeRateInfo } from './types';

const RATE_FRESHNESS_MS = 60 * 1000; // 1 分钟
const RATE_SCALE = 1000000;

type RateUpdateCallback = () => void;

let rateUpdateCallbacks: RateUpdateCallback[] = [];

export const registerRateUpdateCallback = (callback: RateUpdateCallback): void => {
  rateUpdateCallbacks.push(callback);
};

export const unregisterRateUpdateCallback = (callback: RateUpdateCallback): void => {
  rateUpdateCallbacks = rateUpdateCallbacks.filter(cb => cb !== callback);
};

const notifyRateUpdated = (): void => {
  for (const callback of rateUpdateCallbacks) {
    callback();
  }
};

export const getExchangeRate = (currency: Currency): Promise<ExchangeRate | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM exchange_rates WHERE currency = ?', [currency], (err, row) => {
      if (err) {
        reject(err);
        return;
      }
      if (row) {
        resolve(row as ExchangeRate);
      } else {
        resolve(null);
      }
    });
  });
};

export const getAllExchangeRates = (): Promise<ExchangeRate[]> => {
  return new Promise((resolve, reject) => {
    db.all('SELECT * FROM exchange_rates', (err, rows) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(rows as ExchangeRate[]);
    });
  });
};

export const isRateFresh = (rate: ExchangeRate): boolean => {
  const now = Date.now();
  return (now - rate.updated_at) <= RATE_FRESHNESS_MS;
};

export const areAllRatesFresh = async (): Promise<boolean> => {
  const rates = await getAllExchangeRates();
  if (rates.length === 0) return false;
  
  for (const rate of rates) {
    if (!isRateFresh(rate)) {
      return false;
    }
  }
  return true;
};

export const updateExchangeRate = async (currency: Currency, rate: number): Promise<ExchangeRate> => {
  const now = Date.now();
  
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM exchange_rates WHERE currency = ?', [currency], (err, row) => {
      if (err) {
        reject(err);
        return;
      }
      
      if (row) {
        db.run(
          'UPDATE exchange_rates SET rate = ?, updated_at = ? WHERE currency = ?',
          [rate, now, currency],
          (updateErr) => {
            if (updateErr) {
              reject(updateErr);
              return;
            }
            const updatedRate: ExchangeRate = { currency, rate, updated_at: now };
            notifyRateUpdated();
            resolve(updatedRate);
          }
        );
      } else {
        db.run(
          'INSERT INTO exchange_rates (currency, rate, updated_at) VALUES (?, ?, ?)',
          [currency, rate, now],
          (insertErr) => {
            if (insertErr) {
              reject(insertErr);
              return;
            }
            const newRate: ExchangeRate = { currency, rate, updated_at: now };
            notifyRateUpdated();
            resolve(newRate);
          }
        );
      }
    });
  });
};

export const convertCurrency = async (
  amount: number,
  sourceCurrency: Currency,
  targetCurrency: Currency
): Promise<{ amount: number; rateInfo: ExchangeRateInfo }> => {
  const sourceRate = await getExchangeRate(sourceCurrency);
  const targetRate = await getExchangeRate(targetCurrency);
  
  if (!sourceRate || !targetRate) {
    throw new Error('汇率数据不存在');
  }
  
  if (!isRateFresh(sourceRate) || !isRateFresh(targetRate)) {
    throw new Error('汇率过期');
  }
  
  const cnyAmount = Math.floor((amount * sourceRate.rate) / RATE_SCALE);
  const targetAmount = Math.floor((cnyAmount * RATE_SCALE) / targetRate.rate);
  
  return {
    amount: targetAmount,
    rateInfo: {
      sourceRate: sourceRate.rate,
      targetRate: targetRate.rate,
      timestamp: Date.now()
    }
  };
};

export const simulateRatePush = (): void => {
  const rates: Record<Currency, number> = {
    [Currency.CNY]: RATE_SCALE,
    [Currency.USD]: 7100000,
    [Currency.EUR]: 7700000,
    [Currency.JPY]: 50000
  };
  
  const jitter = (rate: number): number => {
    const variation = Math.floor(rate * 0.005);
    return rate + Math.floor(Math.random() * variation * 2) - variation;
  };
  
  let variation = 1;
  setInterval(async () => {
    for (const currency of SUPPORTED_CURRENCIES) {
      let newRate = rates[currency];
      if (currency !== Currency.CNY) {
        newRate = jitter(rates[currency]);
      }
      await updateExchangeRate(currency, newRate);
    }
    console.log(`[${new Date().toISOString()}] 汇率已更新（变化因子: ${variation}）`);
    variation += 1;
  }, RATE_FRESHNESS_MS);
};
