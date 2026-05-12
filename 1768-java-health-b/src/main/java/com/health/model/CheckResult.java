package com.health.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CheckResult {
    private boolean healthy;
    private Integer responseTimeMs;
    private String message;
    private LocalDateTime checkedAt;
}