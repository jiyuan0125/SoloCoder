import { Transaction, RiskAssessment } from '../types';
import { RuleEngine } from '../engine/ruleEngine';
import { AssessmentRepository } from '../repositories/assessmentRepository';
import { AccountRepository } from '../repositories/accountRepository';
import { ReviewRepository } from '../repositories/reviewRepository';
import { v4 as uuidv4 } from 'uuid';

export interface AssessmentResult {
  assessment: RiskAssessment;
  isNew: boolean;
}

export class RiskService {
  private static instance: RiskService;
  private ruleEngine: RuleEngine;
  private assessmentRepo: AssessmentRepository;
  private accountRepo: AccountRepository;
  private reviewRepo: ReviewRepository;

  static getInstance(): RiskService {
    if (!RiskService.instance) {
      RiskService.instance = new RiskService();
    }
    return RiskService.instance;
  }

  constructor() {
    this.ruleEngine = RuleEngine.getInstance();
    this.assessmentRepo = AssessmentRepository.getInstance();
    this.accountRepo = AccountRepository.getInstance();
    this.reviewRepo = ReviewRepository.getInstance();
  }

  assessTransaction(transaction: Transaction): AssessmentResult {
    const existing = this.assessmentRepo.findByTransactionId(transaction.id);
    if (existing) {
      return { assessment: existing, isNew: false };
    }

    const evaluation = this.ruleEngine.evaluateAll(transaction);
    const score = evaluation.totalScore;
    let status: RiskAssessment['status'];
    let accountFrozen = false;
    let frozenAttempted = false;
    let freezeFailed = false;

    if (score > 50) {
      status = 'BLOCKED';
      const freezeResult = this.tryFreezeAccount(transaction.userId);
      frozenAttempted = true;
      accountFrozen = freezeResult.success;
      freezeFailed = !freezeResult.success;
    } else if (score >= 30 && score <= 50) {
      status = 'PENDING_REVIEW';
      this.createPendingReview(transaction);
    } else {
      status = 'APPROVED';
    }

    const assessment = this.assessmentRepo.create({
      transactionId: transaction.id,
      score,
      status,
      matchedRuleIds: evaluation.matchedRules.map(r => r.id),
      accountFrozen,
      frozenAttempted,
      freezeFailed
    });

    return { assessment, isNew: true };
  }

  private tryFreezeAccount(userId: string): { success: boolean } {
    try {
      const account = this.accountRepo.findOrCreate(userId);
      if (account.frozen) {
        return { success: true };
      }
      this.accountRepo.freeze(account.id);
      return { success: true };
    } catch (e) {
      console.error(`Failed to freeze account ${userId}:`, e);
      return { success: false };
    }
  }

  private createPendingReview(transaction: Transaction): void {
    try {
      this.reviewRepo.create({
        assessmentId: uuidv4(),
        transactionId: transaction.id,
        decision: 'PENDING'
      });
    } catch (e) {
      console.error(`Failed to create review for transaction ${transaction.id}:`, e);
    }
  }

  retryFreezes(): number {
    const failedAssessments = this.assessmentRepo.findFailedFreezes();
    let successCount = 0;

    for (const assessment of failedAssessments) {
      try {
        const transactionId = assessment.transactionId;
        const transaction = this.getTransactionById(transactionId);

        if (transaction && transaction.userId) {
          const account = this.accountRepo.findByUserId(transaction.userId);
          if (account) {
            this.accountRepo.freeze(account.id);
            this.assessmentRepo.updateFreezeStatus(assessment.id, {
              accountFrozen: true,
              freezeFailed: false
            });
            successCount++;
          }
        }
      } catch (e) {
        console.error(`Failed to retry freeze for assessment ${assessment.id}:`, e);
      }
    }

    return successCount;
  }

  private getTransactionById(transactionId: string): Transaction | null {
    return null;
  }
}
