package com.logaggregator.service;

import com.logaggregator.config.LogAggregatorProperties;
import com.logaggregator.dto.*;
import com.logaggregator.model.LogEntry;
import com.logaggregator.model.LogLevel;
import com.logaggregator.repository.LogEntryRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.nio.charset.StandardCharsets;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Slf4j
@Service
@RequiredArgsConstructor
public class LogService {

    private static final String TRUNCATED_MARKER = "...[truncated]";
    private static final DateTimeFormatter HOUR_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:00");

    private final LogEntryRepository logEntryRepository;
    private final LogAggregatorProperties properties;

    @Transactional
    public BatchLogResponse batchSave(BatchLogRequest request) {
        List<LogEntry> entries = new ArrayList<>(request.getLogs().size());

        for (LogRequest logReq : request.getLogs()) {
            LogEntry entry = LogEntry.builder()
                    .service(logReq.getService())
                    .level(logReq.getLevel())
                    .message(truncateMessage(logReq.getMessage()))
                    .timestamp(logReq.getTimestamp())
                    .traceId(logReq.getTraceId())
                    .spanId(logReq.getSpanId())
                    .build();
            entries.add(entry);
        }

        logEntryRepository.saveAll(entries);

        log.info("Successfully saved {} log entries", entries.size());

        return BatchLogResponse.builder()
                .received(entries.size())
                .message("Logs received successfully")
                .build();
    }

    private String truncateMessage(String message) {
        if (message == null) {
            return "";
        }

        byte[] bytes = message.getBytes(StandardCharsets.UTF_8);
        int maxSize = properties.getMaxMessageSize();

        if (bytes.length <= maxSize) {
            return message;
        }

        byte[] markerBytes = TRUNCATED_MARKER.getBytes(StandardCharsets.UTF_8);
        int availableBytes = maxSize - markerBytes.length;

        if (availableBytes <= 0) {
            return TRUNCATED_MARKER;
        }

        String truncated = new String(bytes, 0, availableBytes, StandardCharsets.UTF_8);
        return truncated + TRUNCATED_MARKER;
    }

    public PageResponse<LogResponse> queryLogs(LogQueryRequest request) {
        int page = request.getPage() != null ? request.getPage() : 0;
        int size = request.getSize() != null ? request.getSize() : properties.getDefaultPageSize();

        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "timestamp"));

        LocalDateTime startTime = request.getStartTime() != null
                ? request.getStartTime()
                : LocalDateTime.now().minusDays(7);
        LocalDateTime endTime = request.getEndTime() != null
                ? request.getEndTime()
                : LocalDateTime.now();

        Page<LogEntry> logPage;

        if (request.getService() != null && !request.getService().isEmpty()
                && request.getLevels() != null && !request.getLevels().isEmpty()) {
            logPage = logEntryRepository.findByServiceAndLevelInAndTimestampBetween(
                    request.getService(), request.getLevels(), startTime, endTime, pageable);
        } else if (request.getService() != null && !request.getService().isEmpty()) {
            logPage = logEntryRepository.findByServiceAndTimestampBetween(
                    request.getService(), startTime, endTime, pageable);
        } else if (request.getLevels() != null && !request.getLevels().isEmpty()) {
            logPage = logEntryRepository.findByLevelInAndTimestampBetween(
                    request.getLevels(), startTime, endTime, pageable);
        } else {
            logPage = logEntryRepository.findAll(pageable);
        }

        List<LogResponse> content = logPage.getContent().stream()
                .map(entry -> toResponse(entry, request.getKeyword()))
                .filter(response -> {
                    if (request.getKeyword() == null || request.getKeyword().isEmpty()) {
                        return true;
                    }
                    return response.getMessage().toLowerCase()
                            .contains(request.getKeyword().toLowerCase());
                })
                .toList();

        return PageResponse.<LogResponse>builder()
                .content(content)
                .page(logPage.getNumber())
                .size(logPage.getSize())
                .totalElements(logPage.getTotalElements())
                .totalPages(logPage.getTotalPages())
                .hasNext(logPage.hasNext())
                .hasPrevious(logPage.hasPrevious())
                .build();
    }

    public List<LogResponse> getTraceChain(String traceId) {
        List<LogEntry> entries = logEntryRepository.findByTraceIdOrderByTimestampAsc(traceId);
        return entries.stream()
                .map(entry -> toResponse(entry, null))
                .toList();
    }

    public AggregationStats getAggregationStats() {
        Map<String, Long> serviceDist = new LinkedHashMap<>();
        List<Object[]> serviceResults = logEntryRepository.countByService();
        for (Object[] row : serviceResults) {
            serviceDist.put((String) row[0], (Long) row[1]);
        }

        Map<LogLevel, Long> levelDist = new EnumMap<>(LogLevel.class);
        List<Object[]> levelResults = logEntryRepository.countByLevel();
        for (Object[] row : levelResults) {
            levelDist.put((LogLevel) row[0], (Long) row[1]);
        }

        Map<String, Long> errorTrend = new LinkedHashMap<>();
        LocalDateTime since = LocalDateTime.now().minusHours(1);
        List<Object[]> errorResults = logEntryRepository.countErrorsByHour(since);
        for (Object[] row : errorResults) {
            errorTrend.put((String) row[0], (Long) row[1]);
        }

        LocalDateTime hourStart = LocalDateTime.now().withMinute(0).withSecond(0).withNano(0);
        String currentHour = hourStart.format(HOUR_FORMATTER);
        errorTrend.putIfAbsent(currentHour, 0L);

        return AggregationStats.builder()
                .serviceDistribution(serviceDist)
                .levelDistribution(levelDist)
                .errorTrendLastHour(errorTrend)
                .build();
    }

    private LogResponse toResponse(LogEntry entry, String keyword) {
        String highlighted = null;
        if (keyword != null && !keyword.isEmpty()) {
            highlighted = highlightKeyword(entry.getMessage(), keyword);
        }

        return LogResponse.builder()
                .id(entry.getId())
                .service(entry.getService())
                .level(entry.getLevel())
                .message(entry.getMessage())
                .highlightedMessage(highlighted)
                .timestamp(entry.getTimestamp())
                .traceId(entry.getTraceId())
                .spanId(entry.getSpanId())
                .build();
    }

    private String highlightKeyword(String message, String keyword) {
        if (message == null || keyword == null || keyword.isEmpty()) {
            return message;
        }

        String lowerMessage = message.toLowerCase();
        String lowerKeyword = keyword.toLowerCase();

        if (!lowerMessage.contains(lowerKeyword)) {
            return message;
        }

        Pattern pattern = Pattern.compile(Pattern.quote(keyword), Pattern.CASE_INSENSITIVE);
        Matcher matcher = pattern.matcher(message);
        StringBuffer sb = new StringBuffer();

        while (matcher.find()) {
            String match = matcher.group();
            matcher.appendReplacement(sb, "[" + match + "]");
        }
        matcher.appendTail(sb);

        return sb.toString();
    }

    @Transactional
    public Map<LocalDate, Long> cleanupExpiredLogs() {
        LocalDateTime cutoff = LocalDateTime.now().minusDays(properties.getRetentionDays());
        Map<LocalDate, Long> deletionStats = new LinkedHashMap<>();

        for (int i = properties.getRetentionDays(); i < properties.getRetentionDays() + 30; i++) {
            LocalDate date = LocalDate.now().minusDays(i);
            LocalDateTime dayStart = date.atStartOfDay();
            LocalDateTime dayEnd = date.atTime(LocalTime.MAX);

            if (dayEnd.isBefore(cutoff)) {
                long count = logEntryRepository.countByTimestampBetween(dayStart, dayEnd);
                if (count > 0) {
                    log.info("Deleting {} logs for date: {}", count, date);
                    deletionStats.put(date, count);
                }
            }
        }

        long totalDeleted = logEntryRepository.deleteByTimestampBefore(cutoff);
        log.info("Cleanup completed. Total deleted: {} logs older than {}", totalDeleted, cutoff);

        return deletionStats;
    }
}
