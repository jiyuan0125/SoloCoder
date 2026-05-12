const http = require('http');

function request(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'localhost',
      port: 3001,
      path,
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({
            status: res.statusCode,
            data: data ? JSON.parse(data) : null,
          });
        } catch (e) {
          resolve({ status: res.statusCode, data: data });
        }
      });
    });

    req.on('error', reject);

    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

async function main() {
  console.log('=== Testing clone issue ===');
  
  const createRes = await request('POST', '/environments', {
    name: 'test-env',
    type: 'dev',
    projectId: 'proj-1',
    resources: {
      serviceInstanceCount: 2,
      database: {
        host: 'localhost',
        port: 5432,
        name: 'test',
        username: 'user',
        password: 'pass'
      },
      middleware: {}
    },
    config: {
      LOG_LEVEL: 'debug'
    }
  });
  
  console.log('Create response:', createRes.status, createRes.data);
  
  const envId = createRes.data.id;
  console.log('Created env id:', envId);

  await request('POST', `/environments/${envId}/status`, { status: 'initializing' });
  const statusRes = await request('POST', `/environments/${envId}/status`, { status: 'running' });
  console.log('Status update response:', statusRes.status, statusRes.data);

  console.log('\n--- Trying to clone while running ---');
  const cloneRes = await request('POST', `/environments/${envId}/clone`, { name: 'cloned-env' });
  console.log('Clone response:', cloneRes.status, cloneRes.data);

  console.log('\n=== Testing quota issue ===');
  for (let i = 0; i < 4; i++) {
    const res = await request('POST', '/environments', {
      name: `dev-env-${i}`,
      type: 'dev',
      projectId: 'proj-2',
      resources: {
        serviceInstanceCount: 1,
        database: {
          host: 'localhost',
          port: 5432,
          name: 'test',
          username: 'user',
          password: 'pass'
        },
        middleware: {}
      },
      config: {}
    });
    console.log(`Create ${i}:`, res.status, res.data);
  }
}

main().catch(console.error);
