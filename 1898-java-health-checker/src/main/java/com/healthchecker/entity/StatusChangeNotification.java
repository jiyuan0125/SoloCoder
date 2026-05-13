package com.healthchecker.entity;

import java.time.LocalDateTime;

public record StatusChangeNotification(
        String serviceName,
        ServiceStatus oldStatus,
        ServiceStatus newStatus,
        LocalDateTime changeTime,
        int consecutiveChecks
) {}