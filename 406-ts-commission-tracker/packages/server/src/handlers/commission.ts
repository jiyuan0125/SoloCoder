import {
  IncomingMessage,
  ServerResponse,
} from 'http';
import {
  GetSalespersonCommissionRequest,
  successResponse,
  errorResponse,
  ErrorCodes,
  getErrorMessage,
  Month,
  Year,
} from '@commission-tracker/shared';
import { store } from '../store';
import { calculateSalespersonCommissionForMonth } from '../services/settlementService';
import { sendJsonResponse } from '../utils';

export async function handleGetSalespersonCommission(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const pathParts = url.pathname.split('/');
    const salespersonId = pathParts[pathParts.length - 4];

    const monthParam = url.searchParams.get('month');
    const yearParam = url.searchParams.get('year');

    if (!salespersonId) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少销售ID')
      );
      return;
    }

    const salesperson = store.getSalesperson(salespersonId);

    if (!salesperson) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.SALESPERSON_NOT_FOUND, getErrorMessage(ErrorCodes.SALESPERSON_NOT_FOUND))
      );
      return;
    }

    const now = new Date();
    let month: Month = monthParam ? (parseInt(monthParam, 10) as Month) : ((now.getMonth() + 1) as Month);
    let year: Year = yearParam ? parseInt(yearParam, 10) : now.getFullYear();

    if (month < 1 || month > 12) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_MONTH, getErrorMessage(ErrorCodes.INVALID_MONTH))
      );
      return;
    }

    const allOrders = store.getOrders();
    const commission = calculateSalespersonCommissionForMonth(salesperson, month, year, allOrders);

    sendJsonResponse(res, 200, successResponse(commission));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}
