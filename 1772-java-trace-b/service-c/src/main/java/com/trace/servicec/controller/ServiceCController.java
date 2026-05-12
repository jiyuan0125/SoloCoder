package com.trace.servicec.controller;

import com.trace.servicec.config.ServiceDClient;
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
public class ServiceCController {

    private final ServiceDClient serviceDClient;
    private final Tracer tracer;

    @Observed(name = "service-c-request")
    @GetMapping("/api/c")
    public String handle() {
        Span span = tracer.currentSpan();
        String traceId = "N/A";
        String spanId = "N/A";
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            span.tag("service.name", "service-c");
        }

        log.info("Service-C processing request");
        log.info("TraceID: {}, SpanID: {}", traceId, spanId);

        String serviceDResponse = serviceDClient.callServiceD();
        return "service-c → " + serviceDResponse;
    }

}
