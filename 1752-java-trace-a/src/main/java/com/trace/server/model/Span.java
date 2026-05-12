package com.trace.server.model;

import lombok.Data;
import java.util.List;
import java.util.Map;

@Data
public class Span {
    private String spanId;
    private String parentSpanId;
    private String traceId;
    private String serviceId;
    private String serviceName;
    private String operationName;
    private Long startTime;
    private Long endTime;
    private Map<String, String> tags;
    private Boolean isError;
    private Boolean archived = false;
}
