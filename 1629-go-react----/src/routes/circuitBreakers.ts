import { Router, Request, Response } from 'express';
import { CircuitBreakerService, ValidationError, NotFoundError, InvalidStateTransitionError } from '../services/circuitBreakerService';
import { StatsService } from '../services/statsService';
import { CreateCircuitBreakerRequest, UpdateCircuitBreakerRequest } from '../types';

const router = Router();
const circuitBreakerService = new CircuitBreakerService();
const statsService = new StatsService();

function formatResponse(item: { config: any; circuitBreaker: any }) {
  return {
    id: item.circuitBreaker.id,
    config: {
      ...item.config,
    },
    state: {
      status: item.circuitBreaker.state,
      stateChangedAt: item.circuitBreaker.stateChangedAt,
      lastOpenReason: item.circuitBreaker.lastOpenReason,
    },
  };
}

router.get('/', (req: Request, res: Response) => {
  try {
    const circuitBreakers = circuitBreakerService.listCircuitBreakers();
    res.json(circuitBreakers.map(formatResponse));
  } catch (error) {
    handleError(error, res);
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const result = circuitBreakerService.getCircuitBreaker(req.params.id);
    if (!result) {
      res.status(404).json({ error: 'Circuit breaker not found' });
      return;
    }
    res.json(formatResponse(result));
  } catch (error) {
    handleError(error, res);
  }
});

router.post('/', (req: Request, res: Response) => {
  try {
    const body = req.body as CreateCircuitBreakerRequest;
    const result = circuitBreakerService.createCircuitBreaker(body);
    res.status(201).json(formatResponse(result));
  } catch (error) {
    handleError(error, res);
  }
});

router.put('/:id', (req: Request, res: Response) => {
  try {
    const body = req.body as UpdateCircuitBreakerRequest;
    const result = circuitBreakerService.updateCircuitBreaker(req.params.id, body);
    res.json(formatResponse(result));
  } catch (error) {
    handleError(error, res);
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  try {
    const deleted = circuitBreakerService.deleteCircuitBreaker(req.params.id);
    if (!deleted) {
      res.status(404).json({ error: 'Circuit breaker not found' });
      return;
    }
    res.status(204).send();
  } catch (error) {
    handleError(error, res);
  }
});

router.get('/:id/stats', (req: Request, res: Response) => {
  try {
    const stats = statsService.getStatsForCircuitBreaker(req.params.id);
    res.json(stats);
  } catch (error) {
    handleError(error, res);
  }
});

router.get('/:id/logs', (req: Request, res: Response) => {
  try {
    const exists = circuitBreakerService.getCircuitBreaker(req.params.id);
    if (!exists) {
      res.status(404).json({ error: 'Circuit breaker not found' });
      return;
    }
    const logs = circuitBreakerService.getStateTransitionLogs(req.params.id);
    res.json(logs);
  } catch (error) {
    handleError(error, res);
  }
});

function handleError(error: unknown, res: Response) {
  if (error instanceof ValidationError) {
    res.status(400).json({ error: error.message });
  } else if (error instanceof NotFoundError) {
    res.status(404).json({ error: error.message });
  } else if (error instanceof InvalidStateTransitionError) {
    res.status(500).json({ error: error.message });
  } else {
    console.error('Unexpected error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export default router;
