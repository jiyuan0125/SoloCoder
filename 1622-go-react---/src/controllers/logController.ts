import { Request, Response } from 'express';
import { insertLog, queryLogs, queryTraceLogs, getAggregationStats, isValidLogLevel } from '../services/logService';
import { LogEntry, VALID_LOG_LEVELS, LogLevel } from '../types';

export async function createLog(req: Request, res: Response) {
  try {
    const log: LogEntry = req.body;

    if (!log.level || !isValidLogLevel(log.level)) {
      return res.status(400).json({
        error: 'Invalid log level',
        validLevels: VALID_LOG_LEVELS,
      });
    }

    if (!log.timestamp) {
      log.timestamp = Date.now();
    }

    if (!log.tags || !Array.isArray(log.tags)) {
      log.tags = [];
    }

    await insertLog(log);
    return res.status(201).json({ success: true });
  } catch (error) {
    console.error('Error creating log:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}

export async function listLogs(req: Request, res: Response) {
  try {
    const startTime = req.query.startTime ? parseInt(req.query.startTime as string, 10) : undefined;
    const endTime = req.query.endTime ? parseInt(req.query.endTime as string, 10) : undefined;
    const levelRaw = req.query.level as string | undefined;
    const keyword = req.query.keyword as string | undefined;
    const page = req.query.page ? parseInt(req.query.page as string, 10) : undefined;
    const pageSize = req.query.pageSize ? parseInt(req.query.pageSize as string, 10) : undefined;

    if (startTime === undefined || endTime === undefined) {
      return res.status(400).json({ error: 'Time range is required (startTime and endTime)' });
    }

    let level: LogLevel | undefined;
    if (levelRaw) {
      if (!isValidLogLevel(levelRaw)) {
        return res.status(400).json({
          error: 'Invalid log level',
          validLevels: VALID_LOG_LEVELS,
        });
      }
      level = levelRaw;
    }

    const result = await queryLogs({
      startTime,
      endTime,
      level,
      keyword,
      page,
      pageSize,
    });

    return res.json(result);
  } catch (error) {
    console.error('Error listing logs:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}

export async function listLogsByService(req: Request, res: Response) {
  try {
    const serviceName = req.params.serviceName;
    const startTime = req.query.startTime ? parseInt(req.query.startTime as string, 10) : undefined;
    const endTime = req.query.endTime ? parseInt(req.query.endTime as string, 10) : undefined;
    const levelRaw = req.query.level as string | undefined;
    const keyword = req.query.keyword as string | undefined;
    const page = req.query.page ? parseInt(req.query.page as string, 10) : undefined;
    const pageSize = req.query.pageSize ? parseInt(req.query.pageSize as string, 10) : undefined;

    if (startTime === undefined || endTime === undefined) {
      return res.status(400).json({ error: 'Time range is required (startTime and endTime)' });
    }

    let level: LogLevel | undefined;
    if (levelRaw) {
      if (!isValidLogLevel(levelRaw)) {
        return res.status(400).json({
          error: 'Invalid log level',
          validLevels: VALID_LOG_LEVELS,
        });
      }
      level = levelRaw;
    }

    const result = await queryLogs({
      startTime,
      endTime,
      serviceName,
      level,
      keyword,
      page,
      pageSize,
    });

    return res.json(result);
  } catch (error) {
    console.error('Error listing logs by service:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getTraceLogs(req: Request, res: Response) {
  try {
    const traceId = req.params.traceId;
    const logs = await queryTraceLogs(traceId);

    if (logs.length === 0) {
      return res.status(404).json({ error: 'Trace not found' });
    }

    return res.json({ logs });
  } catch (error) {
    console.error('Error getting trace logs:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getStats(req: Request, res: Response) {
  try {
    const startTime = req.query.startTime ? parseInt(req.query.startTime as string, 10) : undefined;
    const endTime = req.query.endTime ? parseInt(req.query.endTime as string, 10) : undefined;

    if (startTime === undefined || endTime === undefined) {
      return res.status(400).json({ error: 'Time range is required (startTime and endTime)' });
    }

    const stats = await getAggregationStats(startTime, endTime);
    return res.json(stats);
  } catch (error) {
    console.error('Error getting stats:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}
