import { checkAndRecordRateLimit } from './rateLimiter';
import { analyzeContent } from './contentAnalyzer';
import {
  isUserBlacklisted,
  isUserWhitelisted,
  ensureUserExists,
  getUser,
  getBlacklistEntry,
  setWhitelist,
  markAsBlacklist,
  removeFromBlacklist,
} from './userManager';
import { createPost, getPost, updatePostStatus, getAuditQueue, getAuditQueueCount } from './contentManager';
import type { ContentType, Post, WindowType, MarkBlacklistResult } from '../types';

export interface SubmitContentResult {
  success: boolean;
  code: number;
  message: string;
  post?: Post;
  exceededDimension?: WindowType;
  reasons?: string[];
}

export function submitContent(userId: string, content: string, contentType: ContentType): SubmitContentResult {
  ensureUserExists(userId);

  if (isUserBlacklisted(userId)) {
    return { success: false, code: 403, message: 'user_is_blacklisted' };
  }

  const isWhitelisted = isUserWhitelisted(userId);

  if (!isWhitelisted) {
    const rateLimitResult = checkAndRecordRateLimit(userId, contentType);
    if (!rateLimitResult.allowed && rateLimitResult.exceededDimension) {
      const post = createPost(userId, content, contentType, 'pending_review');
      return {
        success: false,
        code: 429,
        message: 'rate_limit_exceeded',
        exceededDimension: rateLimitResult.exceededDimension,
        post,
      };
    }
  }

  const analysis = analyzeContent(content);

  if (isWhitelisted) {
    const post = createPost(userId, content, contentType, 'published');
    return { success: true, code: 200, message: 'published', post };
  }

  if (analysis.isSuspicious) {
    const post = createPost(userId, content, contentType, 'suspicious');
    return {
      success: false,
      code: 202,
      message: 'content_suspicious',
      reasons: analysis.reasons,
      post,
    };
  }

  const post = createPost(userId, content, contentType, 'published');
  return { success: true, code: 200, message: 'published', post };
}

export {
  getUser,
  getBlacklistEntry,
  setWhitelist,
  markAsBlacklist,
  removeFromBlacklist,
  getPost,
  updatePostStatus,
  getAuditQueue,
  getAuditQueueCount,
};

export type { MarkBlacklistResult };