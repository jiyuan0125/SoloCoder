package com.employee.server.exception;

import com.employee.common.constant.ErrorCode;
import com.employee.common.response.ApiResponse;
import com.employee.server.service.EmployeeService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

@ControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(EmployeeService.BusinessException.class)
    public ResponseEntity<ApiResponse<?>> handleBusinessException(EmployeeService.BusinessException e) {
        return ResponseEntity.status(HttpStatus.OK)
                .body(ApiResponse.error(e.getErrorCode(), e.getMessage()));
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ApiResponse<?>> handleException(Exception e) {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(ApiResponse.error(ErrorCode.SYSTEM_ERROR, e.getMessage()));
    }
}
