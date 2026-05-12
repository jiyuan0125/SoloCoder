import express from 'express';
import categoriesRouter from './routes/categories';
import faqsRouter from './routes/faqs';
import todosRouter from './routes/todos';
import statsRouter from './routes/stats';

const app = express();
const port = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.use('/categories', categoriesRouter);
app.use('/faqs', faqsRouter);
app.use('/todos', todosRouter);
app.use('/stats', statsRouter);

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
