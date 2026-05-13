package com.trace.collector.model;

import java.util.ArrayList;
import java.util.List;

public class TraceNode {

    private String serviceName;
    private String operationName;
    private Long durationMs;
    private String statusCode;
    private String spanId;
    private List<TraceNode> children;

    public TraceNode(Span span) {
        this.serviceName = span.getServiceName();
        this.operationName = span.getOperationName();
        this.durationMs = span.getDurationMs();
        this.statusCode = span.getStatusCode();
        this.spanId = span.getSpanId();
        this.children = new ArrayList<>();
    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public String getOperationName() {
        return operationName;
    }

    public void setOperationName(String operationName) {
        this.operationName = operationName;
    }

    public Long getDurationMs() {
        return durationMs;
    }

    public void setDurationMs(Long durationMs) {
        this.durationMs = durationMs;
    }

    public String getStatusCode() {
        return statusCode;
    }

    public void setStatusCode(String statusCode) {
        this.statusCode = statusCode;
    }

    public String getSpanId() {
        return spanId;
    }

    public void setSpanId(String spanId) {
        this.spanId = spanId;
    }

    public List<TraceNode> getChildren() {
        return children;
    }

    public void setChildren(List<TraceNode> children) {
        this.children = children;
    }

    public void addChild(TraceNode child) {
        this.children.add(child);
    }
}
