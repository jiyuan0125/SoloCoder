import express from 'express';
import bodyParser from 'body-parser';
import sdkRoutes from './routes/sdks';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(bodyParser.json());

app.get('/health', (req, res) => {
  res.json({ status: 'OK', timestamp: new Date().toISOString() });
});

app.use('/sdks', sdkRoutes);

app.use((req, res) => {
  res.status(404).json({ error: 'Not Found' });
});

app.listen(PORT, () => {
  console.log(`SDK Version Manager API is running on port ${PORT}`);
});

export default app;
