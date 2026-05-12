package com.trace.servicee.controller;

import io.micrometer.observation.annotation.Observed;
import io.micrometer.tracing.Span;
import io.micrometer.tracing.Tracer;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@Slf4j
@RestController
@RequiredArgsConstructor
public class ServiceEController {

    private final Tracer tracer;

    @Observed(name = "service-e-request")
    @GetMapping("/api/e")
    public String handle() {
        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            span.tag("service.name", "service-e");
            span.tag("chain.end", "true");
        }

        log.info("Service-E processing request - END OF CHAIN");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        return "service-e (end)";
    }

}
