package com.example.gateway.stats;

public class VersionStats {
    private String version;
    private long totalCalls;
    private long totalResponseTimeMs;
    private long clientErrorCount;
    private long serverErrorCount;

    public VersionStats(String version) {
        this.version = version;
    }

    public String getVersion() {
        return version;
    }

    public long getTotalCalls() {
        return totalCalls;
    }

    public long getTotalResponseTimeMs() {
        return totalResponseTimeMs;
    }

    public long getClientErrorCount() {
        return clientErrorCount;
    }

    public long getServerErrorCount() {
        return serverErrorCount;
    }

    public void setTotalCalls(long totalCalls) {
        this.totalCalls = totalCalls;
    }

    public void setTotalResponseTimeMs(long totalResponseTimeMs) {
        this.totalResponseTimeMs = totalResponseTimeMs;
    }

    public void setClientErrorCount(long clientErrorCount) {
        this.clientErrorCount = clientErrorCount;
    }

    public void setServerErrorCount(long serverErrorCount) {
        this.serverErrorCount = serverErrorCount;
    }

    public double getAvgResponseTimeMs() {
        if (totalCalls == 0) {
            return 0.0;
        }
        return (double) totalResponseTimeMs / totalCalls;
    }

    public double getClientErrorRate() {
        if (totalCalls == 0) {
            return 0.0;
        }
        return (double) clientErrorCount / totalCalls;
    }

    public double getServerErrorRate() {
        if (totalCalls == 0) {
            return 0.0;
        }
        return (double) serverErrorCount / totalCalls;
    }

    public double getErrorRate() {
        if (totalCalls == 0) {
            return 0.0;
        }
        return (double) (clientErrorCount + serverErrorCount) / totalCalls;
    }
}
