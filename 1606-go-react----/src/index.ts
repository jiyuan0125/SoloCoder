import express, { Request, Response } from 'express';
import sensitiveWordsRouter from './routes/sensitiveWords';
import rulesRouter from './routes/rules';
import contentsRouter from './routes/contents';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.use('/api/sensitive-words', sensitiveWordsRouter);
app.use('/api/rules', rulesRouter);
app.use('/api/contents', contentsRouter);

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: 'Not Found' });
});

app.listen(PORT, () => {
  console.log(`Content review system is running on port ${PORT}`);
});

export default app;
