package com.servicemesh.common.model;

import java.util.Objects;

public class ServiceEdge {
    private String source;
    private String target;
    private long qps;
    private long totalRequests;
    private long errorCount;
    private double errorRate;

    public ServiceEdge() {}

    public ServiceEdge(String source, String target) {
        this.source = source;
        this.target = target;
    }

    public void update(boolean success) {
        this.totalRequests++;
        if (!success) {
            this.errorCount++;
        }
        this.errorRate = totalRequests > 0 ? (double) errorCount / totalRequests * 100 : 0.0;
    }

    public void resetQps() {
        this.qps = 0;
    }

    public String getSource() { return source; }
    public void setSource(String source) { this.source = source; }
    public String getTarget() { return target; }
    public void setTarget(String target) { this.target = target; }
    public long getQps() { return qps; }
    public void setQps(long qps) { this.qps = qps; }
    public long getTotalRequests() { return totalRequests; }
    public void setTotalRequests(long totalRequests) { this.totalRequests = totalRequests; }
    public long getErrorCount() { return errorCount; }
    public void setErrorCount(long errorCount) { this.errorCount = errorCount; }
    public double getErrorRate() { return errorRate; }
    public void setErrorRate(double errorRate) { this.errorRate = errorRate; }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ServiceEdge that = (ServiceEdge) o;
        return Objects.equals(source, that.source) && Objects.equals(target, that.target);
    }

    @Override
    public int hashCode() {
        return Objects.hash(source, target);
    }
}
