package com.servicemesh.sidecar.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.servicemesh.sidecar.config.SidecarProperties;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class TelemetryService {

    private static final Logger log = LoggerFactory.getLogger(TelemetryService.class);
    private static final MediaType JSON = MediaType.parse("application/json; charset=utf-8");

    private final SidecarProperties properties;
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;

    private final Map<String, AtomicLong> callCounts = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> errorCounts = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> totalLatency = new ConcurrentHashMap<>();

    public TelemetryService(SidecarProperties properties) {
        this.properties = properties;
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(3, TimeUnit.SECONDS)
                .readTimeout(3, TimeUnit.SECONDS)
                .build();
        this.objectMapper = new ObjectMapper();
    }

    public void recordCall(String source, String target, boolean success, long latencyMs) {
        String key = source + "->" + target;
        callCounts.computeIfAbsent(key, k -> new AtomicLong(0)).incrementAndGet();
        totalLatency.computeIfAbsent(key, k -> new AtomicLong(0)).addAndGet(latencyMs);
        if (!success) {
            errorCounts.computeIfAbsent(key, k -> new AtomicLong(0)).incrementAndGet();
        }
    }

    @Scheduled(fixedRate = 10000)
    public void reportTelemetry() {
        for (Map.Entry<String, AtomicLong> entry : callCounts.entrySet()) {
            String key = entry.getKey();
            long calls = entry.getValue().getAndSet(0);
            if (calls == 0) continue;

            long errors = errorCounts.getOrDefault(key, new AtomicLong(0)).getAndSet(0);
            long totalLat = totalLatency.getOrDefault(key, new AtomicLong(0)).getAndSet(0);
            double avgLatency = calls > 0 ? (double) totalLat / calls : 0.0;

            String[] parts = key.split("->", 2);
            if (parts.length != 2) continue;

            try {
                Map<String, Object> payload = Map.of(
                        "source", parts[0],
                        "target", parts[1],
                        "qps", (double) calls / 10.0,
                        "totalRequests", calls,
                        "errorCount", errors,
                        "errorRate", calls > 0 ? (double) errors / calls * 100 : 0.0,
                        "avgLatencyMs", avgLatency
                );

                RequestBody body = RequestBody.create(objectMapper.writeValueAsString(payload), JSON);
                Request request = new Request.Builder()
                        .url(properties.getControlPlaneUrl() + "/api/topology/telemetry")
                        .post(body)
                        .build();

                httpClient.newCall(request).enqueue(new okhttp3.Callback() {
                    @Override public void onFailure(okhttp3.Call call, java.io.IOException e) {
                        log.debug("Telemetry report failed: {}", e.getMessage());
                    }
                    @Override public void onResponse(okhttp3.Call call, okhttp3.Response response) {
                        response.close();
                    }
                });
            } catch (Exception e) {
                log.debug("Telemetry serialization failed", e);
            }
        }
    }
}
