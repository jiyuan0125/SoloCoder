import express from 'express';
import { initDatabase } from './database';
import gatewayRoutes from './routes/gateway';
import { routeMatcherMiddleware } from './middlewares/routeMatcher';
import { authMiddleware } from './middlewares/auth';
import { rateLimiterMiddleware } from './middlewares/rateLimiter';
import { proxyMiddleware } from './middlewares/proxy';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

initDatabase();

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/gateway', gatewayRoutes);

app.use(routeMatcherMiddleware);
app.use(authMiddleware);
app.use(rateLimiterMiddleware);
app.use(proxyMiddleware);

app.listen(PORT, () => {
  console.log(`API Gateway running on http://localhost:${PORT}`);
});
