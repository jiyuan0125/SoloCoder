import cron from 'node-cron';
import db from '../db';
import {
  getPendingTasks,
  markTaskRunning,
  markTaskCompleted,
  updateCommunitySendResult,
  getTaskTargetCommunities
} from './broadcast';
import { getCommunityById, getCommunityMembers } from './community';
import { getUserTier, getUsersByTier, getUsersByTags } from './userTag';
import { SendResult, UserTier } from '../types';

const MAX_SENDS_PER_MINUTE = 30;
let sendTimestamps: number[] = [];

class RateLimiter {
  private queue: (() => Promise<void>)[] = [];
  private processing = false;

  async acquire(): Promise<void> {
    return new Promise((resolve) => {
      this.queue.push(async () => {
        await this.waitForSlot();
        resolve();
      });
      if (!this.processing) {
        this.processQueue();
      }
    });
  }

  private async waitForSlot(): Promise<void> {
    const now = Date.now();
    const oneMinuteAgo = now - 60000;

    sendTimestamps = sendTimestamps.filter(t => t > oneMinuteAgo);

    while (sendTimestamps.length >= MAX_SENDS_PER_MINUTE) {
      const oldestTimestamp = sendTimestamps[0];
      const waitTime = oldestTimestamp + 60000 - now;
      await new Promise(r => setTimeout(r, Math.max(waitTime, 100)));

      const newNow = Date.now();
      const newOneMinuteAgo = newNow - 60000;
      sendTimestamps = sendTimestamps.filter(t => t > newOneMinuteAgo);
    }

    sendTimestamps.push(Date.now());
  }

  private async processQueue(): Promise<void> {
    this.processing = true;
    while (this.queue.length > 0) {
      const next = this.queue.shift();
      if (next) {
        await next();
      }
    }
    this.processing = false;
  }
}

const rateLimiter = new RateLimiter();

async function simulateSendToCommunity(
  taskId: string,
  communityId: string,
  content: string,
  tierFilter?: UserTier,
  tagFilterIds?: string[],
  tagFilterMode?: 'AND' | 'OR'
): Promise<{ success: boolean; failureReason?: string }> {
  const community = getCommunityById(communityId);

  if (!community) {
    return { success: false, failureReason: '目标不存在' };
  }

  await rateLimiter.acquire();

  try {
    const allMembers = getCommunityMembers(communityId);
    let filteredMembers = allMembers;

    if (tierFilter) {
      const tierUsers = getUsersByTier(tierFilter);
      const tierUserIds = new Set(tierUsers.map(u => u.id));
      filteredMembers = filteredMembers.filter(m => tierUserIds.has(m));
    }

    if (tagFilterIds && tagFilterIds.length > 0) {
      const taggedUsers = getUsersByTags(tagFilterIds, tagFilterMode || 'OR');
      const taggedUserIds = new Set(taggedUsers.map(u => u.id));
      filteredMembers = filteredMembers.filter(m => taggedUserIds.has(m));
    }

    if (filteredMembers.length === 0) {
      return { success: true };
    }

    console.log(`[Task ${taskId}] Sending to community ${communityId}: "${content.slice(0, 50)}..."`);

    return { success: true };
  } catch (e: any) {
    return { success: false, failureReason: e.message || 'Unknown error' };
  }
}

async function executeTask(taskId: string): Promise<void> {
  if (!markTaskRunning(taskId)) {
    return;
  }

  try {
    const taskRow = db.prepare('SELECT * FROM broadcast_tasks WHERE id = ?').get(taskId) as any;
    if (!taskRow) return;

    const communityIds = getTaskTargetCommunities(taskId);

    const tagFilterRows = db.prepare(
      'SELECT tag_id FROM task_tag_filters WHERE task_id = ?'
    ).all(taskId) as any[];
    const tagFilterIds = tagFilterRows.map(r => r.tag_id);

    for (const communityId of communityIds) {
      const result = await simulateSendToCommunity(
        taskId,
        communityId,
        taskRow.content,
        taskRow.tier_filter || undefined,
        tagFilterIds,
        taskRow.tag_filter_mode || undefined
      );

      if (result.success) {
        updateCommunitySendResult(taskId, communityId, SendResult.SUCCESS);
      } else if (result.failureReason === '目标不存在') {
        updateCommunitySendResult(taskId, communityId, SendResult.SKIPPED, result.failureReason);
      } else {
        updateCommunitySendResult(taskId, communityId, SendResult.FAILED, result.failureReason);
      }
    }

    markTaskCompleted(taskId);
  } catch (e: any) {
    console.error(`[Task ${taskId}] Execution error:`, e);
  }
}

let isRunning = false;

async function tick() {
  if (isRunning) return;
  isRunning = true;

  try {
    const tasks = getPendingTasks();
    for (const task of tasks) {
      await executeTask(task.id);
    }
  } finally {
    isRunning = false;
  }
}

export function startBroadcastEngine() {
  cron.schedule('* * * * * *', tick);
  console.log('Broadcast engine started (checking every second)');
}
