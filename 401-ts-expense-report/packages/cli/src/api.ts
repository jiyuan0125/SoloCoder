import http from 'http';
import { SERVER_PORT, SERVER_HOST } from '@expense-report/shared';

interface ApiClientOptions {
  host?: string;
  port?: number;
  userId?: string;
}

class ApiClient {
  private host: string;
  private port: number;
  private userId?: string;

  constructor(options: ApiClientOptions = {}) {
    this.host = options.host || SERVER_HOST;
    this.port = options.port || SERVER_PORT;
    if (options.userId !== undefined) {
      this.userId = options.userId;
    }
  }

  setUserId(userId: string): void {
    this.userId = userId;
  }

  getUserId(): string | undefined {
    return this.userId;
  }

  private request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const bodyStr = body !== undefined ? JSON.stringify(body) : undefined;
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (bodyStr !== undefined) {
        headers['Content-Length'] = Buffer.byteLength(bodyStr).toString();
      }

      if (this.userId) {
        headers['X-User-Id'] = this.userId;
      }

      const options: http.RequestOptions = {
        hostname: this.host,
        port: this.port,
        path,
        method,
        headers,
      };

      const req = http.request(options, (res) => {
        const chunks: Buffer[] = [];

        res.on('data', (chunk: Buffer) => {
          chunks.push(chunk);
        });

        res.on('end', () => {
          const responseBody = Buffer.concat(chunks).toString('utf-8');
          try {
            const data = JSON.parse(responseBody) as T;
            resolve(data);
          } catch (err) {
            reject(new Error(`Failed to parse response: ${responseBody}`));
          }
        });
      });

      req.on('error', (err) => {
        reject(err);
      });

      if (bodyStr !== undefined) {
        req.write(bodyStr);
      }

      req.end();
    });
  }

  async createExpense(params: {
    employeeId: string;
    employeeName: string;
    date: string;
    amount: number;
    category: string;
    reason: string;
    voucherFileName: string;
  }): Promise<unknown> {
    return this.request('POST', '/api/expenses', params);
  }

  async updateExpense(params: {
    expenseId: string;
    date?: string;
    amount?: number;
    category?: string;
    reason?: string;
    voucherFileName?: string;
    supplementaryInfo?: string;
  }): Promise<unknown> {
    return this.request('PUT', '/api/expenses', params);
  }

  async submitExpense(expenseId: string): Promise<unknown> {
    return this.request('POST', '/api/expenses/submit', { expenseId });
  }

  async approveExpense(params: {
    expenseId: string;
    comment?: string;
  }): Promise<unknown> {
    return this.request('POST', '/api/expenses/approve', params);
  }

  async rejectExpense(params: {
    expenseId: string;
    comment?: string;
  }): Promise<unknown> {
    return this.request('POST', '/api/expenses/reject', params);
  }

  async batchApprove(params: {
    expenseIds: string[];
    comment?: string;
  }): Promise<unknown> {
    return this.request('POST', '/api/expenses/batch-approve', params);
  }

  async payExpense(expenseId: string): Promise<unknown> {
    return this.request('POST', '/api/expenses/pay', { expenseId });
  }

  async resubmitExpense(expenseId: string): Promise<unknown> {
    return this.request('POST', '/api/expenses/resubmit', { expenseId });
  }

  async deleteExpense(expenseId: string): Promise<unknown> {
    return this.request('DELETE', '/api/expenses', { expenseId });
  }

  async getExpense(expenseId: string): Promise<unknown> {
    return this.request('GET', `/api/expenses/get?id=${encodeURIComponent(expenseId)}`);
  }

  async listExpenses(params: {
    startDate?: string;
    endDate?: string;
    status?: string;
    employeeId?: string;
    page?: number;
    pageSize?: number;
  } = {}): Promise<unknown> {
    const queryParts: string[] = [];
    if (params.startDate) {
      queryParts.push(`startDate=${encodeURIComponent(params.startDate)}`);
    }
    if (params.endDate) {
      queryParts.push(`endDate=${encodeURIComponent(params.endDate)}`);
    }
    if (params.status) {
      queryParts.push(`status=${encodeURIComponent(params.status)}`);
    }
    if (params.employeeId) {
      queryParts.push(`employeeId=${encodeURIComponent(params.employeeId)}`);
    }
    if (params.page !== undefined) {
      queryParts.push(`page=${params.page}`);
    }
    if (params.pageSize !== undefined) {
      queryParts.push(`pageSize=${params.pageSize}`);
    }
    const queryString = queryParts.length > 0 ? `?${queryParts.join('&')}` : '';
    return this.request('GET', `/api/expenses/list${queryString}`);
  }

  async getOperationLogs(params: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    page?: number;
    pageSize?: number;
  } = {}): Promise<unknown> {
    const queryParts: string[] = [];
    if (params.startDate) {
      queryParts.push(`startDate=${encodeURIComponent(params.startDate)}`);
    }
    if (params.endDate) {
      queryParts.push(`endDate=${encodeURIComponent(params.endDate)}`);
    }
    if (params.userId) {
      queryParts.push(`userId=${encodeURIComponent(params.userId)}`);
    }
    if (params.page !== undefined) {
      queryParts.push(`page=${params.page}`);
    }
    if (params.pageSize !== undefined) {
      queryParts.push(`pageSize=${params.pageSize}`);
    }
    const queryString = queryParts.length > 0 ? `?${queryParts.join('&')}` : '';
    return this.request('GET', `/api/logs/operations${queryString}`);
  }

  async getAuditLogs(params: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    action?: string;
    page?: number;
    pageSize?: number;
  } = {}): Promise<unknown> {
    const queryParts: string[] = [];
    if (params.startDate) {
      queryParts.push(`startDate=${encodeURIComponent(params.startDate)}`);
    }
    if (params.endDate) {
      queryParts.push(`endDate=${encodeURIComponent(params.endDate)}`);
    }
    if (params.userId) {
      queryParts.push(`userId=${encodeURIComponent(params.userId)}`);
    }
    if (params.action) {
      queryParts.push(`action=${encodeURIComponent(params.action)}`);
    }
    if (params.page !== undefined) {
      queryParts.push(`page=${params.page}`);
    }
    if (params.pageSize !== undefined) {
      queryParts.push(`pageSize=${params.pageSize}`);
    }
    const queryString = queryParts.length > 0 ? `?${queryParts.join('&')}` : '';
    return this.request('GET', `/api/logs/audit${queryString}`);
  }

  async getUser(userId: string): Promise<unknown> {
    return this.request('GET', `/api/users/get?id=${encodeURIComponent(userId)}`);
  }
}

export const apiClient = new ApiClient();
