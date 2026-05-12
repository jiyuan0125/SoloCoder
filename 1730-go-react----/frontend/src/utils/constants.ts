import { EducationLevel, EducationStatus, Role, OperationType } from '../types';

export const EDUCATION_LEVEL_MAP: Record<EducationLevel, string> = {
  college: '大专',
  bachelor: '本科',
  master: '硕士',
  doctorate: '博士',
};

export const EDUCATION_STATUS_MAP: Record<EducationStatus, string> = {
  valid: '有效',
  invalid: '无效',
};

export const ROLE_MAP: Record<Role, string> = {
  admin: '管理员',
  verifier: '认证员',
  viewer: '查看员',
};

export const OPERATION_TYPE_MAP: Record<OperationType, string> = {
  create_diploma: '创建学历',
  update_diploma: '更新学历',
  delete_diploma: '删除学历',
  verify_diploma: '学历认证',
  create_user: '创建用户',
  update_user: '更新用户',
  delete_user: '删除用户',
};

export const VERIFICATION_RESULT_MAP = {
  matched: '一致',
  unmatched: '不一致',
};
