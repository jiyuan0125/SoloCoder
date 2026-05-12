import { Content, ContentStatus, ReviewResult } from '../types';
import { contentStore } from '../store/contentStore';
import { reviewLogStore } from '../store/reviewLogStore';
import { sensitiveWordService } from './sensitiveWordService';
import { ruleEngine } from './ruleEngine';
import { imageReviewService } from './imageReviewService';
import { reviewQueue } from './reviewQueue';

interface AutoReviewResult {
  success: boolean;
  error?: string;
  statusCode?: number;
  content?: Content;
}

interface ManualReviewResult {
  success: boolean;
  error?: string;
  statusCode?: number;
  content?: Content;
}

interface FinalizeResult {
  success: boolean;
  error?: string;
  statusCode?: number;
  content?: Content;
}

export class ContentService {
  create(text?: string, imageUrl?: string): Content {
    const content = contentStore.create(text, imageUrl);
    reviewLogStore.add(content.id, 'content_submitted', {
      text,
      imageUrl,
    });
    return content;
  }

  findById(id: string): Content | undefined {
    return contentStore.findById(id);
  }

  findAll(): Content[] {
    return contentStore.findAll();
  }

  autoReview(id: string): AutoReviewResult {
    const content = contentStore.findById(id);
    if (!content) {
      return { success: false, error: '内容不存在', statusCode: 404 };
    }

    if (content.status !== ContentStatus.SUBMITTED) {
      return { 
        success: false, 
        error: `非法状态跳转，当前状态: ${content.status}`, 
        statusCode: 400 
      };
    }

    const reviewDetails: Record<string, unknown> = {};
    let finalResult: ReviewResult = ReviewResult.PASS;
    let needsManualReview = false;

    const matchedWords = content.text ? sensitiveWordService.matchText(content.text) : [];
    const ruleResult = ruleEngine.evaluate(content, matchedWords);

    if (matchedWords.length > 0) {
      reviewDetails.textMatchedWords = matchedWords.map(m => ({
        word: m.word.word,
        level: m.word.level,
      }));
    }

    if (ruleResult.level1Hits.length > 0) {
      reviewDetails.level1Hits = ruleResult.level1Hits.map(w => w.word);
      finalResult = ReviewResult.REJECT;
    } else if (ruleResult.level2Hits.length > 0) {
      reviewDetails.level2Hits = ruleResult.level2Hits.map(w => w.word);
      needsManualReview = true;
      finalResult = ReviewResult.PENDING;
    } else if (ruleResult.level3Hits.length > 0) {
      reviewDetails.level3Hits = ruleResult.level3Hits.map(w => w.word);
    }

    if (ruleResult.action === 'REJECT') {
      finalResult = ReviewResult.REJECT;
      reviewDetails.ruleAction = 'REJECT';
      if (ruleResult.matchedRule) {
        reviewDetails.matchedRule = ruleResult.matchedRule.name;
      }
    } else if (ruleResult.action === 'PENDING') {
      needsManualReview = true;
      finalResult = ReviewResult.PENDING;
      reviewDetails.ruleAction = 'PENDING';
      if (ruleResult.matchedRule) {
        reviewDetails.matchedRule = ruleResult.matchedRule.name;
      }
    }

    if (content.imageUrl) {
      const imageResult = imageReviewService.review(content.imageUrl);
      reviewDetails.imageReview = imageResult;

      if (imageResult.result === ReviewResult.REJECT) {
        finalResult = ReviewResult.REJECT;
      } else if (imageResult.result === ReviewResult.PENDING) {
        if (finalResult !== ReviewResult.REJECT) {
          needsManualReview = true;
          finalResult = ReviewResult.PENDING;
        }
      }
    }

    const updatedContent = contentStore.update(id, {
      status: ContentStatus.AUTO_REVIEWED,
      autoReviewResult: finalResult,
      autoReviewedAt: new Date(),
    });

    reviewLogStore.add(id, 'auto_review_completed', {
      result: finalResult,
      ...reviewDetails,
    });

    if (needsManualReview) {
      reviewLogStore.add(id, 'queue_push_attempted', {});
      
      const pushed = reviewQueue.push(id);
      
      if (!pushed) {
        reviewLogStore.add(id, 'queue_push_failed', {
          reason: '复审队列暂不可用',
        });
        return {
          success: false,
          error: '复审队列暂不可用',
          statusCode: 503,
          content: updatedContent,
        };
      } else {
        reviewLogStore.add(id, 'queue_push_success', {});
      }
    }

    return { success: true, content: updatedContent };
  }

  manualReview(id: string, result: ReviewResult): ManualReviewResult {
    const content = contentStore.findById(id);
    if (!content) {
      return { success: false, error: '内容不存在', statusCode: 404 };
    }

    if (content.status !== ContentStatus.AUTO_REVIEWED) {
      return { 
        success: false, 
        error: `非法状态跳转，当前状态: ${content.status}`, 
        statusCode: 400 
      };
    }

    const updatedContent = contentStore.update(id, {
      status: ContentStatus.MANUAL_REVIEWED,
      manualReviewResult: result,
      manualReviewedAt: new Date(),
    });

    reviewLogStore.add(id, 'manual_review_completed', {
      result,
    });

    return { success: true, content: updatedContent };
  }

  finalize(id: string, result: ReviewResult): FinalizeResult {
    const content = contentStore.findById(id);
    if (!content) {
      return { success: false, error: '内容不存在', statusCode: 404 };
    }

    if (content.status !== ContentStatus.MANUAL_REVIEWED) {
      return { 
        success: false, 
        error: `非法状态跳转，当前状态: ${content.status}`, 
        statusCode: 400 
      };
    }

    const updatedContent = contentStore.update(id, {
      status: ContentStatus.FINALIZED,
      finalResult: result,
      finalizedAt: new Date(),
    });

    reviewLogStore.add(id, 'finalized', {
      result,
    });

    return { success: true, content: updatedContent };
  }
}

export const contentService = new ContentService();
