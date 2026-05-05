import { CouponType, ValidityType, UserFilterType } from './enums';

export interface FixedDateValidity {
  type: ValidityType.FIXED_DATE;
  startDate: string;
  endDate: string;
}

export interface DaysAfterReceiveValidity {
  type: ValidityType.DAYS_AFTER_RECEIVE;
  days: number;
}

export type ValidityConfig = FixedDateValidity | DaysAfterReceiveValidity;

export interface Coupon {
  id: string;
  type: CouponType;
  name: string;
  value: number;
  threshold: number;
  validity: ValidityConfig;
  receiveTime: string;
  expireTime: string;
  isUsed: boolean;
  useTime: string | null;
  userId: string;
  distributionRecordId: string;
}

export interface DistributionRecord {
  id: string;
  distributorId: string;
  receiveUserId: string;
  couponId: string;
  distributionTime: string;
}

export interface User {
  id: string;
  name: string;
  registerTime: string;
}
