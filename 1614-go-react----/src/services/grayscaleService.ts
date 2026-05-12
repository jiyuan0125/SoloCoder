import { shouldUseNewVersion } from '../utils/grayStrategy';
import { planService } from './planService';
import { PlanStatus } from '../types';

export class GrayscaleService {
  resolveVersion(
    planId: string,
    userId: string,
    userAttributes: Record<string, any> = {}
  ): {
    planId: string;
    version: string;
    status: PlanStatus;
    inGrayscale: boolean;
  } {
    const plan = planService.getPlanById(planId);
    let version = plan.oldVersion;
    let inGrayscale = false;

    switch (plan.status) {
      case PlanStatus.PENDING:
        version = plan.oldVersion;
        break;
      case PlanStatus.GRAYSCALE:
        if (shouldUseNewVersion(planId, userId, plan.strategy, userAttributes)) {
          version = plan.newVersion;
          inGrayscale = true;
        } else {
          version = plan.oldVersion;
        }
        break;
      case PlanStatus.FULL_RELEASE:
      case PlanStatus.COMPLETED:
        version = plan.newVersion;
        break;
      case PlanStatus.PAUSED:
      case PlanStatus.ROLLED_BACK:
        version = plan.oldVersion;
        break;
    }

    return {
      planId,
      version,
      status: plan.status,
      inGrayscale
    };
  }

  resolveVersionByAppId(
    appId: string,
    userId: string,
    userAttributes: Record<string, any> = {}
  ): {
    appId: string;
    planId: string | null;
    version: string;
    status: PlanStatus | null;
    inGrayscale: boolean;
  } {
    const allPlans = planService.getAllPlans();
    const appPlans = allPlans.filter(p => p.appId === appId);

    const activePlan = appPlans.find(
      p => p.status === PlanStatus.GRAYSCALE ||
           p.status === PlanStatus.FULL_RELEASE ||
           p.status === PlanStatus.PAUSED
    );

    if (!activePlan) {
      const completedOrRolledBack = appPlans.find(
        p => p.status === PlanStatus.COMPLETED || p.status === PlanStatus.ROLLED_BACK
      );
      if (completedOrRolledBack) {
        return {
          appId,
          planId: completedOrRolledBack.id,
          version: completedOrRolledBack.status === PlanStatus.COMPLETED
            ? completedOrRolledBack.newVersion
            : completedOrRolledBack.oldVersion,
          status: completedOrRolledBack.status,
          inGrayscale: false
        };
      }

      return {
        appId,
        planId: null,
        version: 'unknown',
        status: null,
        inGrayscale: false
      };
    }

    const result = this.resolveVersion(activePlan.id, userId, userAttributes);
    return {
      appId,
      planId: result.planId,
      version: result.version,
      status: result.status,
      inGrayscale: result.inGrayscale
    };
  }
}

export const grayscaleService = new GrayscaleService();
