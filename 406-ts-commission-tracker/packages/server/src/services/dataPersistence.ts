import * as fs from 'fs';
import * as path from 'path';
import {
  Salesperson,
  Order,
  Settlement,
} from '@commission-tracker/shared';

interface PersistedData {
  salespeople: Salesperson[];
  orders: Order[];
  settlements: Settlement[];
}

const DATA_DIR = path.join(process.cwd(), 'data');
const DATA_FILE = path.join(DATA_DIR, 'commission-data.json');

function ensureDataDirectory(): void {
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
}

export function loadData(): PersistedData {
  ensureDataDirectory();
  
  if (!fs.existsSync(DATA_FILE)) {
    return {
      salespeople: [],
      orders: [],
      settlements: [],
    };
  }

  try {
    const content = fs.readFileSync(DATA_FILE, 'utf-8');
    return JSON.parse(content) as PersistedData;
  } catch {
    return {
      salespeople: [],
      orders: [],
      settlements: [],
    };
  }
}

export function saveData(data: PersistedData): void {
  ensureDataDirectory();
  
  const content = JSON.stringify(data, null, 2);
  fs.writeFileSync(DATA_FILE, content, 'utf-8');
}

export class DataPersistenceService {
  private data: PersistedData;

  constructor() {
    this.data = loadData();
  }

  getSalespeople(): Salesperson[] {
    return [...this.data.salespeople];
  }

  getSalesperson(id: string): Salesperson | undefined {
    return this.data.salespeople.find(s => s.id === id);
  }

  saveSalesperson(salesperson: Salesperson): void {
    const index = this.data.salespeople.findIndex(s => s.id === salesperson.id);
    if (index >= 0) {
      this.data.salespeople[index] = salesperson;
    } else {
      this.data.salespeople.push(salesperson);
    }
    this.persist();
  }

  getOrders(): Order[] {
    return [...this.data.orders];
  }

  getOrder(id: string): Order | undefined {
    return this.data.orders.find(o => o.id === id);
  }

  saveOrder(order: Order): void {
    const index = this.data.orders.findIndex(o => o.id === order.id);
    if (index >= 0) {
      this.data.orders[index] = order;
    } else {
      this.data.orders.push(order);
    }
    this.persist();
  }

  getSettlements(): Settlement[] {
    return [...this.data.settlements];
  }

  getSettlement(id: string): Settlement | undefined {
    return this.data.settlements.find(s => s.id === id);
  }

  saveSettlement(settlement: Settlement): void {
    const index = this.data.settlements.findIndex(s => s.id === settlement.id);
    if (index >= 0) {
      this.data.settlements[index] = settlement;
    } else {
      this.data.settlements.push(settlement);
    }
    this.persist();
  }

  private persist(): void {
    saveData(this.data);
  }
}
