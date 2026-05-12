package com.logcollector.service;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.logcollector.config.LogCollectorProperties;
import com.logcollector.model.LogEntry;
import com.logcollector.model.LogQueryParam;
import com.logcollector.model.RawLogEntry;
import com.logcollector.util.TimestampParser;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.time.Instant;
import java.time.LocalDate;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.NavigableMap;
import java.util.Set;
import java.util.TreeMap;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Service
public class LogStorageService {

    private static final Logger logger = LoggerFactory.getLogger(LogStorageService.class);

    private static final Set<String> VALID_LEVELS = Set.of("DEBUG", "INFO", "WARN", "ERROR");
    private static final String DEFAULT_LEVEL = "INFO";
    private static final DateTimeFormatter FILE_DATE_FORMAT = DateTimeFormatter.ofPattern("yyyy-MM-dd");

    private final LogCollectorProperties properties;
    private final ObjectMapper objectMapper;

    private final ConcurrentMap<String, NavigableMap<Long, List<LogEntry>>> serviceTimeIndex = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, List<LogEntry>> requestIdIndex = new ConcurrentHashMap<>();
    private final ReadWriteLock lock = new ReentrantReadWriteLock();

    private Path logDirectory;

    public LogStorageService(LogCollectorProperties properties, ObjectMapper objectMapper) {
        this.properties = properties;
        this.objectMapper = objectMapper;
    }

    @PostConstruct
    public void init() {
        logDirectory = Paths.get(properties.getLogDirectory()).toAbsolutePath().normalize();
        try {
            if (!Files.exists(logDirectory)) {
                Files.createDirectories(logDirectory);
                logger.info("Created log directory: {}", logDirectory);
            }
        } catch (IOException e) {
            logger.error("Failed to create log directory: {}", logDirectory, e);
        }

        loadRecentLogs();
    }

    private void loadRecentLogs() {
        int days = properties.getLoadDays();
        LocalDate today = LocalDate.now();
        lock.writeLock().lock();
        try {
            serviceTimeIndex.clear();
            requestIdIndex.clear();
            for (int i = 0; i < days; i++) {
                LocalDate date = today.minusDays(i);
                loadLogFile(date);
            }
            logger.info("Loaded logs for last {} days", days);
        } finally {
            lock.writeLock().unlock();
        }
    }

    private void loadLogFile(LocalDate date) {
        String fileName = properties.getFilePrefix() + "-" + FILE_DATE_FORMAT.format(date) + ".log";
        Path filePath = logDirectory.resolve(fileName);
        if (!Files.exists(filePath)) {
            return;
        }

        logger.info("Loading log file: {}", filePath);
        try (BufferedReader reader = Files.newBufferedReader(filePath)) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (line.trim().isEmpty()) {
                    continue;
                }
                try {
                    LogEntry entry = objectMapper.readValue(line, LogEntry.class);
                    indexLogEntry(entry);
                } catch (Exception e) {
                    logger.warn("Failed to parse log line: {}", line, e);
                }
            }
        } catch (IOException e) {
            logger.error("Failed to read log file: {}", filePath, e);
        }
    }

    public LogEntry save(RawLogEntry raw) {
        LogEntry entry = normalize(raw);
        persist(entry);
        lock.writeLock().lock();
        try {
            indexLogEntry(entry);
        } finally {
            lock.writeLock().unlock();
        }
        return entry;
    }

    private LogEntry normalize(RawLogEntry raw) {
        LogEntry entry = new LogEntry();

        Long ts = TimestampParser.parse(raw.getTimestamp());
        if (ts == null) {
            ts = System.currentTimeMillis();
        }
        entry.setTimestamp(ts);

        String service = raw.getService();
        if (service == null || service.trim().isEmpty()) {
            service = "unknown";
        }
        entry.setService(service);

        String level = raw.getLevel();
        if (level == null || !VALID_LEVELS.contains(level.toUpperCase())) {
            level = DEFAULT_LEVEL;
        } else {
            level = level.toUpperCase();
        }
        entry.setLevel(level);

        String message = raw.getMessage();
        if (message == null) {
            message = "";
        }
        entry.setMessage(message);

        if (raw.getRequestId() != null && !raw.getRequestId().trim().isEmpty()) {
            entry.setRequestId(raw.getRequestId().trim());
        }

        if (raw.getAdditionalProperties() != null && !raw.getAdditionalProperties().isEmpty()) {
            Map<String, Object> extra = new HashMap<>(raw.getAdditionalProperties());
            extra.remove("timestamp");
            extra.remove("service");
            extra.remove("level");
            extra.remove("message");
            extra.remove("request-id");
            if (!extra.isEmpty()) {
                entry.setExtra(extra);
            }
        }

        return entry;
    }

    private void persist(LogEntry entry) {
        LocalDate date = Instant.ofEpochMilli(entry.getTimestamp()).atZone(ZoneId.of("UTC")).toLocalDate();
        String fileName = properties.getFilePrefix() + "-" + FILE_DATE_FORMAT.format(date) + ".log";
        Path filePath = logDirectory.resolve(fileName);

        try {
            if (!Files.exists(filePath.getParent())) {
                Files.createDirectories(filePath.getParent());
            }
            String json = objectMapper.writeValueAsString(entry);
            try (BufferedWriter writer = Files.newBufferedWriter(filePath,
                    StandardOpenOption.CREATE, StandardOpenOption.APPEND, StandardOpenOption.WRITE)) {
                writer.write(json);
                writer.newLine();
            }
        } catch (IOException e) {
            logger.error("Failed to persist log entry: {}", entry, e);
        }
    }

    private void indexLogEntry(LogEntry entry) {
        NavigableMap<Long, List<LogEntry>> timeMap = serviceTimeIndex
                .computeIfAbsent(entry.getService(), k -> Collections.synchronizedNavigableMap(new TreeMap<>()));
        timeMap.computeIfAbsent(entry.getTimestamp(), k -> Collections.synchronizedList(new ArrayList<>())).add(entry);

        if (entry.getRequestId() != null) {
            requestIdIndex
                    .computeIfAbsent(entry.getRequestId(), k -> Collections.synchronizedList(new ArrayList<>()))
                    .add(entry);
        }
    }

    public List<LogEntry> query(LogQueryParam param) {
        lock.readLock().lock();
        try {
            List<LogEntry> candidates;
            if (param.getService() != null && !param.getService().trim().isEmpty()) {
                NavigableMap<Long, List<LogEntry>> timeMap = serviceTimeIndex.get(param.getService().trim());
                if (timeMap == null) {
                    return Collections.emptyList();
                }
                candidates = filterByTime(timeMap, param.getStartTime(), param.getEndTime());
            } else {
                candidates = serviceTimeIndex.values().stream()
                        .flatMap(tm -> filterByTime(tm, param.getStartTime(), param.getEndTime()).stream())
                        .collect(Collectors.toList());
            }

            final Set<String> levelFilter;
            if (param.getLevels() != null && !param.getLevels().isEmpty()) {
                levelFilter = param.getLevels().stream()
                        .map(String::toUpperCase)
                        .filter(VALID_LEVELS::contains)
                        .collect(Collectors.toCollection(HashSet::new));
            } else {
                levelFilter = null;
            }

            List<LogEntry> result = candidates.stream()
                    .filter(e -> levelFilter == null || levelFilter.contains(e.getLevel()))
                    .sorted(Comparator.comparingLong(LogEntry::getTimestamp).reversed())
                    .collect(Collectors.toList());

            int page = Math.max(param.getPage(), 0);
            int pageSize = Math.max(param.getPageSize(), 1);
            int from = page * pageSize;
            int to = Math.min(from + pageSize, result.size());
            if (from >= result.size()) {
                return Collections.emptyList();
            }
            return result.subList(from, to);
        } finally {
            lock.readLock().unlock();
        }
    }

    private List<LogEntry> filterByTime(NavigableMap<Long, List<LogEntry>> timeMap, Long start, Long end) {
        if (start == null && end == null) {
            return timeMap.values().stream()
                    .flatMap(Collection::stream)
                    .collect(Collectors.toList());
        }

        Long fromKey = start != null ? start : Long.MIN_VALUE;
        Long toKey = end != null ? end : Long.MAX_VALUE;

        return timeMap.subMap(fromKey, true, toKey, true).values().stream()
                .flatMap(Collection::stream)
                .collect(Collectors.toList());
    }

    public List<LogEntry> getTrace(String requestId) {
        lock.readLock().lock();
        try {
            List<LogEntry> entries = requestIdIndex.get(requestId);
            if (entries == null) {
                return Collections.emptyList();
            }
            return entries.stream()
                    .sorted(Comparator.comparingLong(LogEntry::getTimestamp))
                    .collect(Collectors.toList());
        } finally {
            lock.readLock().unlock();
        }
    }
}
