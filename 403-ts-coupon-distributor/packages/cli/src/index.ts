#!/usr/bin/env node

import {
  CouponType,
  DistributeCouponRequest,
  DistributeCouponResponse,
  QueryUserCouponsRequest,
  QueryUserCouponsResponse,
  RedeemCouponRequest,
  RedeemCouponResponse,
  GetUserCouponRecommendRequest,
  GetUserCouponRecommendResponse,
  UserFilterType,
  ValidityType,
} from '@coupon/shared';
import { httpClient } from './http-client';
import {
  formatCoupons,
  formatDistributionRecords,
  formatError,
  formatRedeemResult,
  formatRecommendResult,
  printHelp,
} from './utils';

function parseArgs(args: string[]): { command: string | null; options: Record<string, string | boolean | number> } {
  const options: Record<string, string | boolean | number> = {};
  let command: string | null = null;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    if (arg.startsWith('--')) {
      const key = arg.slice(2).replace(/-/g, '-');
      const nextArg = args[i + 1];

      if (nextArg === undefined || nextArg.startsWith('--')) {
        options[key] = true;
      } else {
        const numValue = Number(nextArg);
        options[key] = Number.isNaN(numValue) ? nextArg : numValue;
        i++;
      }
    } else if (!arg.startsWith('-') && command === null) {
      command = arg;
    }
  }

  return { command, options };
}

function getRequiredOption(options: Record<string, unknown>, key: string): string {
  const value = options[key];
  if (value === undefined || value === null) {
    throw new Error(`缺少必要参数: --${key}`);
  }
  return String(value);
}

function getOptionalOption(options: Record<string, unknown>, key: string, defaultValue: string): string {
  const value = options[key];
  if (value === undefined || value === null) {
    return defaultValue;
  }
  return String(value);
}

function getNumberOption(options: Record<string, unknown>, key: string): number {
  const value = options[key];
  if (value === undefined || value === null) {
    throw new Error(`缺少必要参数: --${key}`);
  }
  const num = Number(value);
  if (Number.isNaN(num)) {
    throw new Error(`参数 --${key} 必须是数字`);
  }
  return num;
}

function parseCommaSeparatedList(value: string): string[] {
  return value.split(',').map(s => s.trim()).filter(s => s.length > 0);
}

async function handleDistribute(options: Record<string, unknown>): Promise<void> {
  const type = getRequiredOption(options, 'type') as CouponType;
  const name = getRequiredOption(options, 'name');
  const value = getNumberOption(options, 'value');
  const threshold = getNumberOption(options, 'threshold');
  const validityType = getRequiredOption(options, 'validity-type') as ValidityType;
  const distributionType = getRequiredOption(options, 'distribution-type') as 'TARGETED' | 'ALL';
  const distributorId = getRequiredOption(options, 'distributor-id');

  const request: DistributeCouponRequest = {
    type,
    name,
    value,
    threshold,
    validity: { type: validityType },
    distributionType,
    distributorId,
  };

  if (validityType === ValidityType.FIXED_DATE) {
    request.validity.startDate = getRequiredOption(options, 'start-date');
    request.validity.endDate = getRequiredOption(options, 'end-date');
  } else {
    request.validity.days = getNumberOption(options, 'days');
  }

  if (distributionType === 'TARGETED') {
    const userIdsStr = getRequiredOption(options, 'user-ids');
    request.userIds = parseCommaSeparatedList(userIdsStr);
  } else {
    const userFilter = options['user-filter'];
    if (userFilter) {
      request.userFilter = userFilter as UserFilterType;
    }
    const newUserDays = options['new-user-days'];
    if (newUserDays !== undefined && newUserDays !== null) {
      request.newUserThresholdDays = Number(newUserDays);
    }
  }

  const response = await httpClient.post<DistributeCouponResponse>('/api/coupons/distribute', request);

  if (response.success) {
    console.log(formatDistributionRecords(response.distributionRecords));
  } else {
    console.error(formatError(response.errorCode, response.errorMessage));
    process.exit(1);
  }
}

async function handleRedeem(options: Record<string, unknown>): Promise<void> {
  const userId = getRequiredOption(options, 'user-id');
  const orderAmount = getNumberOption(options, 'order-amount');

  let couponIds: string[] = [];
  const couponIdsStr = options['coupon-ids'];
  if (couponIdsStr) {
    couponIds = parseCommaSeparatedList(String(couponIdsStr));
  }

  if (couponIds.length === 0) {
    const recommendRequest: GetUserCouponRecommendRequest = {
      userId,
      orderAmount,
    };
    const recommendResponse = await httpClient.post<GetUserCouponRecommendResponse>('/api/coupons/recommend', recommendRequest);

    if (recommendResponse.success && recommendResponse.recommendedCouponIds.length > 0) {
      couponIds = recommendResponse.recommendedCouponIds;
      console.log(`自动选择推荐的优惠券: ${couponIds.join(', ')}`);
    }
  }

  const request: RedeemCouponRequest = {
    userId,
    orderAmount,
    couponIds,
  };

  const response = await httpClient.post<RedeemCouponResponse>('/api/coupons/redeem', request);

  if (response.success) {
    console.log(formatRedeemResult(
      response.originalAmount,
      response.discountAmount,
      response.finalAmount,
      response.usedCoupons
    ));
  } else {
    console.error(formatError(response.errorCode, response.errorMessage));
    process.exit(1);
  }
}

async function handleQuery(options: Record<string, unknown>): Promise<void> {
  const userId = getRequiredOption(options, 'user-id');
  const includeUsed = Boolean(options['include-used']);
  const includeExpired = Boolean(options['include-expired']);

  const request: QueryUserCouponsRequest = {
    userId,
    includeUsed,
    includeExpired,
  };

  const response = await httpClient.post<QueryUserCouponsResponse>('/api/coupons/query', request);

  if (response.success) {
    console.log(formatCoupons(response.coupons));
  } else {
    console.error(formatError(response.errorCode, response.errorMessage));
    process.exit(1);
  }
}

async function handleRecommend(options: Record<string, unknown>): Promise<void> {
  const userId = getRequiredOption(options, 'user-id');
  const orderAmount = getNumberOption(options, 'order-amount');

  const request: GetUserCouponRecommendRequest = {
    userId,
    orderAmount,
  };

  const response = await httpClient.post<GetUserCouponRecommendResponse>('/api/coupons/recommend', request);

  if (response.success) {
    console.log(formatRecommendResult(response.recommendedCouponIds, response.totalDiscount));
  } else {
    console.error(formatError(response.errorCode, response.errorMessage));
    process.exit(1);
  }
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    printHelp();
    process.exit(0);
  }

  const { command, options } = parseArgs(args);

  try {
    switch (command) {
      case 'distribute':
        await handleDistribute(options);
        break;
      case 'redeem':
        await handleRedeem(options);
        break;
      case 'query':
        await handleQuery(options);
        break;
      case 'recommend':
        await handleRecommend(options);
        break;
      case 'help':
      default:
        printHelp();
        break;
    }
  } catch (error) {
    if (error instanceof Error) {
      console.error(`错误: ${error.message}`);
    } else {
      console.error('发生未知错误');
    }
    process.exit(1);
  }
}

main().catch((error) => {
  console.error('程序执行失败:', error);
  process.exit(1);
});
