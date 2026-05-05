import {
  IncomingMessage,
  ServerResponse,
} from 'http';
import {
  CreateSalespersonRequest,
  Salesperson,
  Month,
  successResponse,
  errorResponse,
  ErrorCodes,
  getErrorMessage,
} from '@commission-tracker/shared';
import { store } from '../store';
import { performResignationSettlement } from '../services/settlementService';
import { parseRequestBody, sendJsonResponse, extractSalespersonId } from '../utils';

function generateSalespersonId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 6);
  return `SP-${timestamp}-${random}`;
}

export async function handleCreateSalesperson(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const body = await parseRequestBody(req);
    const request: CreateSalespersonRequest = JSON.parse(body);

    if (!request.name || !request.joinDate) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少必要参数：name 或 joinDate')
      );
      return;
    }

    const joinDate = new Date(request.joinDate);
    if (isNaN(joinDate.getTime())) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_DATE, '日期格式无效')
      );
      return;
    }

    const salesperson: Salesperson = {
      id: generateSalespersonId(),
      name: request.name,
      joinDate: request.joinDate,
      status: 'active',
      balance: 0,
    };

    store.saveSalesperson(salesperson);

    sendJsonResponse(res, 201, successResponse(salesperson));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleGetSalesperson(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const id = extractSalespersonId(url.pathname);

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少销售ID')
      );
      return;
    }

    const salesperson = store.getSalesperson(id);

    if (!salesperson) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.SALESPERSON_NOT_FOUND, getErrorMessage(ErrorCodes.SALESPERSON_NOT_FOUND))
      );
      return;
    }

    sendJsonResponse(res, 200, successResponse(salesperson));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleListSalespeople(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const salespeople = store.getSalespeople();
    sendJsonResponse(res, 200, successResponse(salespeople));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleResignSalesperson(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const id = extractSalespersonId(url.pathname);

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少销售ID')
      );
      return;
    }

    const salesperson = store.getSalesperson(id);

    if (!salesperson) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.SALESPERSON_NOT_FOUND, getErrorMessage(ErrorCodes.SALESPERSON_NOT_FOUND))
      );
      return;
    }

    if (salesperson.status === 'resigned') {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.SALESPERSON_ALREADY_RESIGNED, getErrorMessage(ErrorCodes.SALESPERSON_ALREADY_RESIGNED))
      );
      return;
    }

    const allOrders = store.getOrders();
    const settlement = performResignationSettlement(salesperson, allOrders);

    const updatedSalesperson = store.getSalesperson(id);

    sendJsonResponse(res, 200, successResponse({
      salesperson: updatedSalesperson,
      settlement,
    }));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleGetSalespersonBalance(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const id = extractSalespersonId(url.pathname);

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少销售ID')
      );
      return;
    }

    const salesperson = store.getSalesperson(id);

    if (!salesperson) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.SALESPERSON_NOT_FOUND, getErrorMessage(ErrorCodes.SALESPERSON_NOT_FOUND))
      );
      return;
    }

    const now = new Date();
    const currentMonth = now.getMonth() + 1;
    const currentYear = now.getFullYear();

    const settlementsThisMonth = store.getSettlementsForMonth(
      currentMonth as Month,
      currentYear
    );

    const thisMonthSettled = settlementsThisMonth
      .filter(s => s.salespersonId === id)
      .reduce((sum, s) => sum + s.summary.finalSettlementAmount, 0);

    sendJsonResponse(res, 200, successResponse({
      salespersonId: salesperson.id,
      balance: salesperson.balance,
      thisMonthSettled,
    }));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}
