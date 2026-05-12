package com.example.configcenter.dto;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class ConfigResponse {
    private String key;
    private String value;
    private Integer version;
    private String message;
}
