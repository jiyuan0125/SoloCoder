package com.edi.converter.exception;

import java.util.ArrayList;
import java.util.List;

public class EdiValidationException extends RuntimeException {
    
    private List<ValidationError> errors = new ArrayList<>();
    
    public EdiValidationException(String message) {
        super(message);
    }
    
    public EdiValidationException(String message, List<ValidationError> errors) {
        super(message);
        this.errors = errors != null ? errors : new ArrayList<>();
    }
    
    public List<ValidationError> getErrors() {
        return errors;
    }
    
    public void addError(ValidationError error) {
        this.errors.add(error);
    }
    
    public static class ValidationError {
        private String segmentTag;
        private Integer segmentPosition;
        private Integer elementPosition;
        private Integer componentPosition;
        private String field;
        private String expectedValue;
        private String actualValue;
        private String message;
        
        public ValidationError() {}
        
        public ValidationError(String message) {
            this.message = message;
        }

        public String getSegmentTag() {
            return segmentTag;
        }

        public void setSegmentTag(String segmentTag) {
            this.segmentTag = segmentTag;
        }

        public Integer getSegmentPosition() {
            return segmentPosition;
        }

        public void setSegmentPosition(Integer segmentPosition) {
            this.segmentPosition = segmentPosition;
        }

        public Integer getElementPosition() {
            return elementPosition;
        }

        public void setElementPosition(Integer elementPosition) {
            this.elementPosition = elementPosition;
        }

        public Integer getComponentPosition() {
            return componentPosition;
        }

        public void setComponentPosition(Integer componentPosition) {
            this.componentPosition = componentPosition;
        }

        public String getField() {
            return field;
        }

        public void setField(String field) {
            this.field = field;
        }

        public String getExpectedValue() {
            return expectedValue;
        }

        public void setExpectedValue(String expectedValue) {
            this.expectedValue = expectedValue;
        }

        public String getActualValue() {
            return actualValue;
        }

        public void setActualValue(String actualValue) {
            this.actualValue = actualValue;
        }

        public String getMessage() {
            return message;
        }

        public void setMessage(String message) {
            this.message = message;
        }
    }
}
