package com.example.cachemiddleware.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class KeyChangeEvent {
    private String namespace;
    private String key;
    private String changeType;
    private Object oldValue;
    private Object newValue;
    private long timestamp;
}
