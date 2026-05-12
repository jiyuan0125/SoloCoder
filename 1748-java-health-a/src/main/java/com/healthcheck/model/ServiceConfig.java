package com.healthcheck.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ServiceConfig {
    private String serviceId;
    private String serviceName;
    private String description;
    private Integer checkInterval = 30;
    private String webhookUrl;
    private List<CheckItemConfig> checkItems;
    private boolean enabled = true;
}
