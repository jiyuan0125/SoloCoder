package com.trace.servicea.controller;

import com.trace.servicea.config.ServiceBClient;
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
public class ServiceAController {

    private final ServiceBClient serviceBClient;
    private final Tracer tracer;

    @Observed(name = "service-a-request")
    @GetMapping("/api/a")
    public String handle() {
        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            span.tag("service.name", "service-a");
        }

        log.info("Service-A processing request");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        String serviceBResponse = serviceBClient.callServiceB();
        return "service-a → " + serviceBResponse;
    }

}
