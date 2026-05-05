import * as http from 'http';
import { handleRequest } from './handlers';

const PORT = parseInt(process.env.PORT || '3000', 10);
const HOST = process.env.HOST || 'localhost';

const server = http.createServer(async (req, res) => {
  await handleRequest(req, res);
});

server.on('error', (error) => {
  console.error('Server error:', error);
});

server.listen(PORT, HOST, () => {
  console.log(`Coupon Distributor Server running at http://${HOST}:${PORT}/`);
  console.log(`Health check: http://${HOST}:${PORT}/health`);
  console.log('API endpoints:');
  console.log('  POST /api/coupons/distribute  - Distribute coupons');
  console.log('  POST /api/coupons/redeem      - Redeem coupons');
  console.log('  POST /api/coupons/query       - Query user coupons');
  console.log('  POST /api/coupons/recommend   - Get recommended coupons');
});

process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down server...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down server...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});
