package com.configcenter.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class AuditLog {
    private String id;
    private String userId;
    private String appName;
    private AuditAction action;
    private String oldValue;
    private String newValue;
    private String configKey;
    private Instant timestamp;
}
