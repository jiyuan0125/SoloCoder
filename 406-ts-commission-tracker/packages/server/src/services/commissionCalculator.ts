import {
  AmountInCents,
  Percentage,
  CommissionTier,
  TierBreakdown,
} from '@commission-tracker/shared';

export const COMMISSION_TIERS: CommissionTier[] = [
  { from: 0, to: 10000000, rate: 0.03 },
  { from: 10000000, to: 50000000, rate: 0.05 },
  { from: 50000000, to: null, rate: 0.08 },
];

export function roundToCents(amount: number): AmountInCents {
  return Math.round(amount);
}

export function calculateCommissionForSales(
  totalSales: AmountInCents,
  tiers: CommissionTier[] = COMMISSION_TIERS
): { commission: AmountInCents; breakdown: TierBreakdown[] } {
  let remainingSales = totalSales;
  let totalCommission = 0;
  const breakdown: TierBreakdown[] = [];

  for (const tier of tiers) {
    if (remainingSales <= 0) break;

    const tierStart = tier.from;
    const tierEnd = tier.to ?? Number.MAX_SAFE_INTEGER;
    const tierSize = tierEnd - tierStart;

    const salesInTier = Math.min(remainingSales, tierSize);
    
    if (salesInTier <= 0) continue;

    const commissionForTier = roundToCents(salesInTier * tier.rate);
    totalCommission += commissionForTier;

    let tierLabel = '';
    if (tier.from === 0 && tier.to !== null) {
      tierLabel = `${(tier.to / 100).toFixed(0)}元以内`;
    } else if (tier.from > 0 && tier.to !== null) {
      tierLabel = `${(tier.from / 100).toFixed(0)}元-${(tier.to / 100).toFixed(0)}元`;
    } else {
      tierLabel = `${(tier.from / 100).toFixed(0)}元以上`;
    }

    breakdown.push({
      tier: tierLabel,
      salesAmount: salesInTier,
      commission: commissionForTier,
    });

    remainingSales -= salesInTier;
  }

  return {
    commission: totalCommission,
    breakdown,
  };
}

export function applyProbationDiscount(
  commission: AmountInCents,
  discountRate: Percentage = 0.7
): AmountInCents {
  return roundToCents(commission * discountRate);
}
