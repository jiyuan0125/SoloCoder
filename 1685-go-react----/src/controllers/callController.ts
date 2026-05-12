import { Request, Response } from 'express';
import { callService, ElderNotFoundError } from '../services/callService';
import { CreateCallInput, RespondToCallInput } from '../types';

export const callController = {
  createCall: (req: Request, res: Response) => {
    try {
      const input: CreateCallInput = req.body;
      const call = callService.createCall(input);
      res.status(201).json(call);
    } catch (error: any) {
      if (error instanceof ElderNotFoundError) {
        res.status(404).json({ error: error.message });
      } else {
        res.status(400).json({ error: error.message });
      }
    }
  },

  getCall: (req: Request, res: Response) => {
    const call = callService.getCallById(req.params.id);
    if (!call) {
      res.status(404).json({ error: '呼叫记录不存在' });
      return;
    }
    res.json(call);
  },

  getCallsByElder: (req: Request, res: Response) => {
    const calls = callService.getCallsByElderId(req.params.elderId);
    res.json(calls);
  },

  getPendingCalls: (_req: Request, res: Response) => {
    const calls = callService.getPendingCalls();
    res.json(calls);
  },

  respondToCall: (req: Request, res: Response) => {
    const input: RespondToCallInput = { ...req.body, callId: req.params.id };
    const call = callService.respondToCall(input);
    if (!call) {
      res.status(404).json({ error: '呼叫记录不存在' });
      return;
    }
    res.json(call);
  },

  checkTimeout: (req: Request, res: Response) => {
    const call = callService.checkAndProcessTimeout(req.params.id);
    if (!call) {
      res.status(404).json({ error: '呼叫记录不存在' });
      return;
    }
    res.json(call);
  },

  getNotifications: (req: Request, res: Response) => {
    const notifications = callService.getNotificationsByCallId(req.params.id);
    res.json(notifications);
  },

  closeCall: (req: Request, res: Response) => {
    const call = callService.closeCall(req.params.id);
    if (!call) {
      res.status(404).json({ error: '呼叫记录不存在' });
      return;
    }
    res.json(call);
  },
};
