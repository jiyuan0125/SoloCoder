export type Role = 'admin' | 'verifier' | 'viewer';

export type EducationLevel = 'college' | 'bachelor' | 'master' | 'doctorate';

export type EducationStatus = 'valid' | 'invalid';

export type OperationType = 
  | 'create_diploma' 
  | 'update_diploma' 
  | 'delete_diploma' 
  | 'verify_diploma' 
  | 'create_user' 
  | 'update_user' 
  | 'delete_user';

export type VerificationResult = 'matched' | 'unmatched';

export interface User {
  id: string;
  username: string;
  role: Role;
}

export interface Diploma {
  id: string;
  name: string;
  id_card: string;
  school: string;
  level: EducationLevel;
  major: string;
  study_years: number;
  enrollment_date: string;
  graduation_date: string;
  status: EducationStatus;
  created_at: string;
  updated_at: string;
}

export interface OperationLog {
  id: string;
  operator_id: string;
  operator_name: string;
  operator_role: Role;
  operation_type: OperationType;
  content: string;
  verification_result?: VerificationResult;
  created_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface CreateDiplomaRequest {
  name: string;
  id_card: string;
  school: string;
  level: EducationLevel;
  major: string;
  study_years: number;
  enrollment_date: string;
  graduation_date: string;
}

export interface UpdateDiplomaRequest {
  name?: string;
  id_card?: string;
  school?: string;
  level?: EducationLevel;
  major?: string;
  study_years?: number;
  enrollment_date?: string;
  graduation_date?: string;
  status?: EducationStatus;
}

export interface VerifyRequest {
  name: string;
  id_card: string;
}

export interface VerifyResponse {
  diplomas: Diploma[];
  result: string;
}

export interface CreateUserRequest {
  username: string;
  password: string;
  role: Role;
}

export interface UpdateUserRequest {
  password?: string;
  role?: Role;
}

export interface PagedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
}
