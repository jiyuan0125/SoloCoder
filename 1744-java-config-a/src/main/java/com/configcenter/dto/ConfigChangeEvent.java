package com.configcenter.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.Map;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ConfigChangeEvent {
    private String projectId;
    private String projectName;
    private String environment;
    private String version;
    private Map<String, String> changes;
    private boolean isGrayRelease;
    private LocalDateTime timestamp;
}
