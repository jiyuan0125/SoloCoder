import {
  Department, AnnualBudget, MonthlyCategoryBudget, ReimbursementRecord, CarryoverRecord,
  Month, BudgetCategory, BudgetStatus
} from '@budget-planner/shared';

function createBudgetKey(departmentId: string, year: number): string {
  return `${departmentId}:${year}`;
}

interface StoreState {
  departments: Map<string, Department>;
  budgets: Map<string, AnnualBudget>;
  reimbursements: ReimbursementRecord[];
  carryovers: CarryoverRecord[];
}

let store: StoreState = {
  departments: new Map(),
  budgets: new Map(),
  reimbursements: [],
  carryovers: [],
};

export function getStore(): StoreState {
  return {
    departments: new Map(store.departments),
    budgets: new Map(store.budgets),
    reimbursements: [...store.reimbursements],
    carryovers: [...store.carryovers],
  };
}

export function addDepartment(department: Department): void {
  const newDepartments = new Map(store.departments);
  newDepartments.set(department.id, department);
  store = { ...store, departments: newDepartments };
}

export function getDepartment(id: string): Department | undefined {
  return store.departments.get(id);
}

export function getAllDepartments(): Department[] {
  return Array.from(store.departments.values());
}

export function addBudget(budget: AnnualBudget): void {
  const key = createBudgetKey(budget.departmentId, budget.year);
  const newBudgets = new Map(store.budgets);
  newBudgets.set(key, budget);
  store = { ...store, budgets: newBudgets };
}

export function getBudget(departmentId: string, year: number): AnnualBudget | undefined {
  const key = createBudgetKey(departmentId, year);
  return store.budgets.get(key);
}

export function getAllBudgets(): AnnualBudget[] {
  return Array.from(store.budgets.values());
}

export function updateBudget(budget: AnnualBudget): void {
  const key = createBudgetKey(budget.departmentId, budget.year);
  if (store.budgets.has(key)) {
    const newBudgets = new Map(store.budgets);
    newBudgets.set(key, budget);
    store = { ...store, budgets: newBudgets };
  }
}

export function getBudgetsByYear(year: number): AnnualBudget[] {
  const results: AnnualBudget[] = [];
  for (const budget of store.budgets.values()) {
    if (budget.year === year) {
      results.push(budget);
    }
  }
  return results;
}

export function addReimbursement(reimbursement: ReimbursementRecord): void {
  store = {
    ...store,
    reimbursements: [...store.reimbursements, reimbursement],
  };
}

export function getReimbursements(): ReimbursementRecord[] {
  return [...store.reimbursements];
}

export function addCarryover(carryover: CarryoverRecord): void {
  store = {
    ...store,
    carryovers: [...store.carryovers, carryover],
  };
}

export function getCarryovers(): CarryoverRecord[] {
  return [...store.carryovers];
}

export function getMonthlyBudget(
  budget: AnnualBudget,
  month: Month,
  category: BudgetCategory
): MonthlyCategoryBudget | undefined {
  return budget.monthlyBudgets.find(
    (b) => b.month === month && b.category === category
  );
}

export function updateMonthlyBudgets(
  budget: AnnualBudget,
  updatedBudgets: MonthlyCategoryBudget[]
): AnnualBudget {
  const newMonthlyBudgets = budget.monthlyBudgets.map((existing) => {
    const updated = updatedBudgets.find(
      (u) => u.month === existing.month && u.category === existing.category
    );
    return updated || existing;
  });

  return {
    ...budget,
    monthlyBudgets: newMonthlyBudgets,
  };
}

export function updateBudgetStatus(
  budget: AnnualBudget,
  status: BudgetStatus,
  confirmedAt?: string
): AnnualBudget {
  return {
    ...budget,
    status,
    confirmedAt,
  };
}
