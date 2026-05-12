export enum UserRole {
  ADMIN = 'admin',
  INSTRUCTOR = 'instructor',
  STUDENT = 'student'
}

export enum InstructorStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  SUSPENDED = 'suspended'
}

export enum CourseStatus {
  DRAFT = 'draft',
  ENROLLING = 'enrolling',
  IN_PROGRESS = 'in_progress',
  COMPLETED = 'completed',
  CANCELLED = 'cancelled'
}

export enum EnrollmentStatus {
  PENDING = 'pending',
  ENROLLED = 'enrolled',
  WAITING_LIST = 'waiting_list',
  DROPPED = 'dropped',
  COMPLETED = 'completed'
}

export interface CreateCourseRequest {
  name: string;
  category: string;
  description: string;
  instructorId: string;
  totalHours: number;
  hoursPerSession: number;
  maxEnrollment: number;
  startDate: string;
  fee?: number;
  prerequisites?: string[];
}

export interface InstructorRegistrationRequest {
  name: string;
  qualifications: string;
  expertise: string;
  teachingExperience: string;
}

export interface EnrollmentRequest {
  courseId: string;
  studentId: string;
}

export interface EvaluationRequest {
  courseId: string;
  studentId: string;
  rating: number;
  comments?: string;
}
