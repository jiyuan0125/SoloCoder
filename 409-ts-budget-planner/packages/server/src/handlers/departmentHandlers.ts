import {
  Department, CreateDepartmentRequest,
  createSuccessResponse, createErrorResponse, ErrorCode
} from '@budget-planner/shared';
import { ParsedRequest, RouteResult } from '../router';
import { addDepartment, getDepartment, getAllDepartments } from '../store/dataStore';
import { isValidString } from '../utils/validation';

function validateCreateDepartmentRequest(body: unknown): body is CreateDepartmentRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  return (
    isValidString(obj.id) &&
    isValidString(obj.name) &&
    isValidString(obj.managerId)
  );
}

export function handleGetDepartments(_req: ParsedRequest): RouteResult {
  const departments = getAllDepartments();
  return {
    statusCode: 200,
    body: createSuccessResponse(departments),
  };
}

export function handlePostDepartments(req: ParsedRequest): RouteResult {
  if (!validateCreateDepartmentRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid department request: required fields: id, name, managerId'
      ),
    };
  }

  const existingDepartment = getDepartment(req.body.id);
  if (existingDepartment) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.DEPARTMENT_NOT_FOUND,
        'Department with this ID already exists'
      ),
    };
  }

  const department: Department = {
    id: req.body.id,
    name: req.body.name,
    managerId: req.body.managerId,
  };

  addDepartment(department);

  return {
    statusCode: 201,
    body: createSuccessResponse(department),
  };
}
