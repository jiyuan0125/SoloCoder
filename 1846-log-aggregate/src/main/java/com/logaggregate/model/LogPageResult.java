package com.logaggregate.model;

import java.util.List;

public class LogPageResult {
    private List<StoredLogEntry> logs;
    private long total;
    private int page;
    private int pageSize;

    public LogPageResult(List<StoredLogEntry> logs, long total, int page, int pageSize) {
        this.logs = logs;
        this.total = total;
        this.page = page;
        this.pageSize = pageSize;
    }

    public List<StoredLogEntry> getLogs() {
        return logs;
    }

    public void setLogs(List<StoredLogEntry> logs) {
        this.logs = logs;
    }

    public long getTotal() {
        return total;
    }

    public void setTotal(long total) {
        this.total = total;
    }

    public int getPage() {
        return page;
    }

    public void setPage(int page) {
        this.page = page;
    }

    public int getPageSize() {
        return pageSize;
    }

    public void setPageSize(int pageSize) {
        this.pageSize = pageSize;
    }
}
