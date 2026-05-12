import { ReviewResult } from '../types';

export class ImageReviewService {
  review(imageUrl: string): { result: ReviewResult; reason: string } {
    const hash = this.simpleHash(imageUrl);
    const probability = (hash % 100) / 100;

    if (probability < 0.05) {
      return { result: ReviewResult.REJECT, reason: '图片违规' };
    } else if (probability < 0.35) {
      return { result: ReviewResult.PENDING, reason: '图片需要人工复审' };
    }

    return { result: ReviewResult.PASS, reason: '图片通过' };
  }

  private simpleHash(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash;
    }
    return Math.abs(hash);
  }
}

export const imageReviewService = new ImageReviewService();
