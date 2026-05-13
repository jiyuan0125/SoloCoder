package com.healthchecker.service;

import com.healthchecker.entity.CheckResult;
import com.healthchecker.entity.ServiceConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.LocalDateTime;

@Service
public class HttpChecker {

    private static final Logger logger = LoggerFactory.getLogger(HttpChecker.class);
    private final HttpClient httpClient;

    public HttpChecker() {
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    public CheckResult check(ServiceConfig config) {
        String url = config.getCheckUrl();
        int timeoutSeconds = config.getTimeoutSeconds();
        LocalDateTime startTime = LocalDateTime.now();
        long startMs = System.currentTimeMillis();

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .timeout(Duration.ofSeconds(timeoutSeconds))
                .GET()
                .build();

        try {
            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            long durationMs = System.currentTimeMillis() - startMs;
            boolean success = response.statusCode() >= 200 && response.statusCode() < 300;

            logger.debug("服务 [{}] 检查完成: 状态码={}, 耗时={}ms, 成功={}",
                    config.getServiceName(), response.statusCode(), durationMs, success);

            return new CheckResult(startTime, durationMs, response.statusCode(), success);

        } catch (IOException | InterruptedException e) {
            long durationMs = System.currentTimeMillis() - startMs;
            logger.warn("服务 [{}] 检查失败: {}", config.getServiceName(), e.getMessage());

            if (e instanceof InterruptedException) {
                Thread.currentThread().interrupt();
            }

            return new CheckResult(startTime, durationMs, 0, false);
        }
    }
}