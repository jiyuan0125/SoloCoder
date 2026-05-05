import {
  AmountInCents,
  SalespersonId,
  SalesContribution,
} from '@commission-tracker/shared';
import { roundToCents } from './commissionCalculator';

export interface AllocationResult {
  salespersonId: SalespersonId;
  allocatedAmount: AmountInCents;
  isPrimary: boolean;
}

export function allocateTeamOrderAmount(
  totalAmount: AmountInCents,
  contributions: SalesContribution[]
): AllocationResult[] {
  if (contributions.length === 0) {
    return [];
  }

  if (contributions.length === 1) {
    return [
      {
        salespersonId: contributions[0].salespersonId,
        allocatedAmount: totalAmount,
        isPrimary: contributions[0].isPrimary,
      },
    ];
  }

  const primaryContribution = contributions.find(c => c.isPrimary);
  const nonPrimaryContributions = contributions.filter(c => !c.isPrimary);

  if (!primaryContribution) {
    throw new Error('团队订单必须有一个主销售');
  }

  const primaryMinimum = roundToCents(totalAmount * 0.5);
  const remainingAfterPrimary = totalAmount - primaryMinimum;
  const nonPrimaryCount = nonPrimaryContributions.length;

  const result: AllocationResult[] = [];

  if (nonPrimaryCount === 0) {
    result.push({
      salespersonId: primaryContribution.salespersonId,
      allocatedAmount: totalAmount,
      isPrimary: true,
    });
    return result;
  }

  const perNonPrimary = roundToCents(remainingAfterPrimary / nonPrimaryCount);

  let totalAllocated = primaryMinimum + perNonPrimary * nonPrimaryCount;
  let roundingDiff = totalAmount - totalAllocated;

  result.push({
    salespersonId: primaryContribution.salespersonId,
    allocatedAmount: primaryMinimum + Math.max(roundingDiff, 0),
    isPrimary: true,
  });

  for (const contrib of nonPrimaryContributions) {
    result.push({
      salespersonId: contrib.salespersonId,
      allocatedAmount: perNonPrimary,
      isPrimary: false,
    });
  }

  return result;
}

export function validateSalesContributions(contributions: SalesContribution[]): void {
  if (contributions.length === 0) {
    throw new Error('订单必须关联至少一个销售');
  }

  const primaryCount = contributions.filter(c => c.isPrimary).length;

  if (primaryCount === 0 && contributions.length > 1) {
    throw new Error('团队订单必须指定一个主销售');
  }

  if (primaryCount > 1) {
    throw new Error('订单只能有一个主销售');
  }

  const uniqueIds = new Set(contributions.map(c => c.salespersonId));
  if (uniqueIds.size !== contributions.length) {
    throw new Error('同一个销售不能在一个订单中出现多次');
  }
}
