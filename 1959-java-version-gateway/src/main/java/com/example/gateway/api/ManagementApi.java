package com.example.gateway.api;

import com.example.gateway.compatibility.CompatibilityRule;
import com.example.gateway.compatibility.FieldMapping;
import com.example.gateway.core.GatewayContext;
import com.example.gateway.routing.RouteRule;
import com.example.gateway.stats.StatsStore;
import com.example.gateway.stats.VersionStats;
import com.example.gateway.util.JsonUtils;
import com.fasterxml.jackson.databind.JsonNode;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponseStatus;

import java.util.*;

public class ManagementApi {
    private final GatewayContext context;

    public ManagementApi(GatewayContext context) {
        this.context = context;
    }

    public ApiResponse handle(String method, String path, String body) {
        HttpMethod httpMethod = HttpMethod.valueOf(method);

        if ("/stats/versions".equals(path) && httpMethod == HttpMethod.GET) {
            return handleGetStats();
        }

        if ("/routes".equals(path)) {
            if (httpMethod == HttpMethod.GET) {
                return handleGetRoutes();
            }
            if (httpMethod == HttpMethod.POST) {
                return handlePostRoute(body);
            }
        }

        if (path.startsWith("/routes/") && httpMethod == HttpMethod.DELETE) {
            String version = path.substring("/routes/".length());
            return handleDeleteRoute(version);
        }

        return null;
    }

    private ApiResponse handleGetStats() {
        StatsStore store = context.getStatsStore();
        Collection<VersionStats> stats = store.getAllStats();
        List<Map<String, Object>> result = new ArrayList<>();
        for (VersionStats vs : stats) {
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("version", vs.getVersion());
            item.put("totalCalls", vs.getTotalCalls());
            item.put("avgResponseTimeMs", vs.getAvgResponseTimeMs());
            item.put("clientErrorCount", vs.getClientErrorCount());
            item.put("serverErrorCount", vs.getServerErrorCount());
            item.put("clientErrorRate", vs.getClientErrorRate());
            item.put("serverErrorRate", vs.getServerErrorRate());
            item.put("errorRate", vs.getErrorRate());
            result.add(item);
        }
        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private ApiResponse handleGetRoutes() {
        Collection<RouteRule> routes = context.getRouteStore().getAllRoutes();
        List<Map<String, Object>> result = new ArrayList<>();
        for (RouteRule rule : routes) {
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("version", rule.getVersion());
            item.put("backends", rule.getBackends());
            item.put("compatibilityEnabled", rule.isCompatibilityEnabled());
            result.add(item);
        }
        return jsonResponse(HttpResponseStatus.OK, result);
    }

    private ApiResponse handlePostRoute(String body) {
        if (body == null || body.isEmpty()) {
            return errorResponse(HttpResponseStatus.BAD_REQUEST, "Empty request body");
        }
        try {
            JsonNode node = JsonUtils.getMapper().readTree(body);
            String version = node.has("version") ? node.get("version").asText() : null;
            List<String> backends = new ArrayList<>();
            if (node.has("backends") && node.get("backends").isArray()) {
                for (JsonNode b : node.get("backends")) {
                    backends.add(b.asText());
                }
            }
            boolean compatibilityEnabled = node.has("compatibilityEnabled")
                    ? node.get("compatibilityEnabled").asBoolean()
                    : false;

            if (version == null || version.isEmpty()) {
                return errorResponse(HttpResponseStatus.BAD_REQUEST, "version is required");
            }
            if (backends.isEmpty()) {
                return errorResponse(HttpResponseStatus.BAD_REQUEST, "backends is required and must be non-empty");
            }

            RouteRule rule = new RouteRule(version, backends, compatibilityEnabled);
            context.getRouteStore().addRoute(rule);

            if (compatibilityEnabled) {
                CompatibilityRule compatRule = parseCompatibilityRule(node, version);
                if (compatRule != null) {
                    context.getCompatibilityService().addRule(compatRule);
                }
            }

            Map<String, Object> response = new LinkedHashMap<>();
            response.put("status", "created");
            response.put("version", version);
            return jsonResponse(HttpResponseStatus.CREATED, response);

        } catch (Exception e) {
            return errorResponse(HttpResponseStatus.BAD_REQUEST, "Invalid JSON: " + e.getMessage());
        }
    }

    private CompatibilityRule parseCompatibilityRule(JsonNode node, String sourceVersion) {
        if (!node.has("compatibility")) {
            return null;
        }
        JsonNode compatNode = node.get("compatibility");
        CompatibilityRule rule = new CompatibilityRule();
        rule.setSourceVersion(sourceVersion);
        rule.setTargetVersion(compatNode.has("targetVersion")
                ? compatNode.get("targetVersion").asText()
                : "v2");
        rule.setRequestMappings(parseFieldMappings(compatNode, "requestMappings"));
        rule.setResponseMappings(parseFieldMappings(compatNode, "responseMappings"));
        return rule;
    }

    private List<FieldMapping> parseFieldMappings(JsonNode compatNode, String key) {
        List<FieldMapping> mappings = new ArrayList<>();
        if (compatNode.has(key) && compatNode.get(key).isArray()) {
            for (JsonNode item : compatNode.get(key)) {
                FieldMapping fm = new FieldMapping();
                if (item.has("from")) fm.setFrom(item.get("from").asText());
                if (item.has("to")) fm.setTo(item.get("to").asText());
                if (item.has("type")) fm.setType(item.get("type").asText());
                if (item.has("defaultValue")) {
                    JsonNode dv = item.get("defaultValue");
                    if (dv.isTextual()) fm.setDefaultValue(dv.asText());
                    else if (dv.isInt()) fm.setDefaultValue(dv.asInt());
                    else if (dv.isLong()) fm.setDefaultValue(dv.asLong());
                    else if (dv.isDouble()) fm.setDefaultValue(dv.asDouble());
                    else if (dv.isBoolean()) fm.setDefaultValue(dv.asBoolean());
                    else fm.setDefaultValue(dv.asText());
                }
                mappings.add(fm);
            }
        }
        return mappings;
    }

    private ApiResponse handleDeleteRoute(String version) {
        boolean removed = context.getRouteStore().removeRoute(version);
        context.getCompatibilityService().removeRule(version);
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("version", version);
        response.put("removed", removed);
        return jsonResponse(HttpResponseStatus.OK, response);
    }

    private ApiResponse jsonResponse(HttpResponseStatus status, Object body) {
        Map<String, String> headers = new HashMap<>();
        headers.put(HttpHeaderNames.CONTENT_TYPE.toString(), "application/json; charset=UTF-8");
        return new ApiResponse(status.code(), JsonUtils.toJson(body), headers);
    }

    private ApiResponse errorResponse(HttpResponseStatus status, String message) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("error", message);
        Map<String, String> headers = new HashMap<>();
        headers.put(HttpHeaderNames.CONTENT_TYPE.toString(), "application/json; charset=UTF-8");
        return new ApiResponse(status.code(), JsonUtils.toJson(body), headers);
    }

    public static class ApiResponse {
        private final int status;
        private final String body;
        private final Map<String, String> headers;

        public ApiResponse(int status, String body, Map<String, String> headers) {
            this.status = status;
            this.body = body;
            this.headers = headers;
        }

        public int getStatus() {
            return status;
        }

        public String getBody() {
            return body;
        }

        public Map<String, String> getHeaders() {
            return headers;
        }
    }
}
