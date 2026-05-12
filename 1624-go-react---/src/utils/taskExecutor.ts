import { TaskParams } from '../models/task';

export async function executeTask(params: TaskParams): Promise<any> {
  console.log(`[TaskExecutor] 执行任务参数:`, params);
  
  const executionType = params.executionType || 'simulated';
  
  switch (executionType) {
    case 'http':
      return await executeHttpTask(params);
    case 'shell':
      return await executeShellTask(params);
    default:
      return await executeSimulatedTask(params);
  }
}

async function executeSimulatedTask(params: TaskParams): Promise<any> {
  const duration = params.simulatedDuration || 1000;
  const shouldFail = params.shouldFail === true;
  
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (shouldFail) {
        reject(new Error('模拟任务执行失败'));
      } else {
        resolve({ 
          success: true, 
          message: '模拟任务执行成功',
          executedAt: new Date().toISOString(),
          params
        });
      }
    }, duration);
  });
}

async function executeHttpTask(params: TaskParams): Promise<any> {
  if (!params.url) {
    throw new Error('HTTP 任务需要指定 url 参数');
  }
  
  const options = {
    method: params.method || 'GET',
    headers: params.headers || {},
    body: params.body
  };
  
  const response = await fetch(params.url, {
    method: options.method,
    headers: options.headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  });
  
  const result = {
    status: response.status,
    statusText: response.statusText,
    data: response.ok ? await response.json() : null
  };
  
  if (!response.ok) {
    throw new Error(`HTTP 请求失败: ${response.status} ${response.statusText}`);
  }
  
  return result;
}

async function executeShellTask(params: TaskParams): Promise<any> {
  if (!params.command) {
    throw new Error('Shell 任务需要指定 command 参数');
  }
  
  return {
    success: true,
    message: 'Shell 命令执行模拟完成（实际执行需要 child_process）',
    command: params.command
  };
}
