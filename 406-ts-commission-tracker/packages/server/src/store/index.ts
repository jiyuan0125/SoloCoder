import {
  SalespersonId,
  OrderId,
  SettlementId,
  Salesperson,
  Order,
  Settlement,
  Month,
  Year,
  SalesContribution,
} from '@commission-tracker/shared';
import { DataPersistenceService } from '../services/dataPersistence';

const persistence = new DataPersistenceService();

export class InMemoryStore {
  private salespeople: Map<SalespersonId, Salesperson>;
  private orders: Map<OrderId, Order>;
  private settlements: Map<SettlementId, Settlement>;

  constructor() {
    this.salespeople = new Map();
    this.orders = new Map();
    this.settlements = new Map();
    this.loadFromPersistence();
  }

  private loadFromPersistence(): void {
    for (const sp of persistence.getSalespeople()) {
      this.salespeople.set(sp.id, sp);
    }
    for (const order of persistence.getOrders()) {
      this.orders.set(order.id, order);
    }
    for (const settlement of persistence.getSettlements()) {
      this.settlements.set(settlement.id, settlement);
    }
  }

  getSalesperson(id: SalespersonId): Salesperson | undefined {
    return this.salespeople.get(id);
  }

  getSalespeople(): Salesperson[] {
    return Array.from(this.salespeople.values());
  }

  saveSalesperson(salesperson: Salesperson): void {
    this.salespeople.set(salesperson.id, salesperson);
    persistence.saveSalesperson(salesperson);
  }

  getOrder(id: OrderId): Order | undefined {
    return this.orders.get(id);
  }

  getOrders(): Order[] {
    return Array.from(this.orders.values());
  }

  saveOrder(order: Order): void {
    this.orders.set(order.id, order);
    persistence.saveOrder(order);
  }

  getSettlement(id: SettlementId): Settlement | undefined {
    return this.settlements.get(id);
  }

  getSettlements(): Settlement[] {
    return Array.from(this.settlements.values());
  }

  saveSettlement(settlement: Settlement): void {
    this.settlements.set(settlement.id, settlement);
    persistence.saveSettlement(settlement);
  }

  getSettlementsForSalesperson(salespersonId: SalespersonId): Settlement[] {
    return this.getSettlements().filter(s => s.salespersonId === salespersonId);
  }

  getSettlementsForMonth(month: Month, year: Year): Settlement[] {
    return this.getSettlements().filter(s => s.month === month && s.year === year);
  }

  getOrdersForSalesperson(salespersonId: SalespersonId): Order[] {
    return this.getOrders().filter(o =>
      o.salesContributions.some((c: SalesContribution) => c.salespersonId === salespersonId)
    );
  }

  getOrdersForMonth(month: Month, year: Year): Order[] {
    const monthStr = `${year}-${String(month).padStart(2, '0')}`;
    return this.getOrders().filter(o => o.date.startsWith(monthStr));
  }

  lockOrder(id: OrderId): void {
    const order = this.getOrder(id);
    if (order) {
      order.locked = true;
      this.saveOrder(order);
    }
  }

  isOrderLocked(id: OrderId): boolean {
    const order = this.getOrder(id);
    return order?.locked ?? false;
  }
}

export const store = new InMemoryStore();
