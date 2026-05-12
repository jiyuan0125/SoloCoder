package com.trace.serviceb.filter;

import io.micrometer.tracing.Span;
import io.micrometer.tracing.Tracer;
import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

import java.io.IOException;

@Component
public class TraceResponseFilter implements Filter {

    private final Tracer tracer;

    public TraceResponseFilter(Tracer tracer) {
        this.tracer = tracer;
    }

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        Span span = tracer.currentSpan();
        String traceId = null;
        String spanId = null;
        if (span != null) {
            traceId = span.context().traceId();
            spanId = span.context().spanId();
            MDC.put("traceId", traceId);
            MDC.put("spanId", spanId);

            if (response instanceof HttpServletResponse httpResponse) {
                httpResponse.setHeader("X-Trace-Id", traceId);
                httpResponse.setHeader("X-Span-Id", spanId);
            }
        }

        try {
            chain.doFilter(request, response);
        } finally {
            if (traceId != null) {
                MDC.remove("traceId");
            }
            if (spanId != null) {
                MDC.remove("spanId");
            }
        }
    }

}
