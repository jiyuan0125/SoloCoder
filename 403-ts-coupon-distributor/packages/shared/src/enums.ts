export enum CouponType {
  FULL_REDUCTION = 'FULL_REDUCTION',
  DISCOUNT = 'DISCOUNT',
  FIXED_AMOUNT = 'FIXED_AMOUNT',
}

export enum ValidityType {
  FIXED_DATE = 'FIXED_DATE',
  DAYS_AFTER_RECEIVE = 'DAYS_AFTER_RECEIVE',
}

export enum UserFilterType {
  ALL = 'ALL',
  NEW_USERS = 'NEW_USERS',
  OLD_USERS = 'OLD_USERS',
}

export const MIN_DISCOUNT_RATE = 50;
export const MAX_COUPONS_PER_ORDER = 2;
