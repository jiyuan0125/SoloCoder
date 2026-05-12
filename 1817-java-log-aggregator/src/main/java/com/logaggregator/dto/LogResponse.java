package com.logaggregator.dto;

import com.logaggregator.model.LogLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LogResponse {

    private Long id;
    private String service;
    private LogLevel level;
    private String message;
    private String highlightedMessage;
    private LocalDateTime timestamp;
    private String traceId;
    private String spanId;
}
