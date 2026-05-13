package com.trace.collector.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;

import java.util.List;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class TraceTree {
    private String traceId;
    private List<SpanNode> spans;
}
