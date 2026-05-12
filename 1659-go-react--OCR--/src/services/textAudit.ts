import { sensitiveWordManager } from '../utils/sensitiveWords';
import { mapBackToOriginal, normalizeText } from '../utils/normalize';
import { AuditStatus, SensitiveWordMatch } from '../types';
import { createAuditRecord, createContent, getAuditRecordById, getContentById, updateAuditRecordDecision, updateContentStatus } from '../db';

export interface TextAuditResult {
  auditId: string;
  status: AuditStatus;
  matches: SensitiveWordMatch[];
}

export interface BatchTextAuditItem {
  contentId: string;
  text: string;
}

export interface BatchTextAuditResultItem {
  contentId: string;
  auditId: string;
  status: AuditStatus;
  matches: SensitiveWordMatch[];
  error?: string;
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substring(2, 10);
}

export function auditText(contentId: string, text: string): TextAuditResult | { error: string; statusCode: number } {
  if (!text || text.trim().length === 0) {
    return { error: '文本不能为空', statusCode: 400 };
  }

  const { normalized, mappings } = normalizeText(text);
  const normalizedMatches = sensitiveWordManager.scan(normalized);
  const matches = mapBackToOriginal(normalizedMatches, mappings);

  const status: AuditStatus = matches.length > 0 ? 'queue' : 'approved';
  const auditId = generateId();

  const existingContent = getContentById(contentId);
  if (!existingContent) {
    createContent({
      id: contentId,
      type: 'text',
      content: text,
      status
    });
  } else {
    updateContentStatus(contentId, status);
  }

  createAuditRecord({
    id: auditId,
    contentId,
    status,
    matches
  });

  return {
    auditId,
    status,
    matches
  };
}

export function batchAuditText(items: BatchTextAuditItem[]): BatchTextAuditResultItem[] {
  return items.map(item => {
    try {
      const result = auditText(item.contentId, item.text);
      if ('error' in result) {
        return {
          contentId: item.contentId,
          auditId: '',
          status: 'pending' as AuditStatus,
          matches: [],
          error: result.error
        };
      }
      return {
        contentId: item.contentId,
        auditId: result.auditId,
        status: result.status,
        matches: result.matches
      };
    } catch (err: unknown) {
      return {
        contentId: item.contentId,
        auditId: '',
        status: 'pending' as AuditStatus,
        matches: [],
        error: err instanceof Error ? err.message : '未知错误'
      };
    }
  });
}

export function decideTextAudit(
  contentId: string,
  auditId: string,
  decision: 'approved' | 'rejected' | 'warned'
): { success: boolean; error?: string; statusCode?: number } {
  const content = getContentById(contentId);
  if (!content) {
    return { success: false, error: '内容不存在', statusCode: 404 };
  }

  const auditRecord = getAuditRecordById(auditId);
  if (!auditRecord) {
    return { success: false, error: '审核记录不存在', statusCode: 404 };
  }

  if (auditRecord.decision) {
    return { success: false, error: '已审核', statusCode: 409 };
  }

  const statusMap: { [key: string]: AuditStatus } = {
    approved: 'approved',
    rejected: 'rejected',
    warned: 'warned'
  };

  const newStatus = statusMap[decision];
  updateAuditRecordDecision(auditId, newStatus, decision);
  updateContentStatus(contentId, newStatus);

  return { success: true };
}
