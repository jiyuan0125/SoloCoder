package com.example.configcenter.dto;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class ConfigNotification {
    private String action;
    private String key;
    private Integer version;
}
