package com.trace.collector.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;

import java.util.List;
import java.util.Map;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class SpanNode {
    private String traceId;
    private String spanId;
    private String parentSpanId;
    private String service;
    private String operation;
    private java.time.Instant startTime;
    private Long duration;
    private String status;
    private Map<String, String> tags;
    private Boolean slow;
    private Double percentage;
    private List<SpanNode> children;
}
