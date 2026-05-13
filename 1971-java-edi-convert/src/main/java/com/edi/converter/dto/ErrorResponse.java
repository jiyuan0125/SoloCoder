package com.edi.converter.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ErrorResponse {
    private String message;
    private String type;
    private List<ValidationErrorDetail> errors = new ArrayList<>();

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class ValidationErrorDetail {
        private String segmentTag;
        private Integer segmentPosition;
        private Integer elementPosition;
        private Integer componentPosition;
        private String field;
        private String expectedValue;
        private String actualValue;
        private String message;
    }
}
