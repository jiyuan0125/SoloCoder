package com.logaggregate.service.ingest;

import com.logaggregate.model.LogEntry;
import com.logaggregate.model.StoredLogEntry;
import com.logaggregate.service.storage.LogStorageService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class IngestService {

    private static final Logger logger = LoggerFactory.getLogger(IngestService.class);
    private static final int MAX_BATCH_SIZE = 1000;

    private final LogStorageService logStorageService;

    public IngestService(LogStorageService logStorageService) {
        this.logStorageService = logStorageService;
    }

    public IngestResult ingestSingle(LogEntry entry) {
        List<LogEntry> singleList = new ArrayList<>();
        singleList.add(entry);
        return ingestBatch(singleList);
    }

    public IngestResult ingestBatch(List<LogEntry> entries) {
        boolean truncated = false;
        int originalSize = entries.size();
        List<LogEntry> toProcess = entries;

        if (entries.size() > MAX_BATCH_SIZE) {
            toProcess = entries.subList(0, MAX_BATCH_SIZE);
            truncated = true;
        }

        List<StoredLogEntry> validLogs = new ArrayList<>();
        int failed = 0;

        for (LogEntry entry : toProcess) {
            try {
                validateRequiredFields(entry);
                long timestampUtcMs = TimestampParser.parseToUtcMillis(entry.getTimestamp());
                StoredLogEntry stored = new StoredLogEntry(
                        timestampUtcMs,
                        entry.getService(),
                        entry.getLevel(),
                        entry.getMessage()
                );
                validLogs.add(stored);
            } catch (Exception e) {
                failed++;
                logger.warn("Failed to ingest log entry: {}", e.getMessage());
            }
        }

        try {
            logStorageService.storeLogs(validLogs);
        } catch (Exception e) {
            logger.error("Failed to store {} logs: {}", validLogs.size(), e.getMessage(), e);
            return new IngestResult(0, originalSize, validLogs.size(), true, truncated, "Storage error occurred");
        }

        String warning = null;
        if (truncated) {
            warning = "Batch exceeded max size of " + MAX_BATCH_SIZE + ", truncated to first " + MAX_BATCH_SIZE + " entries";
        }

        return new IngestResult(validLogs.size(), originalSize, failed, false, truncated, warning);
    }

    private void validateRequiredFields(LogEntry entry) {
        if (entry.getTimestamp() == null) {
            throw new IllegalArgumentException("timestamp is required");
        }
        if (entry.getService() == null || entry.getService().isBlank()) {
            throw new IllegalArgumentException("service is required");
        }
        if (entry.getLevel() == null || entry.getLevel().isBlank()) {
            throw new IllegalArgumentException("level is required");
        }
        if (entry.getMessage() == null || entry.getMessage().isBlank()) {
            throw new IllegalArgumentException("message is required");
        }
    }

    public static class IngestResult {
        private final int ingested;
        private final int received;
        private final int failed;
        private final boolean storageError;
        private final boolean truncated;
        private final String warning;

        public IngestResult(int ingested, int received, int failed, boolean storageError, boolean truncated, String warning) {
            this.ingested = ingested;
            this.received = received;
            this.failed = failed;
            this.storageError = storageError;
            this.truncated = truncated;
            this.warning = warning;
        }

        public int getIngested() {
            return ingested;
        }

        public int getReceived() {
            return received;
        }

        public int getFailed() {
            return failed;
        }

        public boolean isStorageError() {
            return storageError;
        }

        public boolean isTruncated() {
            return truncated;
        }

        public String getWarning() {
            return warning;
        }
    }
}
