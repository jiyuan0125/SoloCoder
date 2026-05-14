package com.example.proxy.service;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.model.AccessLog;
import com.example.proxy.model.Backend;
import com.example.proxy.model.Route;
import lombok.extern.slf4j.Slf4j;
import org.apache.hc.client5.http.classic.methods.*;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.core5.http.ClassicHttpResponse;
import org.apache.hc.core5.http.Header;
import org.apache.hc.core5.http.HttpEntity;
import org.apache.hc.core5.http.HttpStatus;
import org.apache.hc.core5.http.io.entity.EntityUtils;
import org.apache.hc.core5.http.ContentType;
import org.apache.hc.core5.http.io.entity.InputStreamEntity;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.InputStream;
import java.net.SocketTimeoutException;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.Enumeration;
import java.util.List;
import java.util.UUID;

@Slf4j
@Service
public class ProxyService {
    
    private static final List<String> HOP_BY_HOP_HEADERS = Arrays.asList(
            "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
            "te", "trailers", "transfer-encoding", "upgrade"
    );
    
    private final ProxyProperties properties;
    private final RouteService routeService;
    private final LoadBalanceService loadBalanceService;
    private final HealthCheckService healthCheckService;
    private final AccessLogService accessLogService;
    private final CloseableHttpClient httpClient;
    
    public ProxyService(ProxyProperties properties, RouteService routeService,
                        LoadBalanceService loadBalanceService, HealthCheckService healthCheckService,
                        AccessLogService accessLogService, CloseableHttpClient httpClient) {
        this.properties = properties;
        this.routeService = routeService;
        this.loadBalanceService = loadBalanceService;
        this.healthCheckService = healthCheckService;
        this.accessLogService = accessLogService;
        this.httpClient = httpClient;
    }
    
    public void forwardRequest(HttpServletRequest request, HttpServletResponse response) throws IOException {
        long startTime = System.currentTimeMillis();
        String traceId = generateTraceId();
        String requestPath = request.getRequestURI();
        String method = request.getMethod();
        String sourceIp = getClientIp(request);
        
        AccessLog.AccessLogBuilder logBuilder = AccessLog.builder()
                .traceId(traceId)
                .requestTime(LocalDateTime.now())
                .sourceIp(sourceIp)
                .requestPath(requestPath)
                .method(method);
        
        try {
            Route route = routeService.matchRoute(requestPath);
            if (route == null) {
                log.warn("No route found for path: {}", requestPath);
                sendErrorResponse(response, HttpStatus.SC_NOT_FOUND, "Not Found");
                logBuilder.responseStatus(HttpStatus.SC_NOT_FOUND)
                        .errorMessage("No route found");
                return;
            }
            
            Backend backend = loadBalanceService.selectBackend(route);
            if (backend == null) {
                log.warn("No available backend for route: {}", route.getPath());
                sendErrorResponse(response, HttpStatus.SC_SERVICE_UNAVAILABLE, "Service Unavailable");
                logBuilder.responseStatus(HttpStatus.SC_SERVICE_UNAVAILABLE)
                        .errorMessage("No available backend");
                return;
            }
            
            logBuilder.targetBackend(backend.getUrl());
            log.info("[{}] Forwarding {} {} to {}", traceId, method, requestPath, backend.getUrl());
            
            String targetUrl = buildTargetUrl(backend.getUrl(), request);
            HttpUriRequestBase targetRequest = createTargetRequest(method, targetUrl, request);
            
            setForwardedHeaders(targetRequest, request, traceId);
            
            try (ClassicHttpResponse targetResponse = httpClient.execute(targetRequest)) {
                int statusCode = targetResponse.getCode();
                
                if (statusCode >= 500) {
                    healthCheckService.markBackendUnhealthy(backend, "5xx response: " + statusCode);
                } else {
                    healthCheckService.resetBackendStatus(backend);
                }
                
                copyResponseHeaders(targetResponse, response);
                response.setStatus(statusCode);
                
                HttpEntity entity = targetResponse.getEntity();
                if (entity != null) {
                    try (InputStream inputStream = entity.getContent()) {
                        byte[] buffer = new byte[4096];
                        int bytesRead;
                        while ((bytesRead = inputStream.read(buffer)) != -1) {
                            response.getOutputStream().write(buffer, 0, bytesRead);
                        }
                    }
                    EntityUtils.consume(entity);
                }
                
                logBuilder.responseStatus(statusCode);
                log.debug("[{}] Response status: {}", traceId, statusCode);
            }
            
        } catch (SocketTimeoutException e) {
            log.error("[{}] Request timeout: {}", traceId, e.getMessage());
            sendErrorResponse(response, HttpStatus.SC_GATEWAY_TIMEOUT, "Gateway Timeout");
            logBuilder.responseStatus(HttpStatus.SC_GATEWAY_TIMEOUT)
                    .errorMessage("Request timeout: " + e.getMessage());
            
            if (logBuilder.build().getTargetBackend() != null) {
                Backend backend = getBackendByUrl(logBuilder.build().getTargetBackend());
                if (backend != null) {
                    healthCheckService.markBackendTimeout(backend);
                }
            }
            
        } catch (Exception e) {
            log.error("[{}] Error forwarding request: {}", traceId, e.getMessage(), e);
            sendErrorResponse(response, HttpStatus.SC_BAD_GATEWAY, "Bad Gateway");
            logBuilder.responseStatus(HttpStatus.SC_BAD_GATEWAY)
                    .errorMessage("Error: " + e.getMessage());
        } finally {
            long duration = System.currentTimeMillis() - startTime;
            logBuilder.durationMs(duration);
            accessLogService.record(logBuilder.build());
            log.info("[{}] Request completed in {}ms with status {}", traceId, duration, 
                    logBuilder.build().getResponseStatus());
        }
    }
    
    private String generateTraceId() {
        return UUID.randomUUID().toString().replace("-", "");
    }
    
    private String getClientIp(HttpServletRequest request) {
        String xForwardedFor = request.getHeader("X-Forwarded-For");
        if (xForwardedFor != null && !xForwardedFor.isEmpty()) {
            return xForwardedFor.split(",")[0].trim();
        }
        
        String xRealIp = request.getHeader("X-Real-IP");
        if (xRealIp != null && !xRealIp.isEmpty()) {
            return xRealIp;
        }
        
        return request.getRemoteAddr();
    }
    
    private String buildTargetUrl(String backendUrl, HttpServletRequest request) {
        String requestPath = request.getRequestURI();
        String queryString = request.getQueryString();
        
        StringBuilder targetUrl = new StringBuilder(backendUrl);
        
        if (!backendUrl.endsWith("/") && !requestPath.startsWith("/")) {
            targetUrl.append("/");
        }
        targetUrl.append(requestPath);
        
        if (queryString != null && !queryString.isEmpty()) {
            targetUrl.append("?").append(queryString);
        }
        
        return targetUrl.toString();
    }
    
    private HttpUriRequestBase createTargetRequest(String method, String targetUrl, HttpServletRequest request) 
            throws IOException {
        HttpUriRequestBase targetRequest;
        
        switch (method.toUpperCase()) {
            case "GET":
                targetRequest = new HttpGet(targetUrl);
                break;
            case "POST":
                targetRequest = new HttpPost(targetUrl);
                break;
            case "PUT":
                targetRequest = new HttpPut(targetUrl);
                break;
            case "DELETE":
                targetRequest = new HttpDelete(targetUrl);
                break;
            case "PATCH":
                targetRequest = new HttpPatch(targetUrl);
                break;
            case "HEAD":
                targetRequest = new HttpHead(targetUrl);
                break;
            case "OPTIONS":
                targetRequest = new HttpOptions(targetUrl);
                break;
            default:
                throw new IllegalArgumentException("Unsupported method: " + method);
        }
        
        copyRequestHeaders(request, targetRequest);
        
        if (hasBody(method) && request.getInputStream() != null) {
            String contentType = request.getContentType();
            ContentType ct = contentType != null ? ContentType.parse(contentType) : ContentType.APPLICATION_OCTET_STREAM;
            
            if (targetRequest instanceof HttpPost) {
                ((HttpPost) targetRequest).setEntity(new InputStreamEntity(request.getInputStream(), ct));
            } else if (targetRequest instanceof HttpPut) {
                ((HttpPut) targetRequest).setEntity(new InputStreamEntity(request.getInputStream(), ct));
            } else if (targetRequest instanceof HttpPatch) {
                ((HttpPatch) targetRequest).setEntity(new InputStreamEntity(request.getInputStream(), ct));
            }
        }
        
        return targetRequest;
    }
    
    private boolean hasBody(String method) {
        return "POST".equalsIgnoreCase(method) || 
               "PUT".equalsIgnoreCase(method) || 
               "PATCH".equalsIgnoreCase(method);
    }
    
    private void copyRequestHeaders(HttpServletRequest request, HttpUriRequestBase targetRequest) {
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            String lowerHeaderName = headerName.toLowerCase();
            if (!HOP_BY_HOP_HEADERS.contains(lowerHeaderName) && 
                !"host".equalsIgnoreCase(headerName) &&
                !"content-length".equalsIgnoreCase(lowerHeaderName)) {
                String headerValue = request.getHeader(headerName);
                targetRequest.setHeader(headerName, headerValue);
            }
        }
    }
    
    private void setForwardedHeaders(HttpUriRequestBase targetRequest, HttpServletRequest request, String traceId) {
        String clientIp = getClientIp(request);
        String existingXff = request.getHeader("X-Forwarded-For");
        
        String xForwardedFor;
        if (existingXff != null && !existingXff.isEmpty()) {
            xForwardedFor = existingXff + ", " + clientIp;
        } else {
            xForwardedFor = clientIp;
        }
        
        targetRequest.setHeader("X-Forwarded-For", xForwardedFor);
        targetRequest.setHeader("X-Forwarded-Proto", request.getScheme());
        targetRequest.setHeader("X-Forwarded-Host", request.getServerName());
        targetRequest.setHeader("X-Forwarded-Port", String.valueOf(request.getServerPort()));
        targetRequest.setHeader("X-Trace-Id", traceId);
    }
    
    private void copyResponseHeaders(ClassicHttpResponse targetResponse, HttpServletResponse response) {
        Header[] headers = targetResponse.getHeaders();
        for (Header header : headers) {
            String headerName = header.getName();
            if (!HOP_BY_HOP_HEADERS.contains(headerName.toLowerCase())) {
                response.setHeader(headerName, header.getValue());
            }
        }
    }
    
    private void sendErrorResponse(HttpServletResponse response, int statusCode, String message) 
            throws IOException {
        response.setStatus(statusCode);
        response.setContentType("application/json");
        response.getWriter().write(String.format("{\"error\": \"%s\", \"status\": %d}", message, statusCode));
    }
    
    private Backend getBackendByUrl(String url) {
        for (Route route : routeService.getAllRoutes()) {
            for (Backend backend : route.getBackends()) {
                if (backend.getUrl().equals(url)) {
                    return backend;
                }
            }
        }
        return null;
    }
}
