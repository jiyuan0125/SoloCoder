export type Title = '助教' | '讲师' | '副教授' | '教授';
export type TrainingType = '教学能力' | '科研能力' | '管理能力' | '师德师风';
export type TrainingForm = '讲座' | '工作坊' | '在线课程' | '实践研修';
export type AttendanceStatus = '出勤' | '迟到' | '请假' | '缺席';
export type ReviewStage = '学院初审' | '校外专家盲审' | '校评审委员会终审';
export type ReviewStatus = '待评审' | '通过' | '不通过';

export interface Teacher {
  id: string;
  name: string;
  current_title: Title;
}

export interface Training {
  id: string;
  name: string;
  type: TrainingType;
  form: TrainingForm;
  date: string;
  hours: number;
  lecturer: string;
  capacity: number;
}

export interface AttendanceItem {
  status: AttendanceStatus;
}

export interface Registration {
  id: string;
  training_id: string;
  teacher_id: string;
  attendance?: AttendanceItem[];
  score?: number;
  study_report?: boolean;
}

export interface Score {
  student?: number;
  supervisor?: number;
}

export interface TeachingEvaluation {
  id: string;
  teacher_id: string;
  semester: string;
  attitude?: Score;
  content?: Score;
  method?: Score;
  effect?: Score;
  final_score: number;
  qualified: boolean;
}

export interface ReviewResult {
  status: ReviewStatus;
  comment?: string;
}

export interface TitleApplication {
  id: string;
  teacher_id: string;
  year: number;
  apply_title: Title;
  materials: string;
  achievement_summary: string;
  current_stage: ReviewStage;
  initial_review?: ReviewResult;
  external_review?: ReviewResult;
  final_review?: ReviewResult;
  final_status?: boolean;
}

export interface StatsResponse {
  training_hours: Record<Title, HoursStats>;
  evaluations: EvaluationStats;
  review_rate: ReviewStats;
}

export interface HoursStats {
  completed: number;
  required: number;
  count: number;
}

export interface EvaluationStats {
  total: number;
  qualified: number;
  avg_score: number;
}

export interface ReviewStats {
  total: number;
  passed: number;
  rate: number;
}
