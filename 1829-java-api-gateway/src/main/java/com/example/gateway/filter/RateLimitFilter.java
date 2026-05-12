package com.example.gateway.filter;

import com.example.gateway.model.RouteRule;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RateLimitFilter implements GatewayFilter {

    private final ObjectMapper objectMapper = new ObjectMapper();
    private final ConcurrentHashMap<String, WindowCounter> rateLimitMap = new ConcurrentHashMap<>();
    private static final int MAX_REQUESTS_PER_SECOND = 10;
    private static final int WINDOW_SIZE_MILLIS = 1000;

    @Override
    public String getName() {
        return "rateLimit";
    }

    @Override
    public FilterResult filter(HttpServletRequest request, HttpServletResponse response, RouteRule route) throws Exception {
        String key = getRateLimitKey(request);
        WindowCounter counter = rateLimitMap.computeIfAbsent(key, k -> new WindowCounter());
        
        synchronized (counter) {
            long now = System.currentTimeMillis();
            if (now - counter.windowStart > WINDOW_SIZE_MILLIS) {
                counter.windowStart = now;
                counter.count.set(0);
            }
            if (counter.count.incrementAndGet() > MAX_REQUESTS_PER_SECOND) {
                sendErrorResponse(response, 429, "TOO_MANY_REQUESTS", "限流触发：请求过于频繁");
                return FilterResult.REJECT;
            }
        }
        return FilterResult.PASS;
    }

    private String getRateLimitKey(HttpServletRequest request) {
        String clientIp = request.getRemoteAddr();
        String path = request.getRequestURI();
        return clientIp + ":" + path;
    }

    private void sendErrorResponse(HttpServletResponse response, int status, String code, String message) throws Exception {
        response.setStatus(status);
        response.setContentType("application/json;charset=UTF-8");
        Map<String, String> error = new HashMap<>();
        error.put("error", code);
        error.put("message", message);
        response.getWriter().write(objectMapper.writeValueAsString(error));
    }

    private static class WindowCounter {
        long windowStart = System.currentTimeMillis();
        AtomicInteger count = new AtomicInteger(0);
    }
}
