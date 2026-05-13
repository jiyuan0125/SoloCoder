package com.healthchecker.service;

import com.healthchecker.entity.StatusChangeNotification;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

@Service
public class NotificationService {

    private static final Logger logger = LoggerFactory.getLogger(NotificationService.class);
    private static final int MAX_RETRY = 1;
    private static final int TIMEOUT_SECONDS = 10;

    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public NotificationService() {
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(TIMEOUT_SECONDS))
                .build();
        this.objectMapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    public void sendNotification(String callbackUrl, StatusChangeNotification notification) {
        if (callbackUrl == null || callbackUrl.isBlank()) {
            logger.warn("服务 [{}] 未配置回调URL，跳过通知", notification.serviceName());
            return;
        }

        try {
            String jsonBody = objectMapper.writeValueAsString(notification);
            logger.info("发送状态变更通知: 服务={}, 旧状态={}, 新状态={}",
                    notification.serviceName(), notification.oldStatus(), notification.newStatus());

            boolean sent = sendWithRetry(callbackUrl, jsonBody);
            if (!sent) {
                logger.error("通知发送失败，已重试{}次: 服务={}", MAX_RETRY, notification.serviceName());
            } else {
                logger.info("通知发送成功: 服务={}", notification.serviceName());
            }
        } catch (Exception e) {
            logger.error("序列化通知失败: 服务={}", notification.serviceName(), e);
        }
    }

    private boolean sendWithRetry(String callbackUrl, String jsonBody) {
        int attempts = 0;
        while (attempts <= MAX_RETRY) {
            attempts++;
            try {
                HttpRequest request = HttpRequest.newBuilder()
                        .uri(URI.create(callbackUrl))
                        .timeout(Duration.ofSeconds(TIMEOUT_SECONDS))
                        .header("Content-Type", "application/json")
                        .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
                        .build();

                HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
                int statusCode = response.statusCode();

                if (statusCode >= 200 && statusCode < 300) {
                    return true;
                }

                logger.warn("通知响应状态码异常: {} (第{}次)", statusCode, attempts);

            } catch (IOException | InterruptedException e) {
                logger.warn("通知发送失败 (第{}次): {}", attempts, e.getMessage());
                if (e instanceof InterruptedException) {
                    Thread.currentThread().interrupt();
                    return false;
                }
            }

            if (attempts <= MAX_RETRY) {
                try {
                    Thread.sleep(500);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    return false;
                }
            }
        }
        return false;
    }
}