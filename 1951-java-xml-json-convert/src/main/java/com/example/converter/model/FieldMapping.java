package com.example.converter.model;

import lombok.Data;

@Data
public class FieldMapping {
    private String sourceName;
    private String targetName;
    private FieldType type;
    private Object defaultValue;
    private Boolean isAttribute = false;

    public enum FieldType {
        STRING, INTEGER, LONG, DOUBLE, BOOLEAN, NUMBER
    }
}
