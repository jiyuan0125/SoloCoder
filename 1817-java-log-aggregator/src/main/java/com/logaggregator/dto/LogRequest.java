package com.logaggregator.dto;

import com.logaggregator.model.LogLevel;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LogRequest {

    @NotBlank(message = "service is required")
    private String service;

    @NotNull(message = "level is required")
    private LogLevel level;

    @NotBlank(message = "message is required")
    private String message;

    @NotNull(message = "timestamp is required")
    private LocalDateTime timestamp;

    private String traceId;

    private String spanId;
}
