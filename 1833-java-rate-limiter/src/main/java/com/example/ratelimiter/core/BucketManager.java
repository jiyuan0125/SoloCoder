package com.example.ratelimiter.core;

import com.example.ratelimiter.model.BucketStats;
import com.example.ratelimiter.model.LimiterConfig;
import com.example.ratelimiter.model.QueuedRequest;
import jakarta.servlet.*;
import jakarta.servlet.http.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.DispatcherServlet;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.io.UnsupportedEncodingException;
import java.security.Principal;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicLong;

@Component
public class BucketManager {
    private static final Logger log = LoggerFactory.getLogger(BucketManager.class);

    private final Map<String, LeakyBucket> buckets = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> overflowCounts = new ConcurrentHashMap<>();
    private final DispatcherServlet dispatcherServlet;
    private final ExecutorService executorService;

    public BucketManager(DispatcherServlet dispatcherServlet) {
        this.dispatcherServlet = dispatcherServlet;
        this.executorService = Executors.newCachedThreadPool();
    }

    public Optional<LeakyBucket> getBucket(String path) {
        return Optional.ofNullable(buckets.get(path));
    }

    public boolean hasConfig(String path) {
        return buckets.containsKey(path);
    }

    public void updateConfig(LimiterConfig config) {
        LeakyBucket bucket = buckets.get(config.getPath());
        if (bucket == null) {
            bucket = new LeakyBucket(config.getPath(), config.getCapacity(), config.getRate());
            buckets.put(config.getPath(), bucket);
            overflowCounts.put(config.getPath(), new AtomicLong(0));
            log.info("Created new bucket for path: {}, capacity: {}, rate: {}", 
                    config.getPath(), config.getCapacity(), config.getRate());
        } else {
            bucket.setCapacity(config.getCapacity());
            bucket.setRate(config.getRate());
            log.info("Updated bucket for path: {}, capacity: {}, rate: {}", 
                    config.getPath(), config.getCapacity(), config.getRate());
        }
    }

    public void updateConfigs(List<LimiterConfig> configs) {
        for (LimiterConfig config : configs) {
            updateConfig(config);
        }
    }

    public boolean deleteConfig(String path) {
        LeakyBucket bucket = buckets.remove(path);
        overflowCounts.remove(path);
        if (bucket != null) {
            log.info("Deleted bucket for path: {}", path);
            while (bucket.getCurrentQueueSize() > 0) {
                QueuedRequest request = bucket.poll();
                if (request != null) {
                    request.markProcessed();
                }
            }
            return true;
        }
        return false;
    }

    public List<BucketStats> getAllStats() {
        List<BucketStats> stats = new ArrayList<>();
        for (Map.Entry<String, LeakyBucket> entry : buckets.entrySet()) {
            LeakyBucket bucket = entry.getValue();
            AtomicLong overflowCounter = overflowCounts.get(entry.getKey());
            stats.add(new BucketStats(
                    entry.getKey(),
                    bucket.getCurrentQueueSize(),
                    bucket.getCapacity(),
                    bucket.getRate(),
                    overflowCounter != null ? overflowCounter.get() : 0
            ));
        }
        stats.sort(Comparator.comparing(BucketStats::getPath));
        return stats;
    }

    public void recordOverflow(String path) {
        AtomicLong counter = overflowCounts.get(path);
        if (counter != null) {
            counter.incrementAndGet();
        }
    }

    @Scheduled(fixedDelay = 100)
    public void drainBuckets() {
        long nowNanos = System.nanoTime();
        for (Map.Entry<String, LeakyBucket> entry : buckets.entrySet()) {
            LeakyBucket bucket = entry.getValue();
            while (shouldDrain(bucket, nowNanos)) {
                QueuedRequest request = bucket.poll();
                if (request == null) {
                    break;
                }
                try {
                    processRequest(request);
                } catch (Exception e) {
                    log.error("Error processing request for path: {}", entry.getKey(), e);
                }
                bucket.updateLastDrainTime(nowNanos);
            }
        }
    }

    private boolean shouldDrain(LeakyBucket bucket, long nowNanos) {
        if (bucket.getCurrentQueueSize() == 0) {
            return false;
        }
        long elapsedNanos = nowNanos - bucket.getLastDrainTime();
        return elapsedNanos >= bucket.getIntervalNanos();
    }

    private void processRequest(QueuedRequest queuedRequest) {
        String path = queuedRequest.getRequest().getRequestURI();
        log.debug("Processing queued request for path: {}", path);
        
        executorService.submit(() -> {
            try {
                HttpServletRequestWrapper wrappedRequest = new HttpServletRequestWrapper(queuedRequest.getRequest());
                HttpServletResponse response = new HttpServletResponse() {
                    @Override public void addCookie(Cookie cookie) {}
                    @Override public boolean containsHeader(String name) { return false; }
                    @Override public String encodeURL(String url) { return url; }
                    @Override public String encodeRedirectURL(String url) { return url; }
                    @Override public void sendError(int sc, String msg) throws IOException {}
                    @Override public void sendError(int sc) throws IOException {}
                    @Override public void sendRedirect(String location) throws IOException {}
                    @Override public void setDateHeader(String name, long date) {}
                    @Override public void addDateHeader(String name, long date) {}
                    @Override public void setHeader(String name, String value) {}
                    @Override public void addHeader(String name, String value) {}
                    @Override public void setIntHeader(String name, int value) {}
                    @Override public void addIntHeader(String name, int value) {}
                    @Override public void setStatus(int sc) {}
                    @Override public int getStatus() { return 200; }
                    @Override public String getHeader(String name) { return null; }
                    @Override public Collection<String> getHeaders(String name) { return Collections.emptyList(); }
                    @Override public Collection<String> getHeaderNames() { return Collections.emptyList(); }
                    @Override public String getCharacterEncoding() { return "UTF-8"; }
                    @Override public String getContentType() { return "application/json"; }
                    @Override public ServletOutputStream getOutputStream() throws IOException { return null; }
                    @Override public PrintWriter getWriter() throws IOException { return null; }
                    @Override public void setCharacterEncoding(String charset) {}
                    @Override public void setContentLength(int len) {}
                    @Override public void setContentLengthLong(long len) {}
                    @Override public void setContentType(String type) {}
                    @Override public void setBufferSize(int size) {}
                    @Override public int getBufferSize() { return 0; }
                    @Override public void flushBuffer() throws IOException {}
                    @Override public void resetBuffer() {}
                    @Override public boolean isCommitted() { return false; }
                    @Override public void reset() {}
                    @Override public void setLocale(Locale loc) {}
                    @Override public Locale getLocale() { return Locale.getDefault(); }
                };
                dispatcherServlet.service(wrappedRequest, response);
                log.debug("Completed processing request for path: {}", path);
            } catch (Exception e) {
                log.error("Error processing queued request for path: {}", path, e);
            } finally {
                queuedRequest.markProcessed();
            }
        });
    }

    public int calculateRetryAfterSeconds(String path) {
        LeakyBucket bucket = buckets.get(path);
        if (bucket == null) {
            return 1;
        }
        double secondsPerRequest = 1.0 / bucket.getRate();
        double approxWaitSeconds = bucket.getCurrentQueueSize() * secondsPerRequest;
        return Math.max(1, (int) Math.ceil(approxWaitSeconds));
    }
}
