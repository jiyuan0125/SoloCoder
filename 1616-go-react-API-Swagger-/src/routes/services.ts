import { Router, Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { getDb } from '../database/db';
import { ApiError } from '../middleware/errorHandler';
import { Service } from '../types';
import { parseSwagger, validateJson } from '../utils/swaggerParser';
import axios, { AxiosError } from 'axios';
import { getConfig } from '../config/config';

const router = Router();

const formatService = (row: any): Service => ({
  id: row.id,
  name: row.name,
  description: row.description || undefined,
  base_url: row.base_url || undefined,
  created_at: row.created_at,
  updated_at: row.updated_at,
  has_swagger: row.has_swagger === 1,
});

const getServiceById = (id: string): Service => {
  const db = getDb();
  const row = db.prepare('SELECT * FROM services WHERE id = ?').get(id) as any;
  if (!row) {
    throw new ApiError(404, 'Service not found');
  }
  return formatService(row);
};

router.get('/', (req: Request, res: Response): void => {
  const db = getDb();
  const rows = db.prepare('SELECT * FROM services ORDER BY created_at DESC').all() as any[];
  const services = rows.map(formatService);
  res.json(services);
});

router.post('/', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const { name, description, base_url } = req.body;

    if (!name || typeof name !== 'string' || name.trim() === '') {
      throw new ApiError(400, 'Service name is required');
    }

    const id = uuidv4();
    const now = new Date().toISOString();

    const db = getDb();
    db.prepare(
      'INSERT INTO services (id, name, description, base_url, created_at, updated_at, has_swagger) VALUES (?, ?, ?, ?, ?, ?, 0)'
    ).run(id, name.trim(), description || null, base_url || null, now, now);

    const service = getServiceById(id);
    res.status(201).json(service);
  } catch (err) {
    next(err);
  }
});

router.get('/:id', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const service = getServiceById(req.params.id);
    res.json(service);
  } catch (err) {
    next(err);
  }
});

router.put('/:id', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const service = getServiceById(req.params.id);
    const { name, description, base_url } = req.body;

    if (name !== undefined && (typeof name !== 'string' || name.trim() === '')) {
      throw new ApiError(400, 'Service name cannot be empty');
    }

    const now = new Date().toISOString();
    const db = getDb();

    const newName = name !== undefined ? name.trim() : service.name;
    const newDesc = description !== undefined ? (description || null) : (service.description || null);
    const newBaseUrl = base_url !== undefined ? (base_url || null) : (service.base_url || null);

    db.prepare(
      'UPDATE services SET name = ?, description = ?, base_url = ?, updated_at = ? WHERE id = ?'
    ).run(newName, newDesc, newBaseUrl, now, service.id);

    const updatedService = getServiceById(service.id);
    res.json(updatedService);
  } catch (err) {
    next(err);
  }
});

router.delete('/:id', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const service = getServiceById(req.params.id);
    const db = getDb();

    const debugCount = db.prepare(
      'SELECT COUNT(*) as count FROM debug_records WHERE service_id = ?'
    ).get(service.id) as any;

    if (debugCount.count > 0) {
      throw new ApiError(409, 'Cannot delete service: debug records exist');
    }

    db.prepare('DELETE FROM services WHERE id = ?').run(service.id);
    res.status(204).send();
  } catch (err) {
    next(err);
  }
});

router.post('/:id/swagger', async (req: Request, res: Response, next: NextFunction): Promise<void> => {
  try {
    const service = getServiceById(req.params.id);
    let swaggerJson: any;

    if (typeof req.body === 'string') {
      swaggerJson = validateJson(req.body);
    } else if (typeof req.body === 'object') {
      swaggerJson = req.body;
    } else {
      throw new ApiError(400, 'Invalid request body: expected JSON');
    }

    const result = parseSwagger(swaggerJson);
    const now = new Date().toISOString();
    const db = getDb();

    const deleteStmt = db.prepare('DELETE FROM api_endpoints WHERE service_id = ?');
    const insertStmt = db.prepare(
      `INSERT INTO api_endpoints 
       (id, service_id, path, method, summary, description, tags, parameters, request_body, responses, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    );

    const updateServiceStmt = db.prepare('UPDATE services SET has_swagger = 1, updated_at = ? WHERE id = ?');

    db.transaction(() => {
      deleteStmt.run(service.id);
      updateServiceStmt.run(now, service.id);

      for (const api of result.apis) {
        insertStmt.run(
          api.id,
          service.id,
          api.path,
          api.method,
          api.summary || null,
          api.description || null,
          JSON.stringify(api.tags),
          JSON.stringify(api.parameters),
          api.requestBody ? JSON.stringify(api.requestBody) : null,
          JSON.stringify(api.responses),
          now
        );
      }
    })();

    res.json({
      message: 'Swagger uploaded successfully',
      total_apis: result.apis.length,
      skipped_apis: result.skipped_apis,
    });
  } catch (err: unknown) {
    if (err instanceof SyntaxError) {
      return next(new ApiError(400, err.message));
    }
    if (err instanceof Error && err.message?.includes('JSON parse')) {
      return next(new ApiError(400, err.message));
    }
    next(err);
  }
});

router.get('/:id/apis', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const service = getServiceById(req.params.id);
    const db = getDb();

    const rows = db.prepare(
      'SELECT id, path, method, summary, tags, created_at FROM api_endpoints WHERE service_id = ? ORDER BY path, method'
    ).all(service.id) as any[];

    const apis = rows.map((row) => ({
      id: row.id,
      path: row.path,
      method: row.method,
      summary: row.summary || undefined,
      tags: JSON.parse(row.tags || '[]'),
      created_at: row.created_at,
    }));

    res.json(apis);
  } catch (err) {
    next(err);
  }
});

const getApiById = (serviceId: string, apiId: string) => {
  const db = getDb();
  const row = db.prepare(
    'SELECT * FROM api_endpoints WHERE id = ? AND service_id = ?'
  ).get(apiId, serviceId) as any;

  if (!row) {
    throw new ApiError(404, 'API not found');
  }

  return {
    id: row.id,
    service_id: row.service_id,
    path: row.path,
    method: row.method,
    summary: row.summary || undefined,
    description: row.description || undefined,
    tags: JSON.parse(row.tags || '[]'),
    parameters: JSON.parse(row.parameters || '[]'),
    request_body: row.request_body ? JSON.parse(row.request_body) : undefined,
    responses: JSON.parse(row.responses || '[]'),
    created_at: row.created_at,
  };
};

router.get('/:id/apis/:apiId', (req: Request, res: Response, next: NextFunction): void => {
  try {
    getServiceById(req.params.id);
    const api = getApiById(req.params.id, req.params.apiId);
    res.json(api);
  } catch (err) {
    next(err);
  }
});

router.post('/:id/apis/:apiId/debug', async (req: Request, res: Response, next: NextFunction): Promise<void> => {
  try {
    const service = getServiceById(req.params.id);
    const api = getApiById(req.params.id, req.params.apiId);
    const config = getConfig();

    const { headers, query, body, base_url, timeout } = req.body;

    const targetBaseUrl = base_url || service.base_url;
    if (!targetBaseUrl) {
      throw new ApiError(400, 'No base URL configured for this service');
    }

    const requestTimeout = timeout || config.proxy_timeout;
    const startTime = Date.now();

    let path = api.path;
    if (path.includes('{')) {
      const pathParams = (req.body.path_params || {}) as Record<string, string>;
      for (const [key, value] of Object.entries(pathParams)) {
        path = path.replace(`{${key}}`, encodeURIComponent(String(value)));
      }
    }

    const url = new URL(path, targetBaseUrl.endsWith('/') ? targetBaseUrl : targetBaseUrl + '/');
    
    if (query && typeof query === 'object') {
      for (const [key, value] of Object.entries(query)) {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value));
        }
      }
    }

    try {
      const response = await axios.request({
        method: api.method.toLowerCase(),
        url: url.toString(),
        headers: headers || {},
        data: body,
        timeout: requestTimeout,
        validateStatus: () => true,
      });

      const duration = Date.now() - startTime;

      const debugRecord = {
        id: uuidv4(),
        service_id: service.id,
        api_id: api.id,
        request: JSON.stringify({
          method: api.method,
          url: url.toString(),
          headers: headers || {},
          query: query || {},
          body: body,
        }),
        response: JSON.stringify({
          status: response.status,
          headers: response.headers,
          data: response.data,
        }),
        status_code: response.status,
        duration_ms: duration,
        created_at: new Date().toISOString(),
      };

      const db = getDb();
      db.prepare(
        `INSERT INTO debug_records (id, service_id, api_id, request, response, status_code, duration_ms, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
      ).run(
        debugRecord.id,
        debugRecord.service_id,
        debugRecord.api_id,
        debugRecord.request,
        debugRecord.response,
        debugRecord.status_code,
        debugRecord.duration_ms,
        debugRecord.created_at
      );

      res.json({
        status: response.status,
        headers: response.headers,
        data: response.data,
        duration_ms: duration,
      });
    } catch (error: any) {
      const duration = Date.now() - startTime;

      if (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT') {
        res.status(504).json({
          error: 'Gateway Timeout',
          message: 'Request timed out',
          duration_ms: duration,
          timeout_ms: requestTimeout,
        });
        return;
      }

      if (error.code === 'ECONNREFUSED' || error.code === 'ENOTFOUND' || error.code === 'EAI_AGAIN') {
        res.status(502).json({
          error: 'Bad Gateway',
          message: 'Target service is unreachable',
          details: error.message,
          duration_ms: duration,
        });
        return;
      }

      throw new ApiError(502, `Proxy error: ${error.message}`);
    }
  } catch (err) {
    next(err);
  }
});

router.get('/:id/analysis', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const service = getServiceById(req.params.id);
    const db = getDb();

    const rows = db.prepare(
      'SELECT parameters, tags, responses FROM api_endpoints WHERE service_id = ?'
    ).all(service.id) as any[];

    const tags: Record<string, number> = {};
    const requiredParamsByType: Record<string, number> = {};
    const statusCodes: Record<string, number> = {};

    for (const row of rows) {
      const apiTags = JSON.parse(row.tags || '[]') as string[];
      for (const tag of apiTags) {
        tags[tag] = (tags[tag] || 0) + 1;
      }

      const params = JSON.parse(row.parameters || '[]') as any[];
      for (const param of params) {
        if (param.required) {
          const paramType = param.in || 'unknown';
          requiredParamsByType[paramType] = (requiredParamsByType[paramType] || 0) + 1;
        }
      }

      const responses = JSON.parse(row.responses || '[]') as any[];
      for (const response of responses) {
        const statusCode = response.statusCode;
        statusCodes[statusCode] = (statusCodes[statusCode] || 0) + 1;
      }
    }

    res.json({
      tags,
      required_params_by_type: requiredParamsByType,
      status_codes: statusCodes,
    });
  } catch (err) {
    next(err);
  }
});

export { router as servicesRouter };
