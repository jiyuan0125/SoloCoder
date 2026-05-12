import { ReviewRecord } from '../types';
import { ReviewRepository } from '../repositories/reviewRepository';
import { AssessmentRepository } from '../repositories/assessmentRepository';

export class ReviewNotFoundException extends Error {
  constructor() {
    super('Review record not found');
    this.name = 'ReviewNotFoundException';
  }
}

export class ReviewService {
  private static instance: ReviewService;
  private reviewRepo: ReviewRepository;
  private assessmentRepo: AssessmentRepository;

  static getInstance(): ReviewService {
    if (!ReviewService.instance) {
      ReviewService.instance = new ReviewService();
    }
    return ReviewService.instance;
  }

  constructor() {
    this.reviewRepo = ReviewRepository.getInstance();
    this.assessmentRepo = AssessmentRepository.getInstance();
  }

  findById(id: string): ReviewRecord | null {
    return this.reviewRepo.findById(id);
  }

  findPending(): ReviewRecord[] {
    return this.reviewRepo.findPending();
  }

  findByAssessmentId(assessmentId: string): ReviewRecord | null {
    return this.reviewRepo.findByAssessmentId(assessmentId);
  }

  decide(
    id: string,
    decision: 'NORMAL' | 'FRAUD',
    reviewerId?: string
  ): ReviewRecord | null {
    const existing = this.reviewRepo.findById(id);
    if (!existing) {
      throw new ReviewNotFoundException();
    }

    return this.reviewRepo.decide(id, decision, reviewerId);
  }

  getReviewWithAssessment(id: string) {
    const review = this.reviewRepo.findById(id);
    if (!review) {
      throw new ReviewNotFoundException();
    }

    const assessment = this.assessmentRepo.findById(review.assessmentId);
    return { review, assessment };
  }
}
