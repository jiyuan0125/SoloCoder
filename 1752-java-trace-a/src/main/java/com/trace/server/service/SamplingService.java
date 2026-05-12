package com.trace.server.service;

import com.trace.server.config.TraceConfig;
import com.trace.server.model.Span;
import org.springframework.stereotype.Service;

import java.util.Random;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class SamplingService {

    private final TraceConfig traceConfig;
    private final Random random = new Random();
    private final Set<String> sampledTraces = ConcurrentHashMap.newKeySet();

    public SamplingService(TraceConfig traceConfig) {
        this.traceConfig = traceConfig;
    }

    public boolean shouldSample(Span span) {
        String traceId = span.getTraceId();
        if (traceId == null || traceId.isEmpty()) {
            return false;
        }

        if (Boolean.TRUE.equals(span.getIsError())) {
            sampledTraces.add(traceId);
            return true;
        }

        if (sampledTraces.contains(traceId)) {
            return true;
        }

        double rate = traceConfig.getSampleRate();
        if (rate >= 1.0) {
            sampledTraces.add(traceId);
            return true;
        }
        if (rate <= 0.0) {
            return false;
        }

        if (random.nextDouble() < rate) {
            sampledTraces.add(traceId);
            return true;
        }

        return false;
    }
}
