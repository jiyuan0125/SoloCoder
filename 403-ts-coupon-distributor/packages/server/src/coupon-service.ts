import {
  Coupon,
  CouponType,
  DistributionRecord,
  ErrorCode,
  errorMessages,
  FixedDateValidity,
  MAX_COUPONS_PER_ORDER,
  MIN_DISCOUNT_RATE,
  User,
  UserFilterType,
  ValidityConfig,
  ValidityType,
} from '@coupon/shared';
import { dataStore } from './data-store';
import * as crypto from 'crypto';

function generateId(): string {
  return crypto.randomUUID();
}

function getCurrentTime(): string {
  return new Date().toISOString();
}

function calculateExpireTime(validity: ValidityConfig, receiveTime: string): string {
  if (validity.type === ValidityType.FIXED_DATE) {
    return validity.endDate;
  } else {
    const receiveDate = new Date(receiveTime);
    receiveDate.setDate(receiveDate.getDate() + validity.days);
    return receiveDate.toISOString();
  }
}

function isCouponValid(coupon: Coupon, currentTime: string): boolean {
  if (coupon.isUsed) {
    return false;
  }
  const now = new Date(currentTime);
  const expireDate = new Date(coupon.expireTime);
  return now <= expireDate;
}

function isOrderAmountSufficient(coupon: Coupon, orderAmount: number): boolean {
  return orderAmount >= coupon.threshold;
}

function sortCouponsByExpireTime(coupons: Coupon[]): Coupon[] {
  return [...coupons].sort((a, b) => {
    return new Date(a.expireTime).getTime() - new Date(b.expireTime).getTime();
  });
}

export class CouponService {
  public distributeCoupon(
    type: CouponType,
    name: string,
    value: number,
    threshold: number,
    validity: ValidityConfig,
    distributionType: 'TARGETED' | 'ALL',
    distributorId: string,
    userIds?: string[],
    userFilter?: UserFilterType,
    newUserThresholdDays?: number
  ): { success: boolean; count: number; records: DistributionRecord[]; errorCode?: ErrorCode; errorMessage?: string } {
    if (type === CouponType.DISCOUNT && value < MIN_DISCOUNT_RATE) {
      return {
        success: false,
        count: 0,
        records: [],
        errorCode: ErrorCode.INVALID_DISCOUNT_RATE,
        errorMessage: errorMessages[ErrorCode.INVALID_DISCOUNT_RATE],
      };
    }

    if (validity.type === ValidityType.FIXED_DATE) {
      const startDate = new Date(validity.startDate);
      const endDate = new Date(validity.endDate);
      if (startDate >= endDate) {
        return {
          success: false,
          count: 0,
          records: [],
          errorCode: ErrorCode.INVALID_VALIDITY_DATE,
          errorMessage: errorMessages[ErrorCode.INVALID_VALIDITY_DATE],
        };
      }
    }

    if (validity.type === ValidityType.DAYS_AFTER_RECEIVE && validity.days <= 0) {
      return {
        success: false,
        count: 0,
        records: [],
        errorCode: ErrorCode.INVALID_VALIDITY_DATE,
        errorMessage: errorMessages[ErrorCode.INVALID_VALIDITY_DATE],
      };
    }

    const targetUsers = this.getTargetUsers(distributionType, userIds, userFilter, newUserThresholdDays);

    if (targetUsers.length === 0) {
      return {
        success: true,
        count: 0,
        records: [],
      };
    }

    const currentTime = getCurrentTime();
    const records: DistributionRecord[] = [];

    for (const user of targetUsers) {
      const couponId = generateId();
      const recordId = generateId();

      const coupon: Coupon = {
        id: couponId,
        type,
        name,
        value,
        threshold,
        validity,
        receiveTime: currentTime,
        expireTime: calculateExpireTime(validity, currentTime),
        isUsed: false,
        useTime: null,
        userId: user.id,
        distributionRecordId: recordId,
      };

      const record: DistributionRecord = {
        id: recordId,
        distributorId,
        receiveUserId: user.id,
        couponId,
        distributionTime: currentTime,
      };

      dataStore.addCoupon(coupon);
      dataStore.addDistributionRecord(record);
      records.push(record);
    }

    return {
      success: true,
      count: records.length,
      records,
    };
  }

  private getTargetUsers(
    distributionType: 'TARGETED' | 'ALL',
    userIds?: string[],
    userFilter?: UserFilterType,
    newUserThresholdDays?: number
  ): User[] {
    if (distributionType === 'TARGETED' && userIds) {
      return userIds
        .map(id => dataStore.getUser(id))
        .filter((user): user is User => user !== undefined);
    }

    let users = dataStore.getAllUsers();

    if (userFilter === UserFilterType.NEW_USERS) {
      const thresholdDays = newUserThresholdDays ?? 7;
      const cutoffTime = new Date();
      cutoffTime.setDate(cutoffTime.getDate() - thresholdDays);
      users = users.filter(user => new Date(user.registerTime) >= cutoffTime);
    } else if (userFilter === UserFilterType.OLD_USERS) {
      const thresholdDays = newUserThresholdDays ?? 7;
      const cutoffTime = new Date();
      cutoffTime.setDate(cutoffTime.getDate() - thresholdDays);
      users = users.filter(user => new Date(user.registerTime) < cutoffTime);
    }

    return users;
  }

  public getUserAvailableCoupons(userId: string): Coupon[] {
    const userCoupons = dataStore.getUserCoupons(userId);
    const currentTime = getCurrentTime();
    const validCoupons = userCoupons.filter(coupon => isCouponValid(coupon, currentTime));
    return sortCouponsByExpireTime(validCoupons);
  }

  public getUserAllCoupons(userId: string, includeUsed: boolean = false, includeExpired: boolean = false): Coupon[] {
    const userCoupons = dataStore.getUserCoupons(userId);
    const currentTime = getCurrentTime();

    let coupons = userCoupons;

    if (!includeUsed) {
      coupons = coupons.filter(c => !c.isUsed);
    }

    if (!includeExpired) {
      coupons = coupons.filter(c => {
        const expireDate = new Date(c.expireTime);
        const now = new Date(currentTime);
        return now <= expireDate;
      });
    }

    return sortCouponsByExpireTime(coupons);
  }

  public getRecommendedCoupons(
    userId: string,
    orderAmount: number
  ): { couponIds: string[]; totalDiscount: number } {
    const availableCoupons = this.getUserAvailableCoupons(userId);
    const eligibleCoupons = availableCoupons.filter(c => isOrderAmountSufficient(c, orderAmount));

    if (eligibleCoupons.length === 0) {
      return { couponIds: [], totalDiscount: 0 };
    }

    const fixedAmountCoupons = eligibleCoupons.filter(c => c.type === CouponType.FIXED_AMOUNT);
    const fullReductionCoupons = eligibleCoupons.filter(c => c.type === CouponType.FULL_REDUCTION);
    const discountCoupons = eligibleCoupons.filter(c => c.type === CouponType.DISCOUNT);

    let bestCombination: Coupon[] = [];
    let bestDiscount = 0;

    if (discountCoupons.length > 0) {
      const bestDiscountCoupon = discountCoupons[0];
      const discountAmount = this.calculateDiscount(bestDiscountCoupon, orderAmount);
      bestCombination = [bestDiscountCoupon];
      bestDiscount = discountAmount;
    }

    if (fixedAmountCoupons.length > 0 || fullReductionCoupons.length > 0) {
      const fixedCombinations: Coupon[][] = [];

      if (fixedAmountCoupons.length > 0) {
        fixedCombinations.push([fixedAmountCoupons[0]]);
      }
      if (fullReductionCoupons.length > 0) {
        fixedCombinations.push([fullReductionCoupons[0]]);
      }
      if (fixedAmountCoupons.length > 0 && fullReductionCoupons.length > 0) {
        fixedCombinations.push([fixedAmountCoupons[0], fullReductionCoupons[0]]);
      }

      for (const combo of fixedCombinations) {
        const totalDiscount = combo.reduce((sum, coupon) => sum + this.calculateDiscount(coupon, orderAmount), 0);
        if (totalDiscount > bestDiscount) {
          bestDiscount = totalDiscount;
          bestCombination = combo;
        }
      }
    }

    return {
      couponIds: bestCombination.map(c => c.id),
      totalDiscount: bestDiscount,
    };
  }

  public redeemCoupons(
    userId: string,
    orderAmount: number,
    couponIds: string[]
  ): {
    success: boolean;
    originalAmount: number;
    discountAmount: number;
    finalAmount: number;
    usedCoupons: { couponId: string; type: CouponType; name: string; value: number; discountAmount: number }[];
    errorCode?: ErrorCode;
    errorMessage?: string;
  } {
    if (couponIds.length === 0) {
      return {
        success: true,
        originalAmount: orderAmount,
        discountAmount: 0,
        finalAmount: orderAmount,
        usedCoupons: [],
      };
    }

    if (couponIds.length > MAX_COUPONS_PER_ORDER) {
      return {
        success: false,
        originalAmount: orderAmount,
        discountAmount: 0,
        finalAmount: orderAmount,
        usedCoupons: [],
        errorCode: ErrorCode.TOO_MANY_COUPONS,
        errorMessage: errorMessages[ErrorCode.TOO_MANY_COUPONS],
      };
    }

    const coupons = couponIds.map(id => dataStore.getCoupon(id)).filter((c): c is Coupon => c !== undefined);

    if (coupons.length !== couponIds.length) {
      return {
        success: false,
        originalAmount: orderAmount,
        discountAmount: 0,
        finalAmount: orderAmount,
        usedCoupons: [],
        errorCode: ErrorCode.COUPON_NOT_FOUND,
        errorMessage: errorMessages[ErrorCode.COUPON_NOT_FOUND],
      };
    }

    for (const coupon of coupons) {
      if (coupon.userId !== userId) {
        return {
          success: false,
          originalAmount: orderAmount,
          discountAmount: 0,
          finalAmount: orderAmount,
          usedCoupons: [],
          errorCode: ErrorCode.COUPON_NOT_FOUND,
          errorMessage: errorMessages[ErrorCode.COUPON_NOT_FOUND],
        };
      }
    }

    const currentTime = getCurrentTime();

    for (const coupon of coupons) {
      if (!isCouponValid(coupon, currentTime)) {
        if (coupon.isUsed) {
          return {
            success: false,
            originalAmount: orderAmount,
            discountAmount: 0,
            finalAmount: orderAmount,
            usedCoupons: [],
            errorCode: ErrorCode.COUPON_ALREADY_USED,
            errorMessage: errorMessages[ErrorCode.COUPON_ALREADY_USED],
          };
        }
        return {
          success: false,
          originalAmount: orderAmount,
          discountAmount: 0,
          finalAmount: orderAmount,
          usedCoupons: [],
          errorCode: ErrorCode.COUPON_EXPIRED,
          errorMessage: errorMessages[ErrorCode.COUPON_EXPIRED],
        };
      }
    }

    if (coupons.length === 2) {
      const hasDiscount = coupons.some(c => c.type === CouponType.DISCOUNT);
      const hasFixed = coupons.some(c => c.type === CouponType.FIXED_AMOUNT);
      const hasFullReduction = coupons.some(c => c.type === CouponType.FULL_REDUCTION);

      if (hasDiscount && (hasFixed || hasFullReduction)) {
        return {
          success: false,
          originalAmount: orderAmount,
          discountAmount: 0,
          finalAmount: orderAmount,
          usedCoupons: [],
          errorCode: ErrorCode.COUPON_CANNOT_STACK,
          errorMessage: errorMessages[ErrorCode.COUPON_CANNOT_STACK],
        };
      }

      if (coupons[0].type === coupons[1].type) {
        return {
          success: false,
          originalAmount: orderAmount,
          discountAmount: 0,
          finalAmount: orderAmount,
          usedCoupons: [],
          errorCode: ErrorCode.COUPON_CANNOT_STACK,
          errorMessage: errorMessages[ErrorCode.COUPON_CANNOT_STACK],
        };
      }
    }

    for (const coupon of coupons) {
      if (!isOrderAmountSufficient(coupon, orderAmount)) {
        return {
          success: false,
          originalAmount: orderAmount,
          discountAmount: 0,
          finalAmount: orderAmount,
          usedCoupons: [],
          errorCode: ErrorCode.ORDER_AMOUNT_NOT_ENOUGH,
          errorMessage: errorMessages[ErrorCode.ORDER_AMOUNT_NOT_ENOUGH],
        };
      }
    }

    let remainingAmount = orderAmount;
    const usedCoupons: {
      couponId: string;
      type: CouponType;
      name: string;
      value: number;
      discountAmount: number;
    }[] = [];

    const sortedCoupons = [...coupons];
    const fixedAmountIndex = sortedCoupons.findIndex(c => c.type === CouponType.FIXED_AMOUNT);
    if (fixedAmountIndex > 0) {
      const [fixed] = sortedCoupons.splice(fixedAmountIndex, 1);
      sortedCoupons.unshift(fixed);
    }

    for (const coupon of sortedCoupons) {
      const discountAmount = this.calculateDiscount(coupon, remainingAmount);
      usedCoupons.push({
        couponId: coupon.id,
        type: coupon.type,
        name: coupon.name,
        value: coupon.value,
        discountAmount,
      });
      remainingAmount = Math.max(0, remainingAmount - discountAmount);
    }

    const totalDiscount = orderAmount - remainingAmount;

    const useTime = getCurrentTime();
    for (const coupon of coupons) {
      const updatedCoupon = { ...coupon, isUsed: true, useTime };
      dataStore.updateCoupon(updatedCoupon);
    }

    return {
      success: true,
      originalAmount: orderAmount,
      discountAmount: totalDiscount,
      finalAmount: remainingAmount,
      usedCoupons,
    };
  }

  private calculateDiscount(coupon: Coupon, currentAmount: number): number {
    switch (coupon.type) {
      case CouponType.FULL_REDUCTION:
        return Math.min(coupon.value, currentAmount);
      case CouponType.FIXED_AMOUNT:
        return Math.min(coupon.value, currentAmount);
      case CouponType.DISCOUNT:
        return Math.floor(currentAmount * (1 - coupon.value / 100));
      default:
        return 0;
    }
  }
}

export const couponService = new CouponService();
