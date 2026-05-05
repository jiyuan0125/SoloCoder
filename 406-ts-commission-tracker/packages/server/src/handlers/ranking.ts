import {
  IncomingMessage,
  ServerResponse,
} from 'http';
import {
  RankingFilter,
  successResponse,
  errorResponse,
  ErrorCodes,
  getErrorMessage,
  Quarter,
} from '@commission-tracker/shared';
import { getRankings } from '../services/rankingService';
import { sendJsonResponse } from '../utils';

export async function handleGetRankings(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);

    const dimension = url.searchParams.get('dimension') || 'monthly';
    const yearParam = url.searchParams.get('year');
    const monthParam = url.searchParams.get('month');
    const quarterParam = url.searchParams.get('quarter');

    const now = new Date();
    const year: number = yearParam ? parseInt(yearParam, 10) : now.getFullYear();

    if (dimension !== 'monthly' && dimension !== 'quarterly' && dimension !== 'yearly') {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_RANKING_DIMENSION, getErrorMessage(ErrorCodes.INVALID_RANKING_DIMENSION))
      );
      return;
    }

    let month: number | undefined;
    let quarter: number | undefined;

    if (dimension === 'monthly') {
      month = monthParam ? parseInt(monthParam, 10) : now.getMonth() + 1;
      if (month < 1 || month > 12) {
        sendJsonResponse(
          res,
          400,
          errorResponse(ErrorCodes.INVALID_MONTH, getErrorMessage(ErrorCodes.INVALID_MONTH))
        );
        return;
      }
    } else if (dimension === 'quarterly') {
      quarter = quarterParam ? parseInt(quarterParam, 10) : Math.ceil((now.getMonth() + 1) / 3);
      if (quarter < 1 || quarter > 4) {
        sendJsonResponse(
          res,
          400,
          errorResponse(ErrorCodes.INVALID_QUARTER, getErrorMessage(ErrorCodes.INVALID_QUARTER))
        );
        return;
      }
    }

    const filter: RankingFilter = {
      dimension,
      year,
      month: month as number | undefined,
      quarter: quarter as Quarter | undefined,
    };

    const rankings = getRankings(filter);

    sendJsonResponse(res, 200, successResponse(rankings));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}
