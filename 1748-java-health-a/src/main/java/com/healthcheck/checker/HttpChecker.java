package com.healthcheck.checker;

import com.healthcheck.model.CheckItemConfig;
import com.healthcheck.model.CheckResult;
import com.healthcheck.model.CheckStatus;
import com.healthcheck.model.CheckType;
import org.springframework.stereotype.Component;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.Instant;

@Component
public class HttpChecker implements Checker {

    private final HttpClient httpClient;

    public HttpChecker() {
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    @Override
    public CheckResult execute(CheckItemConfig config, String serviceId) {
        long startTime = System.currentTimeMillis();
        Instant now = Instant.now();
        
        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(config.getTarget()))
                    .timeout(Duration.ofMillis(config.getTimeout()))
                    .method(config.getMethod() == null ? "GET" : config.getMethod(), 
                            HttpRequest.BodyPublishers.noBody())
                    .build();

            HttpResponse<String> response = httpClient.send(request, 
                    HttpResponse.BodyHandlers.ofString());
            
            long responseTime = System.currentTimeMillis() - startTime;
            int statusCode = response.statusCode();
            int expectedCode = config.getExpectedStatusCode() != null ? config.getExpectedStatusCode() : 200;
            
            boolean success = statusCode == expectedCode;
            String message = "HTTP " + statusCode;
            
            if (config.getExpectedContent() != null && !config.getExpectedContent().isEmpty()) {
                String body = response.body();
                if (body != null && body.contains(config.getExpectedContent())) {
                    success = success && true;
                    message += ", content matched";
                } else {
                    success = false;
                    message += ", content not matched";
                }
            }

            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(success ? CheckStatus.NORMAL : CheckStatus.ERROR)
                    .message(message)
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(success)
                    .build();

        } catch (Exception e) {
            long responseTime = System.currentTimeMillis() - startTime;
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.ERROR)
                    .message("HTTP check failed: " + e.getMessage())
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(false)
                    .build();
        }
    }

    @Override
    public boolean supports(CheckItemConfig config) {
        return CheckType.HTTP.equals(config.getType());
    }
}
