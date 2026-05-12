package com.canary.router.netty;

import com.canary.router.core.RouterManager;
import com.canary.router.model.Stats;
import com.canary.router.model.Version;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.http.*;

import java.nio.charset.StandardCharsets;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class ApiHandler {
    private static final ObjectMapper mapper = new ObjectMapper();
    private static final Pattern DELETE_VERSION_PATTERN = Pattern.compile("^/versions/([^/]+)$");
    
    private final RouterManager routerManager = RouterManager.getInstance();
    
    public FullHttpResponse handle(HttpRequest request) {
        String uri = request.uri();
        HttpMethod method = request.method();
        
        if (uri.equals("/versions") && method.equals(HttpMethod.POST)) {
            return handleCreateVersion(request);
        } else if (uri.startsWith("/versions/") && method.equals(HttpMethod.DELETE)) {
            return handleDeleteVersion(uri);
        } else if (uri.equals("/versions") && method.equals(HttpMethod.GET)) {
            return handleListVersions();
        } else if (uri.equals("/stats") && method.equals(HttpMethod.GET)) {
            return handleGetStats();
        } else if (uri.equals("/health") && method.equals(HttpMethod.GET)) {
            return jsonResponse(HttpResponseStatus.OK, Map.of("status", "ok"));
        }
        
        return null;
    }
    
    private FullHttpResponse handleCreateVersion(HttpRequest request) {
        try {
            if (!(request instanceof FullHttpRequest)) {
                return jsonResponse(HttpResponseStatus.BAD_REQUEST, 
                    Map.of("error", "Missing request body"));
            }
            
            FullHttpRequest fullRequest = (FullHttpRequest) request;
            String body = fullRequest.content().toString(StandardCharsets.UTF_8);
            
            Version version = mapper.readValue(body, Version.class);
            
            if (version.getName() == null || version.getName().isEmpty()) {
                return jsonResponse(HttpResponseStatus.BAD_REQUEST, 
                    Map.of("error", "Version name is required"));
            }
            
            routerManager.addVersion(version);
            
            return jsonResponse(HttpResponseStatus.CREATED, version);
        } catch (Exception e) {
            return jsonResponse(HttpResponseStatus.BAD_REQUEST, 
                Map.of("error", "Invalid request: " + e.getMessage()));
        }
    }
    
    private FullHttpResponse handleDeleteVersion(String uri) {
        Matcher matcher = DELETE_VERSION_PATTERN.matcher(uri);
        if (!matcher.matches()) {
            return jsonResponse(HttpResponseStatus.BAD_REQUEST, 
                Map.of("error", "Invalid URL format"));
        }
        
        String versionName = matcher.group(1);
        
        if (routerManager.getVersion(versionName) == null) {
            return jsonResponse(HttpResponseStatus.NOT_FOUND, 
                Map.of("error", "Version not found"));
        }
        
        boolean removed = routerManager.removeVersion(versionName);
        
        if (!removed) {
            return jsonResponse(HttpResponseStatus.CONFLICT, 
                Map.of("error", "Version is referenced by routing rules"));
        }
        
        return jsonResponse(HttpResponseStatus.NO_CONTENT, null);
    }
    
    private FullHttpResponse handleListVersions() {
        Collection<Version> versions = routerManager.getActiveVersions();
        return jsonResponse(HttpResponseStatus.OK, versions);
    }
    
    private FullHttpResponse handleGetStats() {
        Collection<Stats> stats = routerManager.getAllStats();
        Map<String, Object> result = new HashMap<>();
        for (Stats s : stats) {
            Map<String, Long> versionStats = new HashMap<>();
            versionStats.put("requests", s.getRequestCount());
            versionStats.put("errors", s.getErrorCount());
            result.put(s.getVersionName(), versionStats);
        }
        return jsonResponse(HttpResponseStatus.OK, result);
    }
    
    static FullHttpResponse jsonResponse(HttpResponseStatus status, Object body) {
        FullHttpResponse response;
        try {
            if (body == null) {
                response = new DefaultFullHttpResponse(HttpVersion.HTTP_1_1, status);
            } else {
                String json = mapper.writeValueAsString(body);
                response = new DefaultFullHttpResponse(
                    HttpVersion.HTTP_1_1, status,
                    Unpooled.wrappedBuffer(json.getBytes(StandardCharsets.UTF_8)));
                response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json; charset=UTF-8");
            }
            
            if (response.content() != null && response.content().readableBytes() > 0) {
                response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
            } else {
                response.headers().set(HttpHeaderNames.CONTENT_LENGTH, 0);
            }
            
            response.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
            return response;
        } catch (Exception e) {
            return errorResponse(HttpResponseStatus.INTERNAL_SERVER_ERROR, 
                "Error building response: " + e.getMessage());
        }
    }
    
    static FullHttpResponse errorResponse(HttpResponseStatus status, String message) {
        FullHttpResponse response = new DefaultFullHttpResponse(
            HttpVersion.HTTP_1_1, status,
            Unpooled.wrappedBuffer(("{\"error\":\"" + message + "\"}").getBytes(StandardCharsets.UTF_8)));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        response.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
        return response;
    }
}
