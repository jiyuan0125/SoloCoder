package com.example.gateway.compatibility;

public class FieldMapping {
    private String from;
    private String to;
    private String type;
    private Object defaultValue;

    public FieldMapping() {
    }

    public FieldMapping(String from, String to, String type, Object defaultValue) {
        this.from = from;
        this.to = to;
        this.type = type;
        this.defaultValue = defaultValue;
    }

    public String getFrom() {
        return from;
    }

    public void setFrom(String from) {
        this.from = from;
    }

    public String getTo() {
        return to;
    }

    public void setTo(String to) {
        this.to = to;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public Object getDefaultValue() {
        return defaultValue;
    }

    public void setDefaultValue(Object defaultValue) {
        this.defaultValue = defaultValue;
    }
}
