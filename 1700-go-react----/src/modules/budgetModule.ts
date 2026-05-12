import { Budget, BudgetItem } from '../models/types';

const budgets: Map<string, Budget> = new Map();

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
}

export function createBudget(
  period: string,
  totalAmount: number,
  items: Omit<BudgetItem, 'id'>[]
): Budget {
  const id = generateId();
  const now = new Date();

  const budgetItems: BudgetItem[] = items.map(item => ({
    ...item,
    id: generateId(),
  }));

  const calculatedTotal = budgetItems.reduce((sum, item) => sum + item.amount, 0);

  const budget: Budget = {
    id,
    period,
    totalAmount: calculatedTotal,
    originalTotal: calculatedTotal,
    items: budgetItems,
    createdAt: now,
    updatedAt: now,
  };

  budgets.set(id, budget);
  return budget;
}

export function getBudget(id: string): Budget | undefined {
  return budgets.get(id);
}

export function getAllBudgets(): Budget[] {
  return Array.from(budgets.values());
}

export function adjustBudget(
  budgetId: string,
  newTotalAmount: number
): Budget | null {
  const budget = budgets.get(budgetId);
  if (!budget) return null;

  const oldTotal = budget.totalAmount;
  const totalAdjustment = newTotalAmount - oldTotal;

  if (Math.abs(totalAdjustment) < 0.01) {
    return budget;
  }

  const pendingItems = budget.items.filter(item => item.status === 'pending');
  const settledItems = budget.items.filter(item => item.status === 'settled');

  const settledTotal = settledItems.reduce((sum, item) => sum + item.amount, 0);
  const pendingTotal = pendingItems.reduce((sum, item) => sum + item.amount, 0);

  if (pendingTotal === 0) {
    if (settledTotal !== newTotalAmount) {
      throw new Error('所有项目已结算，无法调整预算');
    }
    return budget;
  }

  const updatedPendingItems: BudgetItem[] = pendingItems.map(item => {
    const ratio = item.amount / pendingTotal;
    const adjustment = totalAdjustment * ratio;
    const newAmount = Math.max(0, item.amount + adjustment);

    return {
      ...item,
      amount: Math.round(newAmount * 100) / 100,
    };
  });

  const newPendingTotal = updatedPendingItems.reduce((sum, item) => sum + item.amount, 0);
  const finalTotal = settledTotal + newPendingTotal;

  const updatedBudget: Budget = {
    ...budget,
    totalAmount: Math.round(finalTotal * 100) / 100,
    items: [...settledItems, ...updatedPendingItems],
    updatedAt: new Date(),
  };

  const totalDiff = Math.abs(updatedBudget.totalAmount - newTotalAmount);
  if (totalDiff > 0.02 && updatedPendingItems.length > 0) {
    const remainingDiff = newTotalAmount - updatedBudget.totalAmount;
    updatedPendingItems[0].amount = Math.round((updatedPendingItems[0].amount + remainingDiff) * 100) / 100;
    updatedBudget.totalAmount = Math.round((settledTotal + updatedPendingItems.reduce((sum, item) => sum + item.amount, 0)) * 100) / 100;
  }

  budgets.set(budgetId, updatedBudget);
  return updatedBudget;
}

export function settleBudgetItem(
  budgetId: string,
  itemId: string
): Budget | null {
  const budget = budgets.get(budgetId);
  if (!budget) return null;

  const itemIndex = budget.items.findIndex(item => item.id === itemId);
  if (itemIndex === -1) return null;

  const item = budget.items[itemIndex];
  if (item.status === 'settled') {
    return budget;
  }

  const updatedItem: BudgetItem = {
    ...item,
    status: 'settled',
    settledAmount: item.amount,
  };

  const updatedItems = [...budget.items];
  updatedItems[itemIndex] = updatedItem;

  const updatedBudget: Budget = {
    ...budget,
    items: updatedItems,
    updatedAt: new Date(),
  };

  budgets.set(budgetId, updatedBudget);
  return updatedBudget;
}

export function addBudgetItem(
  budgetId: string,
  newItem: Omit<BudgetItem, 'id'>
): Budget | null {
  const budget = budgets.get(budgetId);
  if (!budget) return null;

  const item: BudgetItem = {
    ...newItem,
    id: generateId(),
  };

  const updatedBudget: Budget = {
    ...budget,
    totalAmount: budget.totalAmount + item.amount,
    items: [...budget.items, item],
    updatedAt: new Date(),
  };

  budgets.set(budgetId, updatedBudget);
  return updatedBudget;
}

export function removeBudgetItem(
  budgetId: string,
  itemId: string
): Budget | null {
  const budget = budgets.get(budgetId);
  if (!budget) return null;

  const itemToRemove = budget.items.find(item => item.id === itemId);
  if (!itemToRemove) return null;

  if (itemToRemove.status === 'settled') {
    throw new Error('已结算的项目不能删除');
  }

  const updatedItems = budget.items.filter(item => item.id !== itemId);
  const updatedTotal = updatedItems.reduce((sum, item) => sum + item.amount, 0);

  const updatedBudget: Budget = {
    ...budget,
    totalAmount: updatedTotal,
    items: updatedItems,
    updatedAt: new Date(),
  };

  budgets.set(budgetId, updatedBudget);
  return updatedBudget;
}
