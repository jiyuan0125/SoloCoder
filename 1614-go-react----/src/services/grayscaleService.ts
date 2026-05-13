import { shouldUseNewVersion } from '../utils/grayStrategy';
import { planService } from './planService';

const STATUS_PENDING = 'pending';
const STATUS_GRAYSCALE = 'grayscale';
const STATUS_FULL_RELEASE = 'full_release';
const STATUS_PAUSED = 'paused';
const STATUS_COMPLETED = 'completed';
const STATUS_ROLLED_BACK = 'rolled_back';

export class GrayscaleService {
  resolveVersion(
    planId: string,
    userId: string,
    userAttributes: Record<string, any> = {}
  ): {
    planId: string;
    version: string;
    status: string;
    inGrayscale: boolean;
  } {
    const plan = planService.getPlanById(planId);
    let version = plan.oldVersion;
    let inGrayscale = false;

    switch (plan.status) {
      case STATUS_PENDING:
        version = plan.oldVersion;
        break;
      case STATUS_GRAYSCALE:
        if (shouldUseNewVersion(planId, userId, plan.strategy, userAttributes)) {
          version = plan.newVersion;
          inGrayscale = true;
        } else {
          version = plan.oldVersion;
        }
        break;
      case STATUS_FULL_RELEASE:
      case STATUS_COMPLETED:
        version = plan.newVersion;
        break;
      case STATUS_PAUSED:
      case STATUS_ROLLED_BACK:
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
    status: string | null;
    inGrayscale: boolean;
  } {
    const allPlans = planService.getAllPlans();
    const appPlans = allPlans.filter(p => p.appId === appId);

    const activePlan = appPlans.find(
      p => p.status === STATUS_GRAYSCALE ||
           p.status === STATUS_FULL_RELEASE ||
           p.status === STATUS_PAUSED
    );

    if (!activePlan) {
      const completedOrRolledBack = appPlans.find(
        p => p.status === STATUS_COMPLETED || p.status === STATUS_ROLLED_BACK
      );
      if (completedOrRolledBack) {
        return {
          appId,
          planId: completedOrRolledBack.id,
          version: completedOrRolledBack.status === STATUS_COMPLETED
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
