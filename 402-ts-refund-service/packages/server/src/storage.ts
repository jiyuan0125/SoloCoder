import fs from 'fs';
import path from 'path';
import { Order, Refund, Config } from '@refund/shared';
import { getConfig } from './config.js';

interface StorageData {
  orders: Order[];
  refunds: Refund[];
  config: Config;
}

const ensureDataDirectory = (): void => {
  const config = getConfig();
  if (!fs.existsSync(config.dataStoragePath)) {
    fs.mkdirSync(config.dataStoragePath, { recursive: true });
  }
};

const getDataFilePath = (): string => {
  const config = getConfig();
  return path.join(config.dataStoragePath, 'refund-data.json');
};

export function loadData(): StorageData {
  const config = getConfig();
  const defaultData: StorageData = {
    orders: [],
    refunds: [],
    config: { ...config }
  };

  try {
    ensureDataDirectory();
    const filePath = getDataFilePath();
    if (fs.existsSync(filePath)) {
      const rawData = fs.readFileSync(filePath, 'utf-8');
      const data = JSON.parse(rawData) as StorageData;
      
      data.orders = data.orders.map(order => ({
        ...order,
        orderTime: new Date(order.orderTime)
      }));

      data.refunds = data.refunds.map(refund => ({
        ...refund,
        createdAt: new Date(refund.createdAt),
        updatedAt: new Date(refund.updatedAt),
        statusHistory: refund.statusHistory.map(item => ({
          ...item,
          time: new Date(item.time)
        }))
      }));

      return data;
    }
  } catch (error) {
    console.error('Error loading data:', error);
  }

  return defaultData;
}

export function saveData(data: StorageData): void {
  try {
    ensureDataDirectory();
    const filePath = getDataFilePath();
    
    const dataToSave = {
      orders: data.orders,
      refunds: data.refunds,
      config: data.config
    };

    fs.writeFileSync(filePath, JSON.stringify(dataToSave, null, 2), 'utf-8');
  } catch (error) {
    console.error('Error saving data:', error);
    throw error;
  }
}
