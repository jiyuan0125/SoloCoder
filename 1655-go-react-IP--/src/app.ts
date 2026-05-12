import express, { Application, Request, Response, NextFunction } from 'express';
import { parseClientIP } from './middleware/ipParser';
import { validateRequestIP } from './middleware/ipValidation';
import { rateLimiter } from './middleware/rateLimiter';
import { riskAssessor } from './middleware/riskAssessor';
import rulesRouter from './routes/rules';
import conditionsRouter from './routes/conditions';

const app: Application = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use(parseClientIP);

app.get('/health', (_req: Request, res: Response): void => {
  res.json({ 
    status: 'ok', 
    timestamp: Date.now(),
    gateway: 'api-security-gateway'
  });
});

app.use('/gateway/rules', rulesRouter);
app.use('/gateway/rules/:ruleId/conditions', conditionsRouter);

app.use('/api', validateRequestIP, rateLimiter, riskAssessor, (req: Request, res: Response): void => {
  res.json({
    message: '请求通过安全网关',
    ip: req.realIP,
    path: req.path,
    method: req.method,
    riskScore: req.riskScore,
    riskFactors: req.riskFactors
  });
});

app.get('/', (_req: Request, res: Response): void => {
  res.json({
    name: 'API Security Gateway',
    version: '1.0.0',
    endpoints: {
      health: '/health',
      rules: '/gateway/rules',
      protected: '/api/*'
    }
  });
});

app.use((_req: Request, res: Response): void => {
  res.status(404).json({ error: '路由不存在' });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction): void => {
  console.error('Gateway Error:', err);
  res.status(500).json({ 
    error: '网关内部错误',
    message: err.message
  });
});

app.listen(PORT, (): void => {
  console.log(`API Security Gateway running on port ${PORT}`);
  console.log(`Management API: /gateway/rules`);
  console.log(`Protected API: /api/*`);
});

export default app;
