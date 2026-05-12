export interface ScaleOption {
  id: string;
  text: string;
  score: number;
}

export interface ScaleQuestion {
  id: string;
  text: string;
  options: ScaleOption[];
  isReverseScored: boolean;
  dimension?: string;
}

export interface ScaleDimension {
  name: string;
  weight: number;
}

export interface Scale {
  id: string;
  name: string;
  targetGroup: string;
  questions: ScaleQuestion[];
  dimensions?: ScaleDimension[];
  scoringType: 'sum' | 'weighted';
  normTable: NormEntry[];
}

export interface NormEntry {
  rawScore: number;
  standardScore: number;
  level: 'normal' | 'mild' | 'moderate' | 'severe';
}

export interface Answer {
  questionId: string;
  optionId: string;
  score: number;
  answerTime: number;
}

export interface Assessment {
  id: string;
  userId: string;
  scaleId: string;
  startTime: number;
  status: 'in_progress' | 'paused' | 'completed' | 'expired';
  answers: Answer[];
  lastActiveTime: number;
}

export interface AssessmentResult {
  id: string;
  assessmentId: string;
  userId: string;
  scaleId: string;
  rawScore: number;
  standardScore: number;
  level: 'normal' | 'mild' | 'moderate' | 'severe';
  dimensionScores?: { [key: string]: number };
  completedAt: number;
  report: string;
}

export interface Alert {
  id: string;
  assessmentResultId: string;
  userId: string;
  scaleId: string;
  level: 'moderate' | 'severe';
  priority: 'normal' | 'urgent';
  status: 'unprocessed' | 'processing' | 'resolved';
  notifiedUsers: string[];
  createdAt: number;
  processingRecord?: string;
  resolvedAt?: number;
  processedBy?: string;
}

export interface Notification {
  id: string;
  alertId: string;
  userId: string;
  recipientId: string;
  content: string;
  sentAt: number;
  readAt?: number;
}
