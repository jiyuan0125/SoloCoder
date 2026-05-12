export type IndustryCategory = '制造' | '服务' | 'IT' | '建筑' | '医疗' | '其他';
export type SkillLevel = '初级' | '中级' | '高级' | '技师' | '高级技师';
export type ExamSubject = '理论知识' | '实操技能' | '综合评审';
export type BatchStatus = '未开始' | '进行中' | '已结束';
export type CertificateStatus = '有效' | '注销' | '过期';

export interface LevelConfig {
  level: SkillLevel;
  subjects: ExamSubject[];
}

export interface Occupation {
  id: string;
  name: string;
  code: string;
  industry: IndustryCategory;
  levels: LevelConfig[];
}

export interface ExamCandidate {
  id: string;
  batchId: string;
  name: string;
  idCard: string;
  phone: string;
  appliedLevel: SkillLevel;
  scores: Record<ExamSubject, number>;
  hasTakenExam: boolean;
  isPassed: boolean;
}

export interface ExamBatch {
  id: string;
  name: string;
  occupationId: string;
  occupationName: string;
  level: SkillLevel;
  examDate: string;
  examRoom: string;
  status: BatchStatus;
  candidates: ExamCandidate[];
  createdAt: string;
}

export interface Certificate {
  id: string;
  certificateNo: string;
  candidateId: string;
  name: string;
  idCard: string;
  occupationId: string;
  occupationName: string;
  level: SkillLevel;
  status: CertificateStatus;
  issueDate: string;
  expiryDate: string;
}

export interface PassRateStats {
  occupationId: string;
  occupationName: string;
  level: SkillLevel;
  totalTaken: number;
  totalPassed: number;
  passRate: number;
}
