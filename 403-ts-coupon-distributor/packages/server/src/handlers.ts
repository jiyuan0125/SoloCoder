import { IncomingMessage, ServerResponse } from 'http';
import {
  CouponType,
  DistributeCouponRequest,
  DistributeCouponResponse,
  ErrorCode,
  errorMessages,
  GetUserCouponRecommendRequest,
  GetUserCouponRecommendResponse,
  HealthCheckResponse,
  QueryUserCouponsRequest,
  QueryUserCouponsResponse,
  RedeemCouponRequest,
  RedeemCouponResponse,
  UserFilterType,
  ValidityType,
} from '@coupon/shared';
import { couponService } from './coupon-service';
import { dataStore } from './data-store';

async function parseRequestBody(req: IncomingMessage): Promise<unknown> {
  return new Promise((resolve, reject) => {
    let body = '';
    req.on('data', (chunk) => {
      body += chunk.toString();
    });
    req.on('end', () => {
      try {
        resolve(body ? JSON.parse(body) : null);
      } catch (error) {
        reject(error);
      }
    });
  });
}

function sendJsonResponse(res: ServerResponse, statusCode: number, data: unknown): void {
  res.setHeader('Content-Type', 'application/json');
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  res.writeHead(statusCode);
  res.end(JSON.stringify(data, null, 2));
}

function handleOptions(req: IncomingMessage, res: ServerResponse): void {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  res.writeHead(204);
  res.end();
}

function handleHealthCheck(req: IncomingMessage, res: ServerResponse): void {
  const response: HealthCheckResponse = {
    success: true,
    timestamp: new Date().toISOString(),
    version: '1.0.0',
  };
  sendJsonResponse(res, 200, response);
}

async function handleDistributeCoupon(req: IncomingMessage, res: ServerResponse): Promise<void> {
  try {
    const body = (await parseRequestBody(req)) as DistributeCouponRequest;

    if (!body.type || !body.name || body.value === undefined || body.threshold === undefined) {
      const response: DistributeCouponResponse = {
        success: false,
        count: 0,
        distributionRecords: [],
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    if (body.distributionType === 'TARGETED' && (!body.userIds || body.userIds.length === 0)) {
      const response: DistributeCouponResponse = {
        success: false,
        count: 0,
        distributionRecords: [],
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    if (!body.distributorId) {
      const response: DistributeCouponResponse = {
        success: false,
        count: 0,
        distributionRecords: [],
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const result = couponService.distributeCoupon(
      body.type as CouponType,
      body.name,
      body.value,
      body.threshold,
      body.validity as any,
      body.distributionType,
      body.distributorId,
      body.userIds,
      body.userFilter as UserFilterType,
      body.newUserThresholdDays
    );

    if (!result.success) {
      const response: DistributeCouponResponse = {
        success: false,
        count: 0,
        distributionRecords: [],
        errorCode: result.errorCode,
        errorMessage: result.errorMessage,
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const response: DistributeCouponResponse = {
      success: true,
      count: result.count,
      distributionRecords: result.records,
    };
    sendJsonResponse(res, 200, response);
  } catch (error) {
    const response: DistributeCouponResponse = {
      success: false,
      count: 0,
      distributionRecords: [],
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: errorMessages[ErrorCode.UNKNOWN_ERROR],
    };
    sendJsonResponse(res, 500, response);
  }
}

async function handleRedeemCoupon(req: IncomingMessage, res: ServerResponse): Promise<void> {
  try {
    const body = (await parseRequestBody(req)) as RedeemCouponRequest;

    if (!body.userId || body.orderAmount === undefined) {
      const response: RedeemCouponResponse = {
        success: false,
        originalAmount: body.orderAmount || 0,
        discountAmount: 0,
        finalAmount: body.orderAmount || 0,
        usedCoupons: [],
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const result = couponService.redeemCoupons(body.userId, body.orderAmount, body.couponIds || []);

    if (!result.success) {
      const response: RedeemCouponResponse = {
        success: false,
        originalAmount: result.originalAmount,
        discountAmount: result.discountAmount,
        finalAmount: result.finalAmount,
        usedCoupons: result.usedCoupons,
        errorCode: result.errorCode,
        errorMessage: result.errorMessage,
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const response: RedeemCouponResponse = {
      success: true,
      originalAmount: result.originalAmount,
      discountAmount: result.discountAmount,
      finalAmount: result.finalAmount,
      usedCoupons: result.usedCoupons,
    };
    sendJsonResponse(res, 200, response);
  } catch (error) {
    const response: RedeemCouponResponse = {
      success: false,
      originalAmount: 0,
      discountAmount: 0,
      finalAmount: 0,
      usedCoupons: [],
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: errorMessages[ErrorCode.UNKNOWN_ERROR],
    };
    sendJsonResponse(res, 500, response);
  }
}

async function handleQueryUserCoupons(req: IncomingMessage, res: ServerResponse): Promise<void> {
  try {
    const body = (await parseRequestBody(req)) as QueryUserCouponsRequest;

    if (!body.userId) {
      const response: QueryUserCouponsResponse = {
        success: false,
        coupons: [],
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const coupons = couponService.getUserAllCoupons(
      body.userId,
      body.includeUsed ?? false,
      body.includeExpired ?? false
    );

    const response: QueryUserCouponsResponse = {
      success: true,
      coupons,
    };
    sendJsonResponse(res, 200, response);
  } catch (error) {
    const response: QueryUserCouponsResponse = {
      success: false,
      coupons: [],
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: errorMessages[ErrorCode.UNKNOWN_ERROR],
    };
    sendJsonResponse(res, 500, response);
  }
}

async function handleGetRecommendCoupons(req: IncomingMessage, res: ServerResponse): Promise<void> {
  try {
    const body = (await parseRequestBody(req)) as GetUserCouponRecommendRequest;

    if (!body.userId || body.orderAmount === undefined) {
      const response: GetUserCouponRecommendResponse = {
        success: false,
        recommendedCouponIds: [],
        totalDiscount: 0,
        errorCode: ErrorCode.INVALID_PARAMS,
        errorMessage: errorMessages[ErrorCode.INVALID_PARAMS],
      };
      sendJsonResponse(res, 400, response);
      return;
    }

    const result = couponService.getRecommendedCoupons(body.userId, body.orderAmount);

    const response: GetUserCouponRecommendResponse = {
      success: true,
      recommendedCouponIds: result.couponIds,
      totalDiscount: result.totalDiscount,
    };
    sendJsonResponse(res, 200, response);
  } catch (error) {
    const response: GetUserCouponRecommendResponse = {
      success: false,
      recommendedCouponIds: [],
      totalDiscount: 0,
      errorCode: ErrorCode.UNKNOWN_ERROR,
      errorMessage: errorMessages[ErrorCode.UNKNOWN_ERROR],
    };
    sendJsonResponse(res, 500, response);
  }
}

export async function handleRequest(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const method = req.method;
  const url = req.url || '/';

  if (method === 'OPTIONS') {
    handleOptions(req, res);
    return;
  }

  if (method === 'GET' && url === '/health') {
    handleHealthCheck(req, res);
    return;
  }

  if (method === 'POST') {
    if (url === '/api/coupons/distribute') {
      await handleDistributeCoupon(req, res);
      return;
    }
    if (url === '/api/coupons/redeem') {
      await handleRedeemCoupon(req, res);
      return;
    }
    if (url === '/api/coupons/query') {
      await handleQueryUserCoupons(req, res);
      return;
    }
    if (url === '/api/coupons/recommend') {
      await handleGetRecommendCoupons(req, res);
      return;
    }
  }

  sendJsonResponse(res, 404, {
    success: false,
    error: 'Not Found',
    message: `Route ${method} ${url} not found`,
  });
}
