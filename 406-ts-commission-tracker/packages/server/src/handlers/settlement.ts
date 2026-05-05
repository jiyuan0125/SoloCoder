import {
  IncomingMessage,
  ServerResponse,
} from 'http';
import {
  successResponse,
  errorResponse,
  ErrorCodes,
  getErrorMessage,
  Month,
  Year,
  AmountInCents,
} from '@commission-tracker/shared';
import { store } from '../store';
import { performMonthlySettlement } from '../services/settlementService';
import { parseRequestBody, sendJsonResponse } from '../utils';

export async function handleTriggerMonthlySettlement(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const body = await parseRequestBody(req);
    const request = JSON.parse(body);

    const now = new Date();
    let month: Month = request.month ? (parseInt(request.month, 10) as Month) : ((now.getMonth() + 1) as Month);
    let year: Year = request.year ? parseInt(request.year, 10) : now.getFullYear();

    if (month < 1 || month > 12) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_MONTH, getErrorMessage(ErrorCodes.INVALID_MONTH))
      );
      return;
    }

    const settlements = performMonthlySettlement(month, year);

    const totalSettled: AmountInCents = settlements.reduce(
      (sum, s) => sum + s.summary.finalSettlementAmount,
      0
    );

    sendJsonResponse(res, 200, successResponse({
      settlements,
      totalSettled,
    }));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleGetSettlements(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const salespersonId = url.searchParams.get('salespersonId');
    const yearParam = url.searchParams.get('year');
    const monthParam = url.searchParams.get('month');

    let settlements = store.getSettlements();

    if (salespersonId) {
      settlements = settlements.filter(s => s.salespersonId === salespersonId);
    }

    if (yearParam) {
      const year = parseInt(yearParam, 10);
      settlements = settlements.filter(s => s.year === year);
    }

    if (monthParam) {
      const month = parseInt(monthParam, 10);
      settlements = settlements.filter(s => s.month === month);
    }

    settlements.sort((a, b) => {
      if (a.year !== b.year) return b.year - a.year;
      return b.month - a.month;
    });

    sendJsonResponse(res, 200, successResponse(settlements));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleGetSettlement(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const id = url.pathname.split('/').pop();

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少结算单ID')
      );
      return;
    }

    const settlement = store.getSettlement(id);

    if (!settlement) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.SETTLEMENT_NOT_FOUND, getErrorMessage(ErrorCodes.SETTLEMENT_NOT_FOUND))
      );
      return;
    }

    sendJsonResponse(res, 200, successResponse(settlement));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}
