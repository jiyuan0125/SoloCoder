import { v4 as uuidv4 } from 'uuid';
import { ReviewLog } from '../types';

const MAX_LOGS = 10000;

class ReviewLogStore {
  private logs: ReviewLog[] = [];

  add(contentId: string, action: string, details: Record<string, unknown>): ReviewLog {
    const log: ReviewLog = {
      id: uuidv4(),
      contentId,
      action,
      details,
      timestamp: new Date(),
    };
    
    this.logs.push(log);
    
    if (this.logs.length > MAX_LOGS) {
      this.logs.shift();
    }
    
    return log;
  }

  findByContentId(contentId: string): ReviewLog[] {
    return this.logs.filter(log => log.contentId === contentId);
  }

  findAll(): ReviewLog[] {
    return [...this.logs];
  }
}

export const reviewLogStore = new ReviewLogStore();
