import express from 'express';
import servicesRouter from './routes/services';
import incidentsRouter from './routes/incidents';
import reportsRouter from './routes/reports';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/services', servicesRouter);
app.use('/incidents', incidentsRouter);
app.use('/reports', reportsRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`SLA Management System running on port ${PORT}`);
});

export default app;
