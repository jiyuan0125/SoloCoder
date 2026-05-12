package com.logaggregate.service.storage;

import com.logaggregate.model.StoredLogEntry;

import java.util.List;
import java.util.Set;

public interface LogStorageService {
    void storeLogs(List<StoredLogEntry> logs);
    List<StoredLogEntry> queryLogs(String service, List<String> levels, Long startTimeUtcMs, Long endTimeUtcMs, String keyword);
    void deleteLogsByServiceBefore(String service, long beforeUtcMs);
    Set<String> getAllServices();
}
