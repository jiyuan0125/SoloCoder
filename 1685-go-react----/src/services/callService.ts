import { db } from '../database';
import {
  CallRecord,
  CreateCallInput,
  RespondToCallInput,
  CallType,
  CallStatus,
  Notification,
} from '../types';
import { generateId, now } from '../utils';
import { elderService } from './elderService';

const EMERGENCY_RESPONSE_LIMIT = 5;
const DAILY_RESPONSE_LIMIT = 30;
const COMMUNITY_NOTIFY_DELAY = 5;
const STREET_NOTIFY_DELAY = 15;

function mapCallRecord(row: any): CallRecord {
  return {
    id: row.id,
    elderId: row.elder_id,
    type: row.type as CallType,
    status: row.status as CallStatus,
    calledAt: row.called_at,
    respondedAt: row.responded_at,
    responseTime: row.response_time,
    staffId: row.staff_id,
    createdAt: row.created_at,
  };
}

function mapNotification(row: any): Notification {
  return {
    id: row.id,
    callRecordId: row.call_record_id,
    recipientType: row.recipient_type as 'COMMUNITY' | 'STREET',
    notifiedAt: row.notified_at,
    message: row.message,
  };
}

export class ElderNotFoundError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ElderNotFoundError';
  }
}

export const callService = {
  getResponseLimitMinutes(type: CallType): number {
    return type === CallType.EMERGENCY ? EMERGENCY_RESPONSE_LIMIT : DAILY_RESPONSE_LIMIT;
  },

  createCall(input: CreateCallInput): CallRecord | null {
    const elder = elderService.getElderById(input.elderId);
    if (!elder) {
      throw new ElderNotFoundError('老人ID不存在');
    }

    const id = generateId();
    const timestamp = now();

    const stmt = db.prepare(`
      INSERT INTO call_records (
        id, elder_id, type, status, called_at, responded_at, response_time, staff_id, created_at
      ) VALUES (?, ?, ?, ?, ?, NULL, NULL, NULL, ?)
    `);
    stmt.run(id, input.elderId, input.type, CallStatus.PENDING, timestamp, timestamp);

    return callService.getCallById(id);
  },

  getCallById(id: string): CallRecord | null {
    const row = db.prepare('SELECT * FROM call_records WHERE id = ?').get(id);
    return row ? mapCallRecord(row) : null;
  },

  getCallsByElderId(elderId: string): CallRecord[] {
    const rows = db.prepare(
      'SELECT * FROM call_records WHERE elder_id = ? ORDER BY called_at DESC'
    ).all(elderId);
    return rows.map(mapCallRecord);
  },

  getPendingCalls(): CallRecord[] {
    const rows = db.prepare(
      'SELECT * FROM call_records WHERE status IN (?, ?, ?) ORDER BY called_at ASC'
    ).all(CallStatus.PENDING, CallStatus.NOTIFY_COMMUNITY, CallStatus.NOTIFY_STREET);
    return rows.map(mapCallRecord);
  },

  respondToCall(input: RespondToCallInput): CallRecord | null {
    const call = callService.getCallById(input.callId);
    if (!call) return null;
    if (call.status === CallStatus.CLOSED) return call;

    const respondedAt = now();
    const calledDate = new Date(call.calledAt);
    const respondedDate = new Date(respondedAt);
    const responseTime = Math.round((respondedDate.getTime() - calledDate.getTime()) / 60000);

    db.prepare(`
      UPDATE call_records 
      SET status = ?, responded_at = ?, response_time = ?, staff_id = ?
      WHERE id = ?
    `).run(CallStatus.RESPONDED, respondedAt, responseTime, input.staffId, input.callId);

    return callService.getCallById(input.callId);
  },

  checkAndProcessTimeout(callId: string): CallRecord | null {
    const call = callService.getCallById(callId);
    if (!call) return null;
    if (call.status === CallStatus.CLOSED || call.status === CallStatus.RESPONDED) {
      return call;
    }

    const nowDate = new Date();
    const calledDate = new Date(call.calledAt);
    const minutesElapsed = (nowDate.getTime() - calledDate.getTime()) / 60000;

    let newStatus = call.status;
    let shouldNotifyCommunity = false;
    let shouldNotifyStreet = false;

    if (minutesElapsed >= COMMUNITY_NOTIFY_DELAY && call.status === CallStatus.PENDING) {
      newStatus = CallStatus.NOTIFY_COMMUNITY;
      shouldNotifyCommunity = true;
    }

    if (minutesElapsed >= STREET_NOTIFY_DELAY && call.status !== CallStatus.NOTIFY_STREET) {
      newStatus = CallStatus.NOTIFY_STREET;
      shouldNotifyStreet = true;
    }

    if (shouldNotifyCommunity) {
      callService.createNotification(
        callId,
        'COMMUNITY',
        `紧急呼叫超时${COMMUNITY_NOTIFY_DELAY}分钟，通知社区负责人`
      );
    }

    if (shouldNotifyStreet) {
      callService.createNotification(
        callId,
        'STREET',
        `紧急呼叫超时${STREET_NOTIFY_DELAY}分钟，通知街道负责人`
      );
    }

    if (newStatus !== call.status) {
      db.prepare('UPDATE call_records SET status = ? WHERE id = ?').run(newStatus, callId);
      return callService.getCallById(callId);
    }

    return call;
  },

  createNotification(
    callRecordId: string,
    recipientType: 'COMMUNITY' | 'STREET',
    message: string
  ): Notification {
    const id = generateId();
    const timestamp = now();
    db.prepare(`
      INSERT INTO notifications (id, call_record_id, recipient_type, notified_at, message)
      VALUES (?, ?, ?, ?, ?)
    `).run(id, callRecordId, recipientType, timestamp, message);

    const row = db.prepare('SELECT * FROM notifications WHERE id = ?').get(id);
    return mapNotification(row);
  },

  getNotificationsByCallId(callId: string): Notification[] {
    const rows = db.prepare(
      'SELECT * FROM notifications WHERE call_record_id = ? ORDER BY notified_at ASC'
    ).all(callId);
    return rows.map(mapNotification);
  },

  closeCall(callId: string): CallRecord | null {
    const call = callService.getCallById(callId);
    if (!call) return null;

    db.prepare('UPDATE call_records SET status = ? WHERE id = ?').run(
      CallStatus.CLOSED,
      callId
    );
    return callService.getCallById(callId);
  },
};
