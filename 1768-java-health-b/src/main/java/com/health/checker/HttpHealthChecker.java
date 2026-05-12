package com.health.checker;

import com.health.model.CheckResult;
import com.health.model.CheckType;
import com.health.model.ServiceRegistration;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.LocalDateTime;

@Component
public class HttpHealthChecker implements HealthChecker {

    private final HttpClient httpClient;

    public HttpHealthChecker() {
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(5))
                .build();
    }

    @Override
    public CheckResult check(ServiceRegistration service) {
        LocalDateTime start = LocalDateTime.now();
        long startTime = System.currentTimeMillis();

        String url = buildUrl(service);
        String method = service.getMethod() != null ? service.getMethod() : "GET";
        int timeoutMs = service.getTimeoutMilliseconds() != null ? 
                service.getTimeoutMilliseconds() : 5000;

        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .method(method, HttpRequest.BodyPublishers.noBody())
                    .timeout(Duration.ofMillis(timeoutMs))
                    .build();

            HttpResponse<Void> response = httpClient.send(request, 
                    HttpResponse.BodyHandlers.discarding());

            int responseTime = (int) (System.currentTimeMillis() - startTime);
            int statusCode = response.statusCode();
            
            boolean healthy = statusCode >= 200 && statusCode < 400;
            
            return CheckResult.builder()
                    .healthy(healthy)
                    .responseTimeMs(responseTime)
                    .message(healthy ? "OK - HTTP " + statusCode : "HTTP Error: " + statusCode)
                    .checkedAt(start)
                    .build();

        } catch (IOException e) {
            int responseTime = (int) (System.currentTimeMillis() - startTime);
            return CheckResult.builder()
                    .healthy(false)
                    .responseTimeMs(responseTime)
                    .message("Connection failed: " + e.getMessage())
                    .checkedAt(start)
                    .build();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            int responseTime = (int) (System.currentTimeMillis() - startTime);
            return CheckResult.builder()
                    .healthy(false)
                    .responseTimeMs(responseTime)
                    .message("Request interrupted")
                    .checkedAt(start)
                    .build();
        }
    }

    private String buildUrl(ServiceRegistration service) {
        String host = service.getEndpoint();
        Integer port = service.getPort();
        String path = service.getPath() != null ? service.getPath() : "";

        if (host.startsWith("http://") || host.startsWith("https://")) {
            return host + (path.startsWith("/") ? path : "/" + path);
        }

        String scheme = (port != null && (port == 443 || port == 8443)) ? "https" : "http";
        StringBuilder url = new StringBuilder(scheme + "://" + host);
        
        if (port != null) {
            url.append(":").append(port);
        }
        
        if (path != null && !path.isEmpty()) {
            if (!path.startsWith("/")) {
                url.append("/");
            }
            url.append(path);
        }

        return url.toString();
    }

    @Override
    public boolean supports(ServiceRegistration service) {
        return CheckType.HTTP.equals(service.getCheckType());
    }
}