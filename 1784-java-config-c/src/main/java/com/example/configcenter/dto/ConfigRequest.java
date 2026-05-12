package com.example.configcenter.dto;

import lombok.Data;

@Data
public class ConfigRequest {
    private String key;
    private String value;
}
