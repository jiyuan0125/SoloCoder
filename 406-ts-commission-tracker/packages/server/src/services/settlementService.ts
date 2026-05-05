import {
  Salesperson,
  Order,
  Settlement,
  SettlementStatus,
  AmountInCents,
  Month,
  Year,
  CommissionCalculation,
  SalesContribution,
} from '@commission-tracker/shared';
import { store } from '../store';
import {
  calculateCommissionForSales,
  applyProbationDiscount,
  roundToCents,
} from './commissionCalculator';
import { allocateTeamOrderAmount } from './teamAllocation';

export function isInProbationPeriod(
  salesperson: Salesperson,
  orderDate: string
): boolean {
  const joinDate = new Date(salesperson.joinDate);
  const orderDt = new Date(orderDate);

  const joinMonth = joinDate.getFullYear() * 12 + joinDate.getMonth();
  const orderMonth = orderDt.getFullYear() * 12 + orderDt.getMonth();

  return orderMonth === joinMonth;
}

export function calculateSalespersonCommissionForMonth(
  salesperson: Salesperson,
  month: Month,
  year: Year,
  orders: Order[]
): CommissionCalculation {
  const monthOrders = orders.filter(o => {
    const orderDate = new Date(o.date);
    return (
      orderDate.getFullYear() === year &&
      orderDate.getMonth() + 1 === month &&
      o.status === 'active'
    );
  });

  let totalSales: AmountInCents = 0;
  let totalCommissionBeforeDiscount: AmountInCents = 0;
  let hasProbationOrder = false;

  for (const order of monthOrders) {
    const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
    const salespersonAllocation = allocations.find(
      a => a.salespersonId === salesperson.id
    );

    if (!salespersonAllocation) continue;

    totalSales += salespersonAllocation.allocatedAmount;

    const inProbation = isInProbationPeriod(salesperson, order.date);
    if (inProbation) {
      hasProbationOrder = true;
    }
  }

  const { commission, breakdown } = calculateCommissionForSales(totalSales);
  totalCommissionBeforeDiscount = commission;

  let discountRate = 1;
  if (hasProbationOrder) {
    discountRate = 0.7;
  }

  const finalCommission =
    discountRate < 1
      ? applyProbationDiscount(totalCommissionBeforeDiscount, discountRate)
      : totalCommissionBeforeDiscount;

  return {
    salespersonId: salesperson.id,
    month,
    year,
    totalSales,
    commissionBeforeDiscount: totalCommissionBeforeDiscount,
    discountRate,
    finalCommission,
    tierBreakdown: breakdown,
  };
}

export function generateSettlementId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 8);
  return `STL-${timestamp}-${random}`;
}

export function performMonthlySettlement(
  month: Month,
  year: Year
): Settlement[] {
  const salespeople = store.getSalespeople().filter(s => s.status === 'active');
  const allOrders = store.getOrders();
  const settlements: Settlement[] = [];

  for (const salesperson of salespeople) {
    const settlement = createSettlementForSalesperson(
      salesperson,
      month,
      year,
      allOrders
    );
    settlements.push(settlement);
  }

  return settlements;
}

function createSettlementForSalesperson(
  salesperson: Salesperson,
  month: Month,
  year: Year,
  allOrders: Order[]
): Settlement {
  const monthOrders = allOrders.filter(o => {
    const orderDate = new Date(o.date);
    return (
      orderDate.getFullYear() === year &&
      orderDate.getMonth() + 1 === month &&
      o.status === 'active' &&
      o.locked === false
    );
  });

  const commissionCalc = calculateSalespersonCommissionForMonth(
    salesperson,
    month,
    year,
    allOrders
  );

  const settledOrders = monthOrders
    .filter(o =>
      o.salesContributions.some((c: SalesContribution) => c.salespersonId === salesperson.id)
    )
    .map(order => {
      const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
      const salespersonAllocation = allocations.find(
        a => a.salespersonId === salesperson.id
      );

      if (!salespersonAllocation) {
        return {
          orderId: order.id,
          amount: 0,
          commission: 0,
        };
      }

      const { commission } = calculateCommissionForSales(
        salespersonAllocation.allocatedAmount
      );

      const inProbation = isInProbationPeriod(salesperson, order.date);
      const finalCommission = inProbation
        ? applyProbationDiscount(commission)
        : commission;

      return {
        orderId: order.id,
        amount: salespersonAllocation.allocatedAmount,
        commission: finalCommission,
      };
    })
    .filter(o => o.amount > 0);

  const refundsInMonth = allOrders.filter(o => {
    const orderDate = new Date(o.date);
    return (
      orderDate.getFullYear() === year &&
      orderDate.getMonth() + 1 === month &&
      o.status === 'refunded' &&
      o.salesContributions.some((c: SalesContribution) => c.salespersonId === salesperson.id)
    );
  });

  const settledRefunds = refundsInMonth.map(order => {
    const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
    const salespersonAllocation = allocations.find(
      a => a.salespersonId === salesperson.id
    );

    if (!salespersonAllocation) {
      return {
        orderId: order.id,
        amount: 0,
        commissionDeducted: 0,
      };
    }

    const { commission } = calculateCommissionForSales(
      salespersonAllocation.allocatedAmount
    );

    const inProbation = isInProbationPeriod(salesperson, order.date);
    const finalCommission = inProbation
      ? applyProbationDiscount(commission)
      : commission;

    return {
      orderId: order.id,
      amount: salespersonAllocation.allocatedAmount,
      commissionDeducted: finalCommission,
    };
  });

  const totalRefundDeduction = settledRefunds.reduce(
    (sum, r) => sum + r.commissionDeducted,
    0
  );

  const balanceAdjustment = salesperson.balance < 0 ? salesperson.balance : 0;

  let finalSettlementAmount =
    commissionCalc.finalCommission - totalRefundDeduction + balanceAdjustment;

  if (salesperson.balance > 0) {
    finalSettlementAmount += salesperson.balance;
  }

  const settlement: Settlement = {
    id: generateSettlementId(),
    salespersonId: salesperson.id,
    month,
    year,
    status: 'completed',
    settledAt: new Date().toISOString(),
    details: {
      orders: settledOrders,
      refunds: settledRefunds,
      balanceAdjustment,
    },
    summary: {
      totalSales: commissionCalc.totalSales,
      totalCommissionBeforeDiscount: commissionCalc.commissionBeforeDiscount,
      totalDiscount:
        commissionCalc.commissionBeforeDiscount - commissionCalc.finalCommission,
      totalRefundDeduction,
      balanceAdjustment,
      finalSettlementAmount,
    },
  };

  for (const order of monthOrders) {
    store.lockOrder(order.id);
  }

  salesperson.balance = 0;
  store.saveSalesperson(salesperson);

  store.saveSettlement(settlement);

  return settlement;
}

export function performResignationSettlement(
  salesperson: Salesperson,
  allOrders: Order[]
): Settlement {
  const now = new Date();
  const currentMonth = now.getMonth() + 1 as Month;
  const currentYear = now.getFullYear();

  const unsettledOrders = allOrders.filter(o => {
    return (
      o.status === 'active' &&
      o.locked === false &&
      o.salesContributions.some((c: SalesContribution) => c.salespersonId === salesperson.id)
    );
  });

  let totalSales: AmountInCents = 0;
  const settledOrders: {
    orderId: string;
    amount: AmountInCents;
    commission: AmountInCents;
  }[] = [];

  for (const order of unsettledOrders) {
    const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
    const salespersonAllocation = allocations.find(
      a => a.salespersonId === salesperson.id
    );

    if (!salespersonAllocation) continue;

    totalSales += salespersonAllocation.allocatedAmount;

    const { commission } = calculateCommissionForSales(
      salespersonAllocation.allocatedAmount
    );

    const inProbation = isInProbationPeriod(salesperson, order.date);
    const finalCommission = inProbation
      ? applyProbationDiscount(commission)
      : commission;

    settledOrders.push({
      orderId: order.id,
      amount: salespersonAllocation.allocatedAmount,
      commission: finalCommission,
    });

    store.lockOrder(order.id);
  }

  const { commission: totalCommission, breakdown } = calculateCommissionForSales(totalSales);
  
  const hasProbation = unsettledOrders.some(o => isInProbationPeriod(salesperson, o.date));
  const finalCommission = hasProbation
    ? applyProbationDiscount(totalCommission)
    : totalCommission;

  const balanceAdjustment = salesperson.balance;
  const finalSettlementAmount = finalCommission + balanceAdjustment;

  const settlement: Settlement = {
    id: generateSettlementId(),
    salespersonId: salesperson.id,
    month: currentMonth,
    year: currentYear,
    status: 'completed',
    settledAt: new Date().toISOString(),
    details: {
      orders: settledOrders,
      refunds: [],
      balanceAdjustment,
    },
    summary: {
      totalSales,
      totalCommissionBeforeDiscount: totalCommission,
      totalDiscount: hasProbation ? totalCommission - finalCommission : 0,
      totalRefundDeduction: 0,
      balanceAdjustment,
      finalSettlementAmount,
    },
  };

  salesperson.balance = 0;
  salesperson.status = 'resigned';
  store.saveSalesperson(salesperson);

  store.saveSettlement(settlement);

  return settlement;
}

export function handleRefund(order: Order): void {
  if (order.locked) {
    const salespeople = store.getSalespeople();
    const affectedSalespersonIds = order.salesContributions.map((c: SalesContribution) => c.salespersonId);

    for (const salespersonId of affectedSalespersonIds) {
      const salesperson = salespeople.find(s => s.id === salespersonId);
      if (!salesperson) continue;

      const allocations = allocateTeamOrderAmount(order.amount, order.salesContributions);
      const salespersonAllocation = allocations.find(
        a => a.salespersonId === salespersonId
      );

      if (!salespersonAllocation) continue;

      const { commission } = calculateCommissionForSales(salespersonAllocation.allocatedAmount);
      const inProbation = isInProbationPeriod(salesperson, order.date);
      const commissionToDeduct = inProbation
        ? applyProbationDiscount(commission)
        : commission;

      salesperson.balance -= commissionToDeduct;
      store.saveSalesperson(salesperson);
    }
  }
}
