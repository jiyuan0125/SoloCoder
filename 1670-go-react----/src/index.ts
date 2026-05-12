import express from 'express';
import { initDB } from './db';
import sellersRoutes from './routes/sellers';
import transactionsRoutes from './routes/transactions';
import settlementsRoutes from './routes/settlements';
import feesRoutes from './routes/fees';

initDB();

const app = express();
const PORT = process.env.PORT || 8100;

app.use(express.json());

app.get('/', (req, res) => {
  res.json({ message: 'Settlement System API' });
});

app.use('/sellers', sellersRoutes);
app.use('/transactions', transactionsRoutes);
app.use('/settlements', settlementsRoutes);
app.use('/fees', feesRoutes);

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});
