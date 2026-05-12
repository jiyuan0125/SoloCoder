package com.example.connectionpool.exception;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.util.Map;

@RestControllerAdvice
@Slf4j
public class GlobalExceptionHandler {

    @ExceptionHandler(PoolTimeoutException.class)
    public ResponseEntity<Map<String, Object>> handlePoolTimeout(PoolTimeoutException ex) {
        log.warn("Pool timeout: {}", ex.getMessage());
        return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body(Map.of(
                        "status", 503,
                        "error", "Service Unavailable",
                        "message", ex.getMessage(),
                        "poolName", ex.getPoolName()
                ));
    }

    @ExceptionHandler(PoolFullException.class)
    public ResponseEntity<Map<String, Object>> handlePoolFull(PoolFullException ex) {
        log.warn("Pool full: {}", ex.getMessage());
        return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body(Map.of(
                        "status", 503,
                        "error", "Service Unavailable",
                        "message", ex.getMessage(),
                        "poolName", ex.getPoolName(),
                        "queueSize", ex.getQueueSize()
                ));
    }

    @ExceptionHandler(InterruptedException.class)
    public ResponseEntity<Map<String, Object>> handleInterrupted(InterruptedException ex) {
        Thread.currentThread().interrupt();
        log.error("Request interrupted: {}", ex.getMessage());
        return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body(Map.of(
                        "status", 503,
                        "error", "Service Unavailable",
                        "message", "请求被中断"
                ));
    }
}
