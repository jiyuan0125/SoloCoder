import express from 'express';
import applicationsRouter from './routes/applications';
import alertsRouter from './routes/alerts';
import topologyRouter from './routes/topology';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json({ limit: '10mb' }));

app.use('/applications', applicationsRouter);
app.use('/alerts', alertsRouter);
app.use('/topology', topologyRouter);

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`APM Service is running on port ${PORT}`);
});
