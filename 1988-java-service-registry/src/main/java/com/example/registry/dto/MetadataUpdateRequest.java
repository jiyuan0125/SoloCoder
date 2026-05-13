package com.example.registry.dto;

import lombok.Data;

import java.util.Map;

@Data
public class MetadataUpdateRequest {
    
    private Map<String, String> metadata;
}
