package com.example.converter.exception;

import com.example.converter.dto.FieldError;
import lombok.Getter;
import java.util.List;

@Getter
public class ValidationException extends RuntimeException {
    private final List<FieldError> fieldErrors;

    public ValidationException(String message, List<FieldError> fieldErrors) {
        super(message);
        this.fieldErrors = fieldErrors;
    }

    public ValidationException(String message) {
        super(message);
        this.fieldErrors = null;
    }
}
