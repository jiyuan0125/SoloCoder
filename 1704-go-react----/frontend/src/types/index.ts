export interface Patient {
  id: number;
  name: string;
  birthDate: string;
  gender: string;
  phone: string;
  status: string;
  plans?: Plan[];
}

export interface Plan {
  id: number;
  patientId: number;
  name: string;
  assessmentType: string;
  startDate: string;
  durationWeeks: number;
  status: string;
  effectiveness: string;
  exercises?: Exercise[];
  tasks?: Task[];
}

export interface Exercise {
  id: number;
  planId: number;
  name: string;
  goalDescription: string;
  frequencyType: 'daily' | 'weekly';
  frequencyCount: number;
  durationMinutes: number;
  difficultyLevel: number;
  notes?: string;
}

export interface Task {
  id: number;
  planId: number;
  exerciseId: number;
  taskDate: string;
  taskTime: string;
  status: 'pending' | 'completed' | 'overdue' | 'abandoned';
  exercise?: Exercise;
  plan?: Plan;
}

export interface TrainingRecord {
  id: number;
  taskId: number;
  exerciseId: number;
  planId: number;
  patientId: number;
  completedAt: string;
  actualDuration: number;
  qualityScore: number;
  subjectiveFeeling: 'easy' | 'normal' | 'hard' | 'very_hard';
  notes?: string;
}

export interface DifficultyLog {
  id: number;
  exerciseId: number;
  oldDifficulty: number;
  newDifficulty: number;
  adjustReason: string;
  adjustedAt: string;
}

export interface Assessment {
  id: number;
  planId: number;
  patientId: number;
  assessmentType: string;
  scheduledDate: string;
  assessmentDate?: string;
  initialScore: number;
  currentScore: number;
  improvementPercent: number;
  notes?: string;
  indicators?: AssessmentIndicator[];
}

export interface AssessmentIndicator {
  id: number;
  assessmentId: number;
  name: string;
  value: number;
  previousValue: number;
  change: number;
  unit: string;
}

export interface ExerciseSummary {
  exerciseId: number;
  exerciseName: string;
  avgQualityScore: number;
  plannedDuration: number;
  avgActualDuration: number;
  durationDeviation: number;
  totalCompleted: number;
  difficultyChanges: DifficultyLog[];
}

export interface AssessmentTrend {
  assessmentDate: string;
  score: number;
  improvementPercent: number;
  indicators: Record<string, number>;
}

export interface PatientSummary {
  patientId: number;
  patientName: string;
  planName: string;
  planStartDate: string;
  planEndDate: string;
  planProgress: number;
  effectiveness: string;
  exerciseSummaries: ExerciseSummary[];
  assessmentTrends: AssessmentTrend[];
}
