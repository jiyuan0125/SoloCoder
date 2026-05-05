import { Config } from '@refund/shared';

export const PORT = 3000;

export const defaultConfig: Config = {
  refundPeriodDays: 30,
  virtualProductRefundThreshold: 60,
  dataStoragePath: './data'
};

let currentConfig: Config = { ...defaultConfig };

export function getConfig(): Config {
  return { ...currentConfig };
}

export function updateConfig(updates: Partial<Config>): Config {
  currentConfig = { ...currentConfig, ...updates };
  return { ...currentConfig };
}
