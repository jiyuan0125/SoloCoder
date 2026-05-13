import { Request, Response, NextFunction } from 'express';
import { environmentService } from '../services/environment.service';
import {
  CreateEnvironmentRequest,
  UpdateEnvironmentRequest,
  UpdateStatusRequest,
  CloneEnvironmentRequest,
  RollbackRequest,
  ApiError,
} from '../types';

export const environmentController = {
  async create(req: Request, res: Response, next: NextFunction) {
    try {
      const request = req.body as CreateEnvironmentRequest;
      const env = environmentService.create(request);
      res.status(201).json(env);
    } catch (err) {
      next(err);
    }
  },

  async list(req: Request, res: Response, next: NextFunction) {
    try {
      const { projectId, type } = req.query;
      const filters: { projectId?: string; type?: string } = {};
      if (typeof projectId === 'string') {
        filters.projectId = projectId;
      }
      if (typeof type === 'string') {
        filters.type = type;
      }
      const environments = environmentService.list(filters);
      res.status(200).json(environments);
    } catch (err) {
      next(err);
    }
  },

  async getById(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const env = environmentService.getById(id);
      res.status(200).json(env);
    } catch (err) {
      next(err);
    }
  },

  async update(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const request = req.body as UpdateEnvironmentRequest;
      const env = await environmentService.update(id, request);
      res.status(200).json(env);
    } catch (err) {
      next(err);
    }
  },

  async delete(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      await environmentService.delete(id);
      res.status(204).send();
    } catch (err) {
      next(err);
    }
  },

  async updateStatus(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const { status } = req.body as UpdateStatusRequest;
      const env = await environmentService.updateStatus(id, status);
      res.status(200).json(env);
    } catch (err) {
      next(err);
    }
  },

  async clone(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const { name } = req.body as CloneEnvironmentRequest;
      const env = await environmentService.clone(id, name);
      res.status(201).json(env);
    } catch (err) {
      next(err);
    }
  },

  async getConfigHistory(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const history = environmentService.getConfigHistory(id);
      res.status(200).json(history);
    } catch (err) {
      next(err);
    }
  },

  async rollback(req: Request, res: Response, next: NextFunction) {
    try {
      const { id } = req.params;
      const { version } = req.body as RollbackRequest;
      const result = await environmentService.rollback(id, version);
      const response: Record<string, unknown> = {
        environment: result.environment,
      };
      if (result.warning) {
        response.warning = result.warning;
      }
      res.status(200).json(response);
    } catch (err) {
      next(err);
    }
  },
};

export function errorHandler(
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction,
) {
  if (err instanceof ApiError) {
    res.status(err.statusCode).json({ error: err.message });
    return;
  }
  console.error(err);
  res.status(500).json({ error: 'internal server error' });
}
