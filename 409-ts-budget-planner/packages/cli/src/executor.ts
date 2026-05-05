import { CliCommand } from './types';
import { sendRequest } from './httpClient';
import { ApiResponse } from '@budget-planner/shared';

async function executeAndHandleResponse(
  response: { statusCode: number; body: unknown }
): Promise<unknown> {
  const apiResponse = response.body as ApiResponse;
  
  if (!apiResponse || apiResponse.success === undefined) {
    throw new Error(`Invalid response: ${JSON.stringify(response.body)}`);
  }

  if (apiResponse.success) {
    return apiResponse.data;
  } else {
    throw new Error(`${apiResponse.error.code}: ${apiResponse.error.message}`);
  }
}

export async function executeCommand(command: CliCommand): Promise<unknown> {
  switch (command.type) {
    case 'department-list': {
      const response = await sendRequest('GET', command.serverUrl, '/departments');
      return executeAndHandleResponse(response);
    }

    case 'department-create': {
      const response = await sendRequest('POST', command.serverUrl, '/departments', undefined, {
        id: command.id,
        name: command.name,
        managerId: command.managerId,
      });
      return executeAndHandleResponse(response);
    }

    case 'budget-get': {
      const response = await sendRequest('GET', command.serverUrl, '/budgets', {
        departmentId: command.departmentId,
        year: command.year,
      });
      return executeAndHandleResponse(response);
    }

    case 'budget-create': {
      const response = await sendRequest('POST', command.serverUrl, '/budgets', undefined, {
        departmentId: command.departmentId,
        year: command.year,
        totalAmount: command.totalAmount,
        allocations: command.allocations,
      });
      return executeAndHandleResponse(response);
    }

    case 'budget-adjust': {
      const response = await sendRequest('POST', command.serverUrl, '/budgets/adjust', undefined, {
        departmentId: command.departmentId,
        year: command.year,
        adjustments: command.adjustments,
        requestedBy: command.requestedBy,
      });
      return executeAndHandleResponse(response);
    }

    case 'budget-draft': {
      const response = await sendRequest('POST', command.serverUrl, '/budgets/draft', undefined, {
        newYear: command.newYear,
        sourceYear: command.sourceYear,
      });
      return executeAndHandleResponse(response);
    }

    case 'budget-confirm': {
      const response = await sendRequest('POST', command.serverUrl, '/budgets/confirm', undefined, {
        departmentId: command.departmentId,
        year: command.year,
        confirmedBy: command.confirmedBy,
      });
      return executeAndHandleResponse(response);
    }

    case 'reimbursement-submit': {
      const response = await sendRequest('POST', command.serverUrl, '/reimbursements', undefined, {
        departmentId: command.departmentId,
        year: command.year,
        month: command.month,
        items: command.items,
        description: command.description,
        requestedBy: command.requestedBy,
      });
      return executeAndHandleResponse(response);
    }

    case 'carryover-request': {
      const response = await sendRequest('POST', command.serverUrl, '/carryovers', undefined, {
        departmentId: command.departmentId,
        year: command.year,
        fromQuarter: command.fromQuarter,
        toQuarter: command.toQuarter,
        amount: command.amount,
        requestedBy: command.requestedBy,
      });
      return executeAndHandleResponse(response);
    }

    case 'stats-usage': {
      const response = await sendRequest('GET', command.serverUrl, '/statistics/usage', {
        departmentId: command.departmentId,
        category: command.category,
        startYear: command.startYear,
        startMonth: command.startMonth,
        endYear: command.endYear,
        endMonth: command.endMonth,
      });
      return executeAndHandleResponse(response);
    }

    case 'stats-trend': {
      const response = await sendRequest('GET', command.serverUrl, '/statistics/trend', {
        departmentId: command.departmentId,
        category: command.category,
        startYear: command.startYear,
        startMonth: command.startMonth,
        endYear: command.endYear,
        endMonth: command.endMonth,
        interval: command.interval,
      });
      return executeAndHandleResponse(response);
    }
  }
}
