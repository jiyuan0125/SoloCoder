import { CouponType, ValidityType, UserFilterType } from './enums';
import { Coupon, DistributionRecord } from './types';

export interface DistributeCouponRequest {
  type: CouponType;
  name: string;
  value: number;
  threshold: number;
  validity: {
    type: ValidityType;
    startDate?: string;
    endDate?: string;
    days?: number;
  };
  distributionType: 'TARGETED' | 'ALL';
  userIds?: string[];
  userFilter?: UserFilterType;
  newUserThresholdDays?: number;
  distributorId: string;
}

export interface DistributeCouponResponse {
  success: boolean;
  count: number;
  distributionRecords: DistributionRecord[];
  errorCode?: number;
  errorMessage?: string;
}

export interface RedeemCouponRequest {
  userId: string;
  orderAmount: number;
  couponIds: string[];
}

export interface RedeemCouponResponse {
  success: boolean;
  originalAmount: number;
  discountAmount: number;
  finalAmount: number;
  usedCoupons: UsedCouponInfo[];
  errorCode?: number;
  errorMessage?: string;
}

export interface UsedCouponInfo {
  couponId: string;
  type: CouponType;
  name: string;
  value: number;
  discountAmount: number;
}

export interface QueryUserCouponsRequest {
  userId: string;
  includeUsed?: boolean;
  includeExpired?: boolean;
}

export interface QueryUserCouponsResponse {
  success: boolean;
  coupons: Coupon[];
  errorCode?: number;
  errorMessage?: string;
}

export interface GetUserCouponRecommendRequest {
  userId: string;
  orderAmount: number;
}

export interface GetUserCouponRecommendResponse {
  success: boolean;
  recommendedCouponIds: string[];
  totalDiscount: number;
  errorCode?: number;
  errorMessage?: string;
}

export interface HealthCheckResponse {
  success: boolean;
  timestamp: string;
  version: string;
}
