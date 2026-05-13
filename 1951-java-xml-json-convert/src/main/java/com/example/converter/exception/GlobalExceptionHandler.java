package com.example.converter.exception;

import com.example.converter.dto.ErrorResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

@ControllerAdvice
@Slf4j
public class GlobalExceptionHandler {

    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<ErrorResponse> handleValidationException(ValidationException e) {
        log.error("Validation error: {}", e.getMessage(), e);

        ErrorResponse response = new ErrorResponse();
        response.setError("VALIDATION_ERROR");
        response.setMessage(e.getMessage());
        response.setFieldErrors(e.getFieldErrors());

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(response);
    }

    @ExceptionHandler(MappingNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleMappingNotFoundException(MappingNotFoundException e) {
        log.error("Mapping not found: {}", e.getMessage(), e);

        ErrorResponse response = new ErrorResponse();
        response.setError("MAPPING_NOT_FOUND");
        response.setMessage(e.getMessage());

        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(response);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGenericException(Exception e) {
        log.error("Unexpected error: {}", e.getMessage(), e);

        ErrorResponse response = new ErrorResponse();
        response.setError("INTERNAL_ERROR");
        response.setMessage("An unexpected error occurred: " + e.getMessage());

        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
    }
}
