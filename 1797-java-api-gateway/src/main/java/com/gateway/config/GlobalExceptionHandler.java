package com.gateway.config;

import com.gateway.core.GatewayProcessor;
import com.gateway.forward.ForwardException;
import com.gateway.model.ErrorResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@Slf4j
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(ForwardException.class)
    public ResponseEntity<ErrorResponse> handleForwardException(ForwardException e) {
        log.error("转发异常: {} - {} (后端: {})", e.getErrorCode(), e.getMessage(), e.getBackendUrl());
        
        String message = e.getMessage();
        if (e.getBackendUrl() != null) {
            message = message + " [后端: " + e.getBackendUrl() + "]";
        }
        
        ErrorResponse response = new ErrorResponse(e.getErrorCode(), message);
        return ResponseEntity.status(e.getHttpStatus()).body(response);
    }

    @ExceptionHandler(GatewayProcessor.FilterRejectException.class)
    public ResponseEntity<ErrorResponse> handleFilterRejectException(GatewayProcessor.FilterRejectException e) {
        log.warn("过滤器拒绝请求: {} - {}", e.getErrorCode(), e.getMessage());
        
        ErrorResponse response = new ErrorResponse(e.getErrorCode(), e.getMessage());
        return ResponseEntity.status(e.getHttpStatus()).body(response);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleException(Exception e) {
        log.error("服务器内部错误", e);
        
        ErrorResponse response = new ErrorResponse("INTERNAL_ERROR", "服务器内部错误");
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
    }
}
