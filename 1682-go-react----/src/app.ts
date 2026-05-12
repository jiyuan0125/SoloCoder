import express from 'express';
import bodyParser from 'body-parser';
import proposalsRouter from './routes/proposals';
import votesRouter from './routes/votes';
import todosRouter from './routes/todos';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(bodyParser.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/api/proposals', proposalsRouter);
app.use('/api', votesRouter);
app.use('/api/todos', todosRouter);

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal server error' });
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
  });
}

export default app;
