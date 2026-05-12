import express from 'express';
import serversRouter from './routes/servers';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/servers', serversRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Capacity Management Service is running on port ${PORT}`);
  });
}

export default app;
