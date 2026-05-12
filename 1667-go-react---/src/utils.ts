import { ContractStatus } from './database';

export function extractVariables(templateContent: string): string[] {
  const regex = /\{\{(\w+)\}\}/g;
  const variables: string[] = [];
  let match;
  while ((match = regex.exec(templateContent)) !== null) {
    if (!variables.includes(match[1])) {
      variables.push(match[1]);
    }
  }
  return variables;
}

export function fillTemplateVariables(
  templateContent: string,
  variables: Record<string, string | number>,
): string {
  let content = templateContent;
  for (const [key, value] of Object.entries(variables)) {
    const regex = new RegExp(`\\{\\{${key}\\}\\}`, 'g');
    content = content.replace(regex, String(value));
  }
  return content;
}

export function validateVariables(
  templateContent: string,
  providedVariables: Record<string, string | number>,
): string[] {
  const requiredVars = extractVariables(templateContent);
  const missingVars: string[] = [];

  for (const required of requiredVars) {
    if (
      !(required in providedVariables) ||
      providedVariables[required] === undefined ||
      providedVariables[required] === null ||
      providedVariables[required] === ''
    ) {
      missingVars.push(required);
    }
  }

  return missingVars;
}

export const STATUS_FLOW: ContractStatus[] = [
  '草稿',
  '审批中',
  '已批准',
  '签署中',
  '已签署',
  '已归档',
];

export function isValidStatusTransition(
  currentStatus: ContractStatus,
  targetStatus: ContractStatus,
): boolean {
  if (currentStatus === '已作废') {
    return false;
  }
  if (targetStatus === '已作废') {
    return true;
  }

  const currentIndex = STATUS_FLOW.indexOf(currentStatus);
  const targetIndex = STATUS_FLOW.indexOf(targetStatus);

  if (currentIndex === -1 || targetIndex === -1) {
    return false;
  }

  return targetIndex === currentIndex + 1;
}

export function getNextStatus(currentStatus: ContractStatus): ContractStatus | null {
  if (currentStatus === '已作废') {
    return null;
  }

  const currentIndex = STATUS_FLOW.indexOf(currentStatus);
  if (currentIndex === -1 || currentIndex === STATUS_FLOW.length - 1) {
    return null;
  }

  return STATUS_FLOW[currentIndex + 1];
}

export function getApproverRole(amount: number): '部门经理' | '总监' | 'CEO' {
  if (amount < 50000) {
    return '部门经理';
  } else if (amount <= 200000) {
    return '总监';
  } else {
    return 'CEO';
  }
}

export function isValidTemplateType(type: string): type is '采购' | '销售' | '服务协议' | '保密协议' {
  return ['采购', '销售', '服务协议', '保密协议'].includes(type);
}
