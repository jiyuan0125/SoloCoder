import { Request, Response, NextFunction } from 'express';
import { planService } from '../services/planService';
import { grayscaleService } from '../services/grayscaleService';

function getParamId(req: Request): string {
  return Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
}

export const planController = {
  createPlan: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.createPlan(req.body);
      res.status(201).json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  getPlan: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.getPlanById(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  getAllPlans: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plans = planService.getAllPlans();
      res.json(plans);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  updatePlan: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.updatePlan(getParamId(req), req.body);
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  deletePlan: async (req: Request, res: Response, next: NextFunction) => {
    try {
      planService.deletePlan(getParamId(req));
      res.status(204).send();
    } catch (error: any) {
      handleError(error, res);
    }
  },

  startGrayscale: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.startGrayscale(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  fullRelease: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.fullRelease(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  complete: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.complete(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  pause: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.pause(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  resume: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.resume(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  rollback: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const plan = planService.rollback(getParamId(req));
      res.json(plan);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  recordRequest: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { isError } = req.body;
      const result = planService.recordErrorAndCheckThreshold(getParamId(req), !!isError);
      res.json(result);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  getErrorRate: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const stats = planService.getErrorRate(getParamId(req));
      res.json(stats);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  resolveVersion: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { userId, userAttributes } = req.body;
      if (!userId) {
        return res.status(400).json({ message: 'userId 是必填字段' });
      }
      const result = grayscaleService.resolveVersion(
        getParamId(req),
        userId,
        userAttributes || {}
      );
      res.json(result);
    } catch (error: any) {
      handleError(error, res);
    }
  },

  resolveVersionByAppId: async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { appId, userId, userAttributes } = req.body;
      if (!appId || !userId) {
        return res.status(400).json({ message: 'appId 和 userId 是必填字段' });
      }
      const result = grayscaleService.resolveVersionByAppId(
        appId,
        userId,
        userAttributes || {}
      );
      res.json(result);
    } catch (error: any) {
      handleError(error, res);
    }
  }
};

function handleError(error: any, res: Response) {
  if (error.status) {
    const response: any = { message: error.message };
    if (error.allowedActions) {
      response.allowedActions = error.allowedActions;
    }
    res.status(error.status).json(response);
  } else {
    res.status(500).json({ message: error.message || '内部服务器错误' });
  }
}
