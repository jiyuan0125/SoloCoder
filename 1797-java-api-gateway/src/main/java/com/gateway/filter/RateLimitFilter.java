package com.gateway.filter;

import com.gateway.model.FilterResult;
import com.gateway.model.RequestContext;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
@Component
public class RateLimitFilter implements GatewayFilter {

    public static final String NAME = "ratelimit";

    private static final int MAX_REQUESTS_PER_SECOND = 100;
    private final Map<String, RateWindow> rateLimitMap = new ConcurrentHashMap<>();

    @Override
    public String getName() {
        return NAME;
    }

    @Override
    public FilterResult filter(RequestContext context) {
        String clientKey = getClientKey(context);
        long currentTime = System.currentTimeMillis();

        RateWindow window = rateLimitMap.compute(clientKey, (key, existing) -> {
            if (existing == null || currentTime - existing.startTime > 1000) {
                return new RateWindow(currentTime);
            }
            existing.increment();
            return existing;
        });

        if (window.count.get() > MAX_REQUESTS_PER_SECOND) {
            return FilterResult.reject("TOO_MANY_REQUESTS", "请求过于频繁", 429);
        }

        return FilterResult.pass();
    }

    private String getClientKey(RequestContext context) {
        String userId = context.getAttribute("userId");
        if (userId != null) {
            return userId;
        }
        return context.getRequest().getRemoteAddr();
    }

    private static class RateWindow {
        final long startTime;
        final AtomicInteger count = new AtomicInteger(1);

        RateWindow(long startTime) {
            this.startTime = startTime;
        }

        void increment() {
            count.incrementAndGet();
        }
    }
}
