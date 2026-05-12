import { Request, Response } from 'express';
import { profilingService } from '../services/profilingService';
import { ProfilingType, StopRequest, CompareRequest, StartRequest } from '../types';

export const startProfiling = (req: Request, res: Response): void => {
  const { appId } = req.params;
  const { type } = req.body as StartRequest;

  if (!type || (type !== 'cpu' && type !== 'memory')) {
    res.status(400).json({ error: 'Invalid profiling type' });
    return;
  }

  try {
    const result = profilingService.startProfiling(appId, type as ProfilingType);
    res.status(201).json(result);
  } catch (err) {
    const error = err as Error;
    if (error.message === 'PROFILING_IN_PROGRESS') {
      res.status(409).json({ error: 'Profiling is already in progress' });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
};

export const stopProfiling = (req: Request, res: Response): void => {
  const { appId } = req.params;
  const body = req.body as StopRequest;

  if (!body.samples) {
    res.status(400).json({ error: 'Samples are required' });
    return;
  }

  try {
    const result = profilingService.stopProfiling(appId, body);
    res.status(200).json(result);
  } catch (err) {
    const error = err as Error;
    if (error.message === 'NO_RUNNING_PROFILING') {
      res.status(404).json({ error: 'No running profiling found' });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
};

export const getRecords = (req: Request, res: Response): void => {
  const { appId } = req.params;
  const records = profilingService.getRecords(appId);
  res.status(200).json(records);
};

export const getFlamegraph = (req: Request, res: Response): void => {
  const { appId, recordId } = req.params;

  try {
    const flamegraph = profilingService.getFlamegraph(appId, recordId);
    res.status(200).json(flamegraph);
  } catch (err) {
    const error = err as Error;
    if (error.message === 'RECORD_NOT_FOUND') {
      res.status(404).json({ error: 'Record not found' });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
};

export const compareRecords = (req: Request, res: Response): void => {
  const { recordId1, recordId2 } = req.body as CompareRequest;

  if (!recordId1 || !recordId2) {
    res.status(400).json({ error: 'Both recordId1 and recordId2 are required' });
    return;
  }

  try {
    const result = profilingService.compare(recordId1, recordId2);
    res.status(200).json(result);
  } catch (err) {
    const error = err as Error;
    if (error.message === 'RECORD_NOT_FOUND') {
      res.status(404).json({ error: 'One or both records not found' });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
};
