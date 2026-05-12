import express, { Request, Response, NextFunction } from 'express';
import warehousesRouter from './routes/warehouses';
import materialsRouter from './routes/materials';
import issuesRouter from './routes/issues';
import allocationsRouter from './routes/allocations';
import ledgerRouter from './routes/ledger';

const app = express();
const PORT = process.env.PORT ? Number(process.env.PORT) : 8206;

app.use(express.json());

app.use('/api/warehouses', warehousesRouter);
app.use('/api/materials', materialsRouter);
app.use('/api/issues', issuesRouter);
app.use('/api/allocations', allocationsRouter);
app.use('/api/ledger', ledgerRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use((err: unknown, req: Request, res: Response, next: NextFunction) => {
  console.error(err);
  res.status(500).json({ message: 'Internal Server Error' });
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});
