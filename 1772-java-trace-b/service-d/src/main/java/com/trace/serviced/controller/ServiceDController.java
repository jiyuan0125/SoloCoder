package com.trace.serviced.controller;

import com.trace.serviced.config.ServiceEClient;
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
public class ServiceDController {

    private final ServiceEClient serviceEClient;
    private final Tracer tracer;

    @Observed(name = "service-d-request")
    @GetMapping("/api/d")
    public String handle() {
        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            span.tag("service.name", "service-d");
        }

        log.info("Service-D processing request");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        String serviceEResponse = serviceEClient.callServiceE();
        return "service-d → " + serviceEResponse;
    }

}
