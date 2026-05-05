import * as http from 'http';
import * as https from 'https';
import * as url from 'url';

export interface HttpClientResponse {
  statusCode: number;
  body: unknown;
}

export async function sendRequest(
  method: 'GET' | 'POST',
  serverUrl: string,
  path: string,
  queryParams?: Record<string, string | number | boolean | undefined>,
  body?: unknown
): Promise<HttpClientResponse> {
  const parsedUrl = url.parse(serverUrl);
  const protocol = parsedUrl.protocol === 'https:' ? https : http;
  
  let fullPath = path;
  if (queryParams) {
    const params = new url.URLSearchParams();
    for (const [key, value] of Object.entries(queryParams)) {
      if (value !== undefined) {
        params.append(key, String(value));
      }
    }
    const queryString = params.toString();
    if (queryString) {
      fullPath += '?' + queryString;
    }
  }

  const options: http.RequestOptions = {
    hostname: parsedUrl.hostname,
    port: parsedUrl.port ? parseInt(parsedUrl.port, 10) : undefined,
    path: fullPath,
    method,
    headers: body ? {
      'Content-Type': 'application/json',
    } : {},
  };

  return new Promise((resolve, reject) => {
    const req = protocol.request(options, (res) => {
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        try {
          const responseBody = data.trim() ? JSON.parse(data) : undefined;
          resolve({
            statusCode: res.statusCode || 500,
            body: responseBody,
          });
        } catch (error) {
          reject(new Error(`Failed to parse response: ${data}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    if (body) {
      req.write(JSON.stringify(body));
    }

    req.end();
  });
}
