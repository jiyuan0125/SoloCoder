import * as http from 'http';
import { handleRequest } from './router';

export function createServer(): http.Server {
  const server = http.createServer(async (req: http.IncomingMessage, res: http.ServerResponse) => {
    try {
      const result = await handleRequest(req);
      
      res.writeHead(result.statusCode, {
        'Content-Type': 'application/json',
      });
      
      res.end(JSON.stringify(result.body));
    } catch (error) {
      console.error('Server error:', error);
      
      res.writeHead(500, {
        'Content-Type': 'application/json',
      });
      
      res.end(JSON.stringify({
        success: false,
        error: {
          code: 'INTERNAL_ERROR',
          message: 'An unexpected error occurred',
        },
      }));
    }
  });

  return server;
}
