import express, { Express, Request, Response } from 'express';
import sensitiveWordsRouter from './routes/sensitiveWords';
import auditRouter from './routes/audit';

const app: Express = express();
const PORT = 3000;

app.use(express.json());

app.use((err: unknown, req: Request, res: Response, next: express.NextFunction) => {
  if (err instanceof SyntaxError && 'body' in err) {
    res.status(400).json({ error: '无效的 JSON' });
    return;
  }
  next(err);
});

app.use('/api/sensitive-words', sensitiveWordsRouter);
app.use('/api/audit', auditRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use((req: Request, res: Response) => {
  res.status(404).json({ error: 'Not Found' });
});

app.use((err: unknown, req: Request, res: Response, next: express.NextFunction) => {
  console.error(err);
  res.status(500).json({ error: '服务器内部错误' });
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Content Audit Service running on port ${PORT}`);
  });
}

export default app;
