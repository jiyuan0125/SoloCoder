package com.example.ratelimiter.filter;

import com.example.ratelimiter.core.BucketManager;
import com.example.ratelimiter.core.LeakyBucket;
import com.example.ratelimiter.model.QueuedRequest;
import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.Optional;

@Component
@Order(1)
public class RateLimiterFilter implements Filter {
    private static final Logger log = LoggerFactory.getLogger(RateLimiterFilter.class);

    private final BucketManager bucketManager;

    public RateLimiterFilter(BucketManager bucketManager) {
        this.bucketManager = bucketManager;
    }

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        HttpServletRequest httpRequest = (HttpServletRequest) request;
        HttpServletResponse httpResponse = (HttpServletResponse) response;
        
        String path = httpRequest.getRequestURI();

        if (shouldSkipRateLimiting(path)) {
            chain.doFilter(request, response);
            return;
        }

        if (!bucketManager.hasConfig(path)) {
            log.debug("No rate limit config for path: {}, passing through", path);
            chain.doFilter(request, response);
            return;
        }

        Optional<LeakyBucket> bucketOpt = bucketManager.getBucket(path);
        if (bucketOpt.isEmpty()) {
            chain.doFilter(request, response);
            return;
        }

        LeakyBucket bucket = bucketOpt.get();
        int capacity = bucket.getCapacity();
        double rate = bucket.getRate();

        QueuedRequest queuedRequest = new QueuedRequest(httpRequest, capacity, rate);
        boolean enqueued = bucket.tryEnqueue(queuedRequest);

        if (!enqueued) {
            log.warn("Bucket full for path: {}, rejecting request", path);
            bucketManager.recordOverflow(path);
            int retryAfter = bucketManager.calculateRetryAfterSeconds(path);
            httpResponse.setStatus(429);
            httpResponse.setHeader("Retry-After", String.valueOf(retryAfter));
            httpResponse.setContentType("application/json");
            httpResponse.getWriter().write("{\"message\":\"Rate limit exceeded\",\"retryAfter\":" + retryAfter + "}");
            return;
        }

        log.debug("Request enqueued for path: {}, returning 202 Accepted", path);
        httpResponse.setStatus(HttpServletResponse.SC_ACCEPTED);
        httpResponse.setContentType("application/json");
        httpResponse.getWriter().write("{\"message\":\"Request accepted and queued\",\"path\":\"" + path + "\"}");
    }

    private boolean shouldSkipRateLimiting(String path) {
        return path.startsWith("/limiter/");
    }
}
