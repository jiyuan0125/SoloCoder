import { v4 as uuidv4 } from 'uuid';
import {
  Plan,
  CreatePlanRequest,
  UpdatePlanRequest
} from '../types';
import { planRepository, errorRepository, db } from '../db';
import {
  getAllowedActions,
  canTransition,
  getNextStatus,
  ActionType,
  getActionDescription
} from '../utils/stateMachine';
import { validateStrategy } from '../utils/grayStrategy';

const STATUS_PENDING = 'pending';
const STATUS_GRAYSCALE = 'grayscale';
const STATUS_FULL_RELEASE = 'full_release';
const STATUS_PAUSED = 'paused';
const STATUS_COMPLETED = 'completed';
const STATUS_ROLLED_BACK = 'rolled_back';

export class PlanService {
  createPlan(request: CreatePlanRequest): Plan {
    this.validateCreateRequest(request);

    const existingGrayscalePlan = planRepository.findByAppIdAndStatus(
      request.appId,
      STATUS_GRAYSCALE as any
    );
    if (existingGrayscalePlan) {
      throw this.createConflictError(
        `App ${request.appId} 已有处于灰度中的发布计划`
      );
    }

    const now = Date.now();
    const plan: Plan = {
      id: uuidv4(),
      name: request.name,
      appId: request.appId,
      oldVersion: request.oldVersion,
      newVersion: request.newVersion,
      strategy: request.strategy,
      status: STATUS_PENDING as any,
      errorRateConfig: request.errorRateConfig || {
        threshold: 5,
        windowMinutes: 5
      },
      createdAt: now,
      updatedAt: now
    };

    planRepository.create(plan);
    return plan;
  }

  getPlanById(id: string): Plan {
    const plan = planRepository.findById(id);
    if (!plan) {
      throw this.createNotFoundError(`发布计划 ${id} 不存在`);
    }
    return plan;
  }

  getAllPlans(): Plan[] {
    return planRepository.findAll();
  }

  updatePlan(id: string, request: UpdatePlanRequest): Plan {
    const plan = this.getPlanById(id);

    if (plan.status !== STATUS_PENDING) {
      throw this.createBadRequestError(
        '只有待开始状态的计划可以修改',
        plan.status
      );
    }

    if (request.strategy) {
      const strategyValidation = validateStrategy(request.strategy);
      if (!strategyValidation.valid) {
        throw this.createBadRequestError(
          strategyValidation.error!,
          plan.status
        );
      }
    }

    if (request.name) {
      plan.name = request.name;
    }
    if (request.strategy) {
      plan.strategy = request.strategy;
    }
    if (request.errorRateConfig) {
      plan.errorRateConfig = request.errorRateConfig;
    }

    plan.updatedAt = Date.now();
    planRepository.update(plan);
    return plan;
  }

  deletePlan(id: string): void {
    const plan = this.getPlanById(id);

    if (plan.status === STATUS_GRAYSCALE || plan.status === STATUS_FULL_RELEASE) {
      throw this.createConflictError(
        `处于 ${plan.status} 状态的发布计划不能删除`
      );
    }

    planRepository.delete(id);
  }

  startGrayscale(id: string): Plan {
    return this.transitionState(id, ActionType.START_GRAYSCALE);
  }

  fullRelease(id: string): Plan {
    return this.transitionState(id, ActionType.FULL_RELEASE);
  }

  complete(id: string): Plan {
    return this.transitionState(id, ActionType.COMPLETE);
  }

  pause(id: string, reason?: string): Plan {
    return this.transitionState(id, ActionType.PAUSE);
  }

  resume(id: string): Plan {
    return this.transitionState(id, ActionType.RESUME);
  }

  rollback(id: string): Plan {
    const plan = this.getPlanById(id);

    if (plan.status !== STATUS_FULL_RELEASE) {
      throw this.createBadRequestError(
        '只有全量发布状态的计划可以执行回滚操作',
        plan.status
      );
    }

    const transaction = db.transaction(() => {
      plan.status = STATUS_ROLLED_BACK as any;
      plan.updatedAt = Date.now();
      planRepository.update(plan);
    });

    transaction();
    return plan;
  }

  private transitionState(id: string, action: ActionType): Plan {
    const plan = this.getPlanById(id);

    if (!canTransition(plan.status, action)) {
      throw this.createBadRequestError(
        `无法执行 ${getActionDescription(action)} 操作`,
        plan.status
      );
    }

    const nextStatus = getNextStatus(plan.status, action);
    if (!nextStatus) {
      throw this.createBadRequestError(
        '无效的状态转换',
        plan.status
      );
    }

    plan.status = nextStatus as any;
    plan.updatedAt = Date.now();
    planRepository.update(plan);
    return plan;
  }

  recordErrorAndCheckThreshold(planId: string, isError: boolean): {
    plan: Plan;
    errorRate: number;
    autoPaused: boolean;
  } {
    const now = Date.now();
    const transaction = db.transaction(() => {
      const plan = this.getPlanById(planId);

      const errorRecord = {
        id: uuidv4(),
        planId,
        timestamp: now,
        isError
      };
      errorRepository.create(errorRecord);

      const windowStart = now - (plan.errorRateConfig.windowMinutes * 60 * 1000);
      const errorStats = errorRepository.calculateErrorRate(planId, windowStart, now);

      let autoPaused = false;
      if (plan.status === STATUS_GRAYSCALE && errorStats.rate >= plan.errorRateConfig.threshold) {
        plan.status = STATUS_PAUSED as any;
        plan.updatedAt = now;
        planRepository.update(plan);
        autoPaused = true;
      }

      return { plan, errorRate: errorStats.rate, autoPaused };
    });

    return transaction();
  }

  getErrorRate(id: string): {
    total: number;
    errors: number;
    rate: number;
    threshold: number;
    windowMinutes: number;
  } {
    const plan = this.getPlanById(id);
    const now = Date.now();
    const windowStart = now - (plan.errorRateConfig.windowMinutes * 60 * 1000);

    const stats = errorRepository.calculateErrorRate(id, windowStart, now);

    return {
      ...stats,
      threshold: plan.errorRateConfig.threshold,
      windowMinutes: plan.errorRateConfig.windowMinutes
    };
  }

  private validateCreateRequest(request: CreatePlanRequest): void {
    const requiredFields = ['name', 'appId', 'oldVersion', 'newVersion', 'strategy'];
    const missingFields = requiredFields.filter(field => !(request as any)[field]);
    if (missingFields.length > 0) {
      throw this.createBadRequestError(
        `缺少必填字段: ${missingFields.join(', ')}`,
        STATUS_PENDING as any
      );
    }

    const strategyValidation = validateStrategy(request.strategy);
    if (!strategyValidation.valid) {
      throw this.createBadRequestError(
        strategyValidation.error!,
        STATUS_PENDING as any
      );
    }
  }

  private createBadRequestError(message: string, currentStatus: string): any {
    return {
      status: 400,
      message,
      allowedActions: getAllowedActions(currentStatus).map(a => getActionDescription(a))
    };
  }

  private createNotFoundError(message: string): any {
    return { status: 404, message };
  }

  private createConflictError(message: string): any {
    return { status: 409, message };
  }
}

export const planService = new PlanService();
