package com.trace.serviceb.controller;

import com.trace.serviceb.config.ServiceCClient;
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
public class ServiceBController {

    private final ServiceCClient serviceCClient;
    private final Tracer tracer;

    @Observed(name = "service-b-request")
    @GetMapping("/api/b")
    public String handle() {
        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            span.tag("service.name", "service-b");
        }

        log.info("Service-B processing request");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        String serviceCResponse = serviceCClient.callServiceC();
        return "service-b → " + serviceCResponse;
    }

}
