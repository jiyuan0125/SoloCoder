export const SOCIAL_INSURANCE_BASE_MIN = 2000 * 100;
export const SOCIAL_INSURANCE_BASE_MAX = 25000 * 100;
export const MINIMUM_WAGE = 2500 * 100;
export const OVERTIME_CAP_RATIO = 0.5;

export const TAX_BRACKETS = [
  { threshold: 3000 * 100, rate: 0.03 },
  { threshold: 12000 * 100, rate: 0.10 },
  { threshold: 25000 * 100, rate: 0.20 },
  { threshold: 35000 * 100, rate: 0.25 },
  { threshold: Infinity, rate: 0.30 },
] as const;

export const OVERTIME_MULTIPLIERS = {
  weekday: 1.5,
  weekend: 2.0,
  holiday: 3.0,
} as const;
