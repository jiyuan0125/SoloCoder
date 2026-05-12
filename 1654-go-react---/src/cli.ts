import * as readline from 'readline';

const BASE_URL = process.env.SESSION_SERVICE_URL || 'http://localhost:3000';

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

interface ApiClient {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, body?: unknown): Promise<T>;
}

function createApiClient(): ApiClient {
  async function request<T>(path: string, method: string, body?: unknown): Promise<T> {
    const url = BASE_URL + path;
    const init: RequestInit = {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    if (body) {
      init.body = JSON.stringify(body);
    }

    const response = await fetch(url, init);
    const text = await response.text();

    if (!response.ok) {
      throw new Error('HTTP ' + response.status + ': ' + (text || response.statusText));
    }

    return text ? (JSON.parse(text) as T) : (undefined as T);
  }

  return {
    get: <T>(path: string) => request<T>(path, 'GET'),
    post: <T>(path: string, body?: unknown) => request<T>(path, 'POST', body),
  };
}

const api = createApiClient();

function question(promptText: string): Promise<string> {
  return new Promise((resolve) => {
    rl.question(promptText, resolve);
  });
}

function printMenu(): void {
  console.log('\n=== Session Management CLI ===');
  console.log('1. Login (create session)');
  console.log('2. Send heartbeat');
  console.log('3. List online sessions');
  console.log('4. Kick session');
  console.log('5. Get statistics');
  console.log('6. Start auto heartbeat (every 30s)');
  console.log('0. Exit');
  console.log('');
}

async function handleLogin(): Promise<void> {
  console.log('\n--- Login ---');
  const userId = await question('User ID: ');
  const deviceInfo = await question('Device Info: ');
  const ipAddress = await question('IP Address: ');

  console.log('Package type: [free|basic|pro|enterprise] (default: basic)');
  const packageTypeInput = await question('Package type: ');
  const packageType = packageTypeInput || 'basic';

  try {
    const session = await api.post('/api/sessions/login', {
      userId,
      deviceInfo,
      ipAddress,
      packageType,
    });

    console.log('\nLogin success! Session:');
    console.log(JSON.stringify(session, null, 2));
  } catch (error) {
    console.error('Login failed:', (error as Error).message);
  }
}

async function handleHeartbeat(): Promise<void> {
  console.log('\n--- Heartbeat ---');
  const sessionId = await question('Session ID: ');

  try {
    const result = await api.post('/api/sessions/' + sessionId + '/heartbeat');
    console.log('\nHeartbeat success:', JSON.stringify(result, null, 2));
  } catch (error) {
    console.error('Heartbeat failed:', (error as Error).message);
  }
}

async function handleListSessions(): Promise<void> {
  console.log('\n--- List Sessions ---');
  const userIdFilter = await question('Filter by User ID (leave empty for no filter): ');
  const ipFilter = await question('Filter by IP Address (leave empty for no filter): ');

  const queryParams: string[] = [];
  if (userIdFilter) queryParams.push('userId=' + encodeURIComponent(userIdFilter));
  if (ipFilter) queryParams.push('ipAddress=' + encodeURIComponent(ipFilter));

  const path = queryParams.length > 0
    ? '/api/admin/sessions?' + queryParams.join('&')
    : '/api/admin/sessions';

  try {
    const sessions = await api.get(path);
    if (Array.isArray(sessions) && sessions.length === 0) {
      console.log('No online sessions');
    } else {
      console.log('\nOnline sessions:');
      console.log(JSON.stringify(sessions, null, 2));
    }
  } catch (error) {
    console.error('Failed to get sessions:', (error as Error).message);
  }
}

async function handleKickSession(): Promise<void> {
  console.log('\n--- Kick Session ---');
  const sessionId = await question('Session ID: ');
  const reason = await question('Reason (optional): ');

  try {
    const body: Record<string, string> = {};
    if (reason) body.reason = reason;

    const result = await api.post(
      '/api/admin/sessions/' + sessionId + '/kick',
      Object.keys(body).length > 0 ? body : undefined
    );

    console.log('\nResult:');
    console.log(JSON.stringify(result, null, 2));
  } catch (error) {
    console.error('Operation failed:', (error as Error).message);
  }
}

async function handleStatistics(): Promise<void> {
  console.log('\n--- Statistics ---');
  try {
    const stats = await api.get('/api/admin/statistics');
    console.log(JSON.stringify(stats, null, 2));
  } catch (error) {
    console.error('Failed to get statistics:', (error as Error).message);
  }
}

let activeInterval: NodeJS.Timeout | null = null;

async function handleAutoHeartbeat(): Promise<void> {
  console.log('\n--- Auto Heartbeat ---');
  const sessionId = await question('Session ID: ');

  if (activeInterval) {
    clearInterval(activeInterval);
    activeInterval = null;
  }

  console.log('\nAuto heartbeat started (every 30 seconds). Press Ctrl+C to stop.');

  activeInterval = setInterval(async () => {
    try {
      const result = await api.post('/api/sessions/' + sessionId + '/heartbeat');
      const now = new Date().toISOString();
      console.log('[' + now + '] Heartbeat OK');
    } catch (error) {
      console.error('[' + new Date().toISOString() + '] Heartbeat FAILED:', (error as Error).message);
      if (activeInterval) {
        clearInterval(activeInterval);
        activeInterval = null;
      }
      console.log('Auto heartbeat stopped.');
    }
  }, 30 * 1000);
}

async function main(): Promise<void> {
  console.log('Session Management Service - CLI Client');
  console.log('Service URL: ' + BASE_URL);

  while (true) {
    printMenu();
    const choice = await question('Select operation (0-6): ');

    switch (choice.trim()) {
      case '1':
        await handleLogin();
        break;
      case '2':
        await handleHeartbeat();
        break;
      case '3':
        await handleListSessions();
        break;
      case '4':
        await handleKickSession();
        break;
      case '5':
        await handleStatistics();
        break;
      case '6':
        await handleAutoHeartbeat();
        break;
      case '0':
        console.log('Goodbye!');
        if (activeInterval) clearInterval(activeInterval);
        rl.close();
        return;
      default:
        console.log('Invalid choice, please try again.');
    }
  }
}

if (require.main === module) {
  main().catch((err) => {
    console.error('Error:', err);
    process.exit(1);
  });
}
