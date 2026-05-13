package com.trace.collector.model;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonInclude;
import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.Map;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Span {

    @NotBlank(message = "trace_id is required")
    private String traceId;

    @NotBlank(message = "span_id is required")
    private String spanId;

    private String parentSpanId;

    @NotBlank(message = "service is required")
    private String service;

    @NotBlank(message = "operation is required")
    private String operation;

    @NotNull(message = "start_time is required")
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", timezone = "UTC")
    private java.time.Instant startTime;

    @NotNull(message = "duration is required")
    @Min(value = 0, message = "duration must be non-negative")
    private Long duration;

    @NotBlank(message = "status is required")
    private String status;

    private Map<String, String> tags;

    private Double percentage;

    public boolean isSlow() {
        return duration != null && duration >= 1000;
    }
}
