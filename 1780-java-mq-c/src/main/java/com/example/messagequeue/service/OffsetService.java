package com.example.messagequeue.service;

import com.example.messagequeue.config.MessageQueueProperties;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Slf4j
@Service
@RequiredArgsConstructor
public class OffsetService {

    private static final Pattern OFFSET_PATTERN = Pattern.compile("^(\\d+)\\|(.+)\\|(.+)\\|(-?\\d+)$");
    private static final String DEFAULT_CONSUMER_ID = "default";

    private final MessageQueueProperties properties;

    private Path offsetDir;
    private final Map<String, Long> cachedOffsets = new ConcurrentHashMap<>();
    private final Map<String, ReadWriteLock> locks = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() throws IOException {
        offsetDir = Paths.get(properties.getOffsetDir());
        if (!Files.exists(offsetDir)) {
            Files.createDirectories(offsetDir);
            log.info("Created offset directory: {}", offsetDir.toAbsolutePath());
        }
        loadAllOffsets();
    }

    @PreDestroy
    public void shutdown() {
        log.info("OffsetService shutting down");
    }

    @Scheduled(fixedRateString = "${message-queue.offset-compaction-interval-minutes:60}", timeUnit = java.util.concurrent.TimeUnit.MINUTES)
    public void compactOffsets() {
        log.info("Starting offset compaction");
        for (String key : cachedOffsets.keySet()) {
            String[] parts = key.split(":", 2);
            if (parts.length == 2) {
                compactOffsetFile(parts[0], parts[1]);
            }
        }
        log.info("Offset compaction completed");
    }

    public void persistOffset(String topic, String consumerId, long offset) {
        if (consumerId == null || consumerId.isEmpty()) {
            consumerId = DEFAULT_CONSUMER_ID;
        }
        String key = topic + ":" + consumerId;
        ReadWriteLock lock = getLock(key);
        lock.writeLock().lock();
        try {
            cachedOffsets.put(key, offset);
            appendOffsetToFile(topic, consumerId, offset);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public long getLastCommittedOffset(String topic, String consumerId) {
        if (consumerId == null || consumerId.isEmpty()) {
            consumerId = DEFAULT_CONSUMER_ID;
        }
        String key = topic + ":" + consumerId;
        ReadWriteLock lock = getLock(key);
        lock.readLock().lock();
        try {
            Long offset = cachedOffsets.get(key);
            return offset != null ? offset : -1;
        } finally {
            lock.readLock().unlock();
        }
    }

    private void loadAllOffsets() throws IOException {
        try (var stream = Files.list(offsetDir)) {
            var files = stream.filter(Files::isRegularFile).toList();
            for (Path file : files) {
                String fileName = file.getFileName().toString();
                if (!fileName.endsWith(".offset")) {
                    continue;
                }
                String baseName = fileName.substring(0, fileName.length() - 7);
                String[] parts = baseName.split("__", 2);
                String topic = parts[0];
                String consumerId = parts.length > 1 ? parts[1] : DEFAULT_CONSUMER_ID;
                long lastOffset = loadLastOffsetFromFile(file);
                if (lastOffset >= 0) {
                    cachedOffsets.put(topic + ":" + consumerId, lastOffset);
                    log.info("Loaded offset for topic {}, consumer {}: {}", topic, consumerId, lastOffset);
                }
            }
        }
    }

    private long loadLastOffsetFromFile(Path file) throws IOException {
        if (!Files.exists(file)) {
            return -1;
        }
        List<String> lines = Files.readAllLines(file, StandardCharsets.UTF_8);
        long lastOffset = -1;
        for (String line : lines) {
            if (line.isBlank()) {
                continue;
            }
            Matcher matcher = OFFSET_PATTERN.matcher(line);
            if (matcher.matches()) {
                String offsetStr = matcher.group(4);
                try {
                    long offset = Long.parseLong(offsetStr);
                    if (offset > lastOffset) {
                        lastOffset = offset;
                    }
                } catch (NumberFormatException e) {
                    log.warn("Invalid offset in file {}: {}", file, line);
                }
            }
        }
        return lastOffset;
    }

    private void appendOffsetToFile(String topic, String consumerId, long offset) {
        Path file = getOffsetFile(topic, consumerId);
        String line = String.format("%d|%s|%s|%d%n", System.currentTimeMillis(), topic, consumerId, offset);
        try {
            Files.writeString(file, line, StandardCharsets.UTF_8,
                    StandardOpenOption.CREATE, StandardOpenOption.APPEND, StandardOpenOption.SYNC);
        } catch (IOException e) {
            log.error("Failed to append offset to file {}: {}", file, e.getMessage(), e);
            throw new RuntimeException("Failed to persist offset", e);
        }
    }

    private void compactOffsetFile(String topic, String consumerId) {
        String key = topic + ":" + consumerId;
        ReadWriteLock lock = getLock(key);
        lock.writeLock().lock();
        try {
            Long currentOffset = cachedOffsets.get(key);
            if (currentOffset == null) {
                return;
            }
            Path file = getOffsetFile(topic, consumerId);
            if (!Files.exists(file)) {
                return;
            }
            String compactedLine = String.format("%d|%s|%s|%d%n",
                    System.currentTimeMillis(), topic, consumerId, currentOffset);
            Files.writeString(file, compactedLine, StandardCharsets.UTF_8,
                    StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING, StandardOpenOption.SYNC);
            log.info("Compacted offset file for topic {}, consumer {}: offset={}", topic, consumerId, currentOffset);
        } catch (IOException e) {
            log.error("Failed to compact offset file for topic {}, consumer {}: {}",
                    topic, consumerId, e.getMessage(), e);
        } finally {
            lock.writeLock().unlock();
        }
    }

    private Path getOffsetFile(String topic, String consumerId) {
        String fileName = topic + "__" + (consumerId != null ? consumerId : DEFAULT_CONSUMER_ID) + ".offset";
        return offsetDir.resolve(fileName);
    }

    private ReadWriteLock getLock(String key) {
        return locks.computeIfAbsent(key, k -> new ReentrantReadWriteLock());
    }
}
