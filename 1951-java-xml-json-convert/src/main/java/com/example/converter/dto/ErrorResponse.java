package com.example.converter.dto;

import lombok.Data;
import java.util.List;

@Data
public class ErrorResponse {
    private String error;
    private String message;
    private List<FieldError> fieldErrors;
}
