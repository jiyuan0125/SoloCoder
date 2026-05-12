import { AuditStatus } from '../types';
import { createAuditRecord, createContent, getAuditRecordById, getContentById, updateAuditRecordDecision, updateContentStatus } from '../db';

export interface ImageAuditResult {
  auditId: string;
  status: AuditStatus;
}

export interface BatchImageAuditItem {
  contentId: string;
  imageUrl: string;
}

export interface BatchImageAuditResultItem {
  contentId: string;
  auditId: string;
  status: AuditStatus;
  error?: string;
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substring(2, 10);
}

function simpleHash(str: string): number {
  let hash = 5381;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash) + str.charCodeAt(i);
    hash = hash & hash;
  }
  return Math.abs(hash);
}

function decideImageStatus(imageUrl: string): AuditStatus {
  const hash = simpleHash(imageUrl);
  const mod = hash % 100;

  if (mod < 5) {
    return 'queue';
  } else if (mod < 35) {
    return 'pending';
  } else {
    return 'approved';
  }
}

export function auditImage(contentId: string, imageUrl: string): ImageAuditResult {
  const status = decideImageStatus(imageUrl);
  const auditId = generateId();

  const existingContent = getContentById(contentId);
  if (!existingContent) {
    createContent({
      id: contentId,
      type: 'image',
      content: imageUrl,
      status
    });
  } else {
    updateContentStatus(contentId, status);
  }

  createAuditRecord({
    id: auditId,
    contentId,
    status
  });

  return {
    auditId,
    status
  };
}

export function batchAuditImages(items: BatchImageAuditItem[]): BatchImageAuditResultItem[] {
  return items.map(item => {
    try {
      const result = auditImage(item.contentId, item.imageUrl);
      return {
        contentId: item.contentId,
        auditId: result.auditId,
        status: result.status
      };
    } catch (err: unknown) {
      return {
        contentId: item.contentId,
        auditId: '',
        status: 'pending' as AuditStatus,
        error: err instanceof Error ? err.message : '未知错误'
      };
    }
  });
}

export function decideImageAudit(
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
