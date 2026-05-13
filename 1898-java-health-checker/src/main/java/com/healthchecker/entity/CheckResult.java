package com.healthchecker.entity;

import java.time.LocalDateTime;

public record CheckResult(
        LocalDateTime timestamp,
        long durationMs,
        int statusCode,
        boolean success
) {}