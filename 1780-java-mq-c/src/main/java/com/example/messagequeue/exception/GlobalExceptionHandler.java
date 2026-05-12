package com.example.messagequeue.exception;

import com.example.messagequeue.model.ErrorResponse;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.util.HashMap;
import java.util.Map;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(MessageTooLargeException.class)
    public ResponseEntity<Map<String, Object>> handleMessageTooLarge(
            MessageTooLargeException ex, HttpServletRequest request) {
        Map<String, Object> body = new HashMap<>();
        body.put("status", HttpStatus.PAYLOAD_TOO_LARGE.value());
        body.put("error", "Payload Too Large");
        body.put("message", String.format(
                "Message size exceeds maximum allowed. Actual: %d bytes, Max: %d bytes. " +
                "Please split the message into smaller parts or compress the content.",
                ex.getActualSize(), ex.getMaxSize()));
        body.put("actualSize", ex.getActualSize());
        body.put("maxSize", ex.getMaxSize());
        body.put("path", request.getRequestURI());
        body.put("timestamp", java.time.Instant.now().toString());

        return ResponseEntity.status(HttpStatus.PAYLOAD_TOO_LARGE).body(body);
    }

    @ExceptionHandler(InvalidJsonException.class)
    public ResponseEntity<Map<String, Object>> handleInvalidJson(
            InvalidJsonException ex, HttpServletRequest request) {
        Map<String, Object> body = new HashMap<>();
        body.put("status", HttpStatus.BAD_REQUEST.value());
        body.put("error", "Bad Request");
        body.put("message", ex.getErrorMessage());
        body.put("path", request.getRequestURI());
        body.put("timestamp", java.time.Instant.now().toString());
        if (ex.getErrorLine() > 0) {
            body.put("errorLine", ex.getErrorLine());
        }
        if (ex.getErrorColumn() > 0) {
            body.put("errorColumn", ex.getErrorColumn());
        }
        if (ex.getErrorLine() > 0 || ex.getErrorColumn() > 0) {
            body.put("detail", String.format("JSON parsing failed at line %d, column %d: %s",
                    ex.getErrorLine(), ex.getErrorColumn(), ex.getErrorMessage()));
        }

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(body);
    }

    @ExceptionHandler(QueueFullException.class)
    public ResponseEntity<Map<String, Object>> handleQueueFull(
            QueueFullException ex, HttpServletRequest request) {
        Map<String, Object> body = new HashMap<>();
        body.put("status", HttpStatus.SERVICE_UNAVAILABLE.value());
        body.put("error", "Service Unavailable");
        body.put("message", String.format(
                "Topic '%s' queue is full and no space became available after waiting %d seconds. " +
                "Current queue size: %d, Max: %d. Please retry later or implement fallback logic.",
                ex.getTopic(), ex.getWaitTimeoutSeconds(), ex.getCurrentSize(), ex.getMaxSize()));
        body.put("topic", ex.getTopic());
        body.put("currentSize", ex.getCurrentSize());
        body.put("maxSize", ex.getMaxSize());
        body.put("waitTimeoutSeconds", ex.getWaitTimeoutSeconds());
        body.put("path", request.getRequestURI());
        body.put("timestamp", java.time.Instant.now().toString());

        return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body(body);
    }

    @ExceptionHandler(TopicNotFoundException.class)
    public ResponseEntity<Map<String, Object>> handleTopicNotFound(
            TopicNotFoundException ex, HttpServletRequest request) {
        Map<String, Object> body = new HashMap<>();
        body.put("status", HttpStatus.NOT_FOUND.value());
        body.put("error", "Not Found");
        body.put("message", String.format("Topic '%s' does not exist. Cannot commit offset for non-existent topic.",
                ex.getTopic()));
        body.put("topic", ex.getTopic());
        body.put("path", request.getRequestURI());
        body.put("timestamp", java.time.Instant.now().toString());

        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(body);
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<Map<String, Object>> handleValidationExceptions(
            MethodArgumentNotValidException ex, HttpServletRequest request) {
        Map<String, Object> body = new HashMap<>();
        body.put("status", HttpStatus.BAD_REQUEST.value());
        body.put("error", "Bad Request");

        StringBuilder message = new StringBuilder("Validation failed: ");
        ex.getBindingResult().getFieldErrors().forEach(error -> {
            message.append(error.getField()).append(" ").append(error.getDefaultMessage()).append("; ");
        });
        body.put("message", message.toString());
        body.put("path", request.getRequestURI());
        body.put("timestamp", java.time.Instant.now().toString());

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(body);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGenericException(
            Exception ex, HttpServletRequest request) {
        ErrorResponse error = new ErrorResponse(
                HttpStatus.INTERNAL_SERVER_ERROR.value(),
                "Internal Server Error",
                ex.getMessage() != null ? ex.getMessage() : "An unexpected error occurred",
                request.getRequestURI()
        );
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(error);
    }
}
