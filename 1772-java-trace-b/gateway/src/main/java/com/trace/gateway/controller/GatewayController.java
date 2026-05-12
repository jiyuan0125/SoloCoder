package com.trace.gateway.controller;

import com.trace.gateway.config.ServiceAClient;
import io.micrometer.observation.annotation.Observed;
import io.micrometer.tracing.Span;
import io.micrometer.tracing.Tracer;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@Slf4j
@RestController
@RequiredArgsConstructor
public class GatewayController {

    private final ServiceAClient serviceAClient;
    private final Tracer tracer;

    @Observed(name = "gateway-request", contextualName = "gateway-entry-point")
    @GetMapping("/api/start")
    public Map<String, Object> start(
            @RequestHeader(value = "X-User-Id", required = false) String userId,
            @RequestHeader(value = "X-Request-Id", required = false) String requestId) {

        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            if (userId != null) {
                span.tag("user.id", userId);
            }
            if (requestId != null) {
                span.tag("request.id", requestId);
            }
            span.tag("service.name", "gateway");
        }

        log.info("Gateway received request, starting trace chain");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        String serviceAResponse = serviceAClient.callServiceA();

        return Map.of(
                "service", "gateway",
                "traceId", traceId,
                "chain", "gateway → " + serviceAResponse,
                "timestamp", System.currentTimeMillis()
        );
    }

}
