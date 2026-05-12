import { validate } from 'node-cron';

export interface CronValidationError {
  valid: boolean;
  message?: string;
  position?: number;
}

export function validateCron(cronExpression: string): CronValidationError {
  try {
    const isValid = validate(cronExpression);
    
    if (!isValid) {
      return {
        valid: false,
        message: `Cron 表达式格式无效: ${cronExpression}`,
        position: 0
      };
    }
    
    return { valid: true };
  } catch (error: any) {
    let message = `Cron 表达式解析失败: ${error.message}`;
    let position = 0;
    
    if (error.message.includes('at position')) {
      const match = error.message.match(/at position (\d+)/);
      if (match) {
        position = parseInt(match[1], 10);
      }
    }
    
    return {
      valid: false,
      message,
      position
    };
  }
}
