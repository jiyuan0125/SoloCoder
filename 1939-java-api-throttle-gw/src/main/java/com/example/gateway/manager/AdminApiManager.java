package com.example.gateway.manager;

import com.example.gateway.model.RouteDefinition;
import com.example.gateway.ratelimit.RateLimitManager;
import com.example.gateway.ratelimit.TokenBucket;
import com.example.gateway.route.RouteManager;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpVersion;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.charset.StandardCharsets;
import java.util.Map;

public class AdminApiManager {
    private static final Logger logger = LoggerFactory.getLogger(AdminApiManager.class);
    private final ObjectMapper mapper = new ObjectMapper();

    private final RouteManager routeManager;
    private final RateLimitManager rateLimitManager;

    public AdminApiManager(RouteManager routeManager, RateLimitManager rateLimitManager) {
        this.routeManager = routeManager;
        this.rateLimitManager = rateLimitManager;
    }

    public boolean isAdminRequest(String uri, HttpMethod method) {
        return uri.startsWith("/_admin/");
    }

    public FullHttpResponse handleRequest(FullHttpRequest request) {
        String uri = request.uri();
        HttpMethod method = request.method();
        String body = request.content().toString(StandardCharsets.UTF_8);

        try {
            if (uri.equals("/_admin/routes")) {
                if (method.equals(HttpMethod.GET)) {
                    return listRoutes();
                } else if (method.equals(HttpMethod.POST)) {
                    return addRoute(body);
                }
            }

            if (uri.startsWith("/_admin/routes/")) {
                String path = uri.substring("/_admin/routes".length());
                if (method.equals(HttpMethod.DELETE)) {
                    return removeRoute(path);
                }
            }

            if (uri.equals("/_admin/ratelimit")) {
                if (method.equals(HttpMethod.GET)) {
                    return listRateLimits();
                } else if (method.equals(HttpMethod.POST)) {
                    return configureRateLimit(body);
                }
            }

            if (uri.equals("/_admin/health")) {
                return healthCheck();
            }

            return jsonError(HttpResponseStatus.NOT_FOUND, "Admin endpoint not found");
        } catch (Exception e) {
            logger.error("Admin API error", e);
            return jsonError(HttpResponseStatus.INTERNAL_SERVER_ERROR, e.getMessage());
        }
    }

    private FullHttpResponse listRoutes() throws Exception {
        ObjectNode result = mapper.createObjectNode();
        ArrayNode routesArray = result.putArray("routes");

        for (Map.Entry<String, RouteDefinition> entry : routeManager.getAllRoutes().entrySet()) {
            ObjectNode routeNode = routesArray.addObject();
            routeNode.put("path", entry.getKey());
            routeNode.put("targetUrl", entry.getValue().getTargetUrl());
        }

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse addRoute(String body) throws Exception {
        JsonNode json = mapper.readTree(body);
        String path = json.get("path").asText();
        String targetUrl = json.get("targetUrl").asText();

        routeManager.addRoute(path, targetUrl);

        ObjectNode result = mapper.createObjectNode();
        result.put("success", true);
        result.put("message", "Route added: " + path);

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse removeRoute(String path) throws Exception {
        routeManager.removeRoute(path);
        rateLimitManager.removeRoute(path);

        ObjectNode result = mapper.createObjectNode();
        result.put("success", true);
        result.put("message", "Route removed: " + path);

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse listRateLimits() throws Exception {
        ObjectNode result = mapper.createObjectNode();
        ArrayNode limitsArray = result.putArray("rateLimits");

        for (Map.Entry<String, TokenBucket> entry : rateLimitManager.getAllBuckets().entrySet()) {
            ObjectNode node = limitsArray.addObject();
            node.put("route", entry.getKey());
            node.put("capacity", entry.getValue().getCapacity());
            node.put("refillRatePerSecond", entry.getValue().getRefillRate());
            node.put("currentTokens", entry.getValue().getTokens());
        }

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse configureRateLimit(String body) throws Exception {
        JsonNode json = mapper.readTree(body);
        String route = json.get("route").asText();
        int capacity = json.get("capacity").asInt();
        int refillRate = json.get("refillRatePerSecond").asInt();

        rateLimitManager.configureRoute(route, capacity, refillRate);

        ObjectNode result = mapper.createObjectNode();
        result.put("success", true);
        result.put("message", "Rate limit configured for route: " + route);

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse healthCheck() throws Exception {
        ObjectNode result = mapper.createObjectNode();
        result.put("status", "UP");
        result.put("timestamp", System.currentTimeMillis());
        result.put("routes", routeManager.getAllRoutes().size());
        result.put("rateLimits", rateLimitManager.getAllBuckets().size());

        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private FullHttpResponse jsonResponse(HttpResponseStatus status, ObjectNode body) throws Exception {
        String jsonStr = mapper.writeValueAsString(body);
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1,
                status,
                Unpooled.copiedBuffer(jsonStr, CharsetUtil.UTF_8)
        );
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        return response;
    }

    private FullHttpResponse jsonError(HttpResponseStatus status, String message) {
        try {
            ObjectNode body = mapper.createObjectNode();
            body.put("error", status.code());
            body.put("message", message);
            return jsonResponse(status, body);
        } catch (Exception e) {
            return new DefaultFullHttpResponse(HttpVersion.HTTP_1_1, status);
        }
    }
}
