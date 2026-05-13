package com.edi.converter.exception;

import com.edi.converter.dto.ErrorResponse;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import java.util.stream.Collectors;

@ControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(EdiParseException.class)
    public ResponseEntity<ErrorResponse> handleEdiParseException(EdiParseException ex) {
        ErrorResponse response = new ErrorResponse();
        response.setMessage(ex.getMessage());
        response.setType("PARSE_ERROR");
        
        ErrorResponse.ValidationErrorDetail detail = new ErrorResponse.ValidationErrorDetail();
        detail.setSegmentTag(ex.getSegmentTag());
        detail.setSegmentPosition(ex.getSegmentPosition());
        detail.setElementPosition(ex.getElementPosition());
        detail.setComponentPosition(ex.getComponentPosition());
        detail.setMessage(ex.getMessage());
        
        response.getErrors().add(detail);
        
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(response);
    }

    @ExceptionHandler(EdiValidationException.class)
    public ResponseEntity<ErrorResponse> handleEdiValidationException(EdiValidationException ex) {
        ErrorResponse response = new ErrorResponse();
        response.setMessage(ex.getMessage());
        response.setType("VALIDATION_ERROR");
        
        response.setErrors(ex.getErrors().stream()
                .map(this::toDetail)
                .collect(Collectors.toList()));
        
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(response);
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<ErrorResponse> handleIllegalArgumentException(IllegalArgumentException ex) {
        ErrorResponse response = new ErrorResponse();
        response.setMessage(ex.getMessage());
        response.setType("BAD_REQUEST");
        
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(response);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleException(Exception ex) {
        ErrorResponse response = new ErrorResponse();
        response.setMessage("服务器内部错误: " + ex.getMessage());
        response.setType("INTERNAL_ERROR");
        
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
    }

    private ErrorResponse.ValidationErrorDetail toDetail(EdiValidationException.ValidationError error) {
        ErrorResponse.ValidationErrorDetail detail = new ErrorResponse.ValidationErrorDetail();
        detail.setSegmentTag(error.getSegmentTag());
        detail.setSegmentPosition(error.getSegmentPosition());
        detail.setElementPosition(error.getElementPosition());
        detail.setComponentPosition(error.getComponentPosition());
        detail.setField(error.getField());
        detail.setExpectedValue(error.getExpectedValue());
        detail.setActualValue(error.getActualValue());
        detail.setMessage(error.getMessage());
        return detail;
    }
}
