package com.healthchecker.dto;

import com.healthchecker.entity.ServiceStatus;
import java.time.LocalDateTime;

public record ServiceStatusSummary(
        String serviceName,
        ServiceStatus status,
        LocalDateTime lastCheckTime,
        int consecutiveSuccess,
        int consecutiveFailure
) {}