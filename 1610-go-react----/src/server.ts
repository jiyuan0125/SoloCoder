import express from 'express';
import { templatesRouter } from './routes/templates';
import { channelsRouter } from './routes/channels';
import { notificationsRouter } from './routes/notifications';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 9100;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
});

app.use('/templates', templatesRouter);
app.use('/channels', channelsRouter);
app.use('/notifications', notificationsRouter);

app.listen(PORT, () => {
  console.log(`Notification platform server listening on port ${PORT}`);
});

export { app };
