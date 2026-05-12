import type { Request, Response, NextFunction } from 'express';
import axios, { AxiosError } from 'axios';
import { updateKeyStats } from './auth';
import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';

const DEFAULT_TIMEOUT = 10000;

function logRequest(
  req: Request,
  statusCode: number,
  responseTimeMs: number
): void {
  const id = uuidv4();
  db.prepare(`
    INSERT INTO request_logs (id, method, path, status_code, response_time_ms, upstream_name, client_ip, key_id, timestamp)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    req.method,
    req.path,
    statusCode,
    responseTimeMs,
    req.matchedService?.service.name || null,
    req.clientIp || null,
    req.apiKey?.id || null,
    Date.now()
  );
}

function normalizeUpstreamUrl(url: string): string {
  return url.endsWith('/') ? url.slice(0, -1) : url;
}

export function proxyMiddleware(req: Request, res: Response, _next: NextFunction): void {
  const startTime = Date.now();

  const matched = req.matchedService;
  if (!matched) {
    const elapsed = Date.now() - startTime;
    res.status(404).json({ error: 'Not Found' });
    logRequest(req, 404, elapsed);
    if (req.apiKey?.id) {
      updateKeyStats(req.apiKey.id, false);
    }
    return;
  }

  const upstreamUrl = normalizeUpstreamUrl(matched.service.upstream_url);
  const targetUrl = upstreamUrl + matched.remainingPath;
  const queryString = req.url.split('?')[1];
  const finalUrl = queryString ? `${targetUrl}?${queryString}` : targetUrl;

  const headers: Record<string, string> = {};
  for (const [key, value] of Object.entries(req.headers)) {
    if (key.toLowerCase() === 'host' || key.toLowerCase() === 'content-length') continue;
    if (Array.isArray(value)) {
      headers[key] = value.join(', ');
    } else if (value !== undefined) {
      headers[key] = value;
    }
  }

  axios({
    method: req.method as any,
    url: finalUrl,
    data: req.body,
    headers,
    timeout: DEFAULT_TIMEOUT,
    responseType: 'stream',
    validateStatus: () => true,
    maxRedirects: 0
  }).then(response => {
    const elapsed = Date.now() - startTime;
    const statusCode = response.status;

    for (const [key, value] of Object.entries(response.headers)) {
      if (key.toLowerCase() === 'transfer-encoding' || key.toLowerCase() === 'connection') continue;
      if (value !== undefined) {
        res.setHeader(key, value);
      }
    }

    res.status(statusCode);
    response.data.pipe(res);

    response.data.on('end', () => {
      logRequest(req, statusCode, elapsed);
      if (req.apiKey?.id) {
        const success = statusCode >= 200 && statusCode < 400;
        updateKeyStats(req.apiKey.id, success);
      }
    });
  }).catch((error: AxiosError) => {
    const elapsed = Date.now() - startTime;

    if (error.code === 'ECONNABORTED' || error.message.includes('timeout')) {
      res.status(504).json({
        error: 'Gateway Timeout',
        message: 'Upstream service request timed out'
      });
      logRequest(req, 504, elapsed);
    } else if (error.code === 'ECONNREFUSED' || error.code === 'ECONNRESET' || error.code === 'ENOTFOUND') {
      res.status(502).json({
        error: 'Bad Gateway',
        message: 'Upstream service is unreachable'
      });
      logRequest(req, 502, elapsed);
    } else {
      res.status(502).json({
        error: 'Bad Gateway',
        message: 'Error connecting to upstream service'
      });
      logRequest(req, 502, elapsed);
    }

    if (req.apiKey?.id) {
      updateKeyStats(req.apiKey.id, false);
    }
  });
}
