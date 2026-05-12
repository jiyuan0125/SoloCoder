package com.gateway.handler;

import com.gateway.engine.RuleMatchingEngine;
import com.gateway.engine.TrafficDistributionEngine;
import com.gateway.manager.RuleManager;
import com.gateway.manager.StatsManager;
import com.gateway.manager.VersionManager;
import com.gateway.model.Rule;
import com.gateway.model.Stats;
import com.gateway.model.Version;
import com.gateway.util.JsonUtil;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import io.netty.util.CharsetUtil;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

public class ApiHandler extends SimpleChannelInboundHandler<FullHttpRequest> {

    private final VersionManager versionManager;
    private final RuleManager ruleManager;
    private final StatsManager statsManager;

    public ApiHandler(VersionManager versionManager, RuleManager ruleManager, StatsManager statsManager) {
        this.versionManager = versionManager;
        this.ruleManager = ruleManager;
        this.statsManager = statsManager;
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) throws Exception {
        String uri = request.uri();
        HttpMethod method = request.method();

        if (uri.equals("/versions")) {
            handleVersions(ctx, request, method);
        } else if (uri.startsWith("/versions/")) {
            handleVersionById(ctx, request, method, uri);
        } else if (uri.equals("/rules")) {
            handleRules(ctx, request, method);
        } else if (uri.startsWith("/rules/")) {
            handleRuleById(ctx, request, method, uri);
        } else if (uri.equals("/stats")) {
            handleStats(ctx, request, method);
        } else {
            sendError(ctx, HttpResponseStatus.NOT_FOUND, "Not Found");
        }
    }

    private void handleVersions(ChannelHandlerContext ctx, FullHttpRequest request, HttpMethod method) throws Exception {
        if (method.equals(HttpMethod.GET)) {
            Map<String, Object> result = new HashMap<>();
            for (Version v : versionManager.getVersionList()) {
                Map<String, Object> versionData = new HashMap<>();
                versionData.put("name", v.getName());
                versionData.put("host", v.getHost() + ":" + v.getPort());
                versionData.put("weight", v.getWeight());
                result.put(v.getName(), versionData);
            }
            sendJsonResponse(ctx, HttpResponseStatus.OK, result);
        } else if (method.equals(HttpMethod.POST)) {
            String body = request.content().toString(CharsetUtil.UTF_8);
            Version version = parseVersionFromBody(body);
            if (version == null || version.getName() == null) {
                sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Invalid version data");
                return;
            }
            if (versionManager.hasVersion(version.getName())) {
                sendError(ctx, HttpResponseStatus.CONFLICT, "Version already exists");
                return;
            }
            versionManager.addVersion(version);
            sendJsonResponse(ctx, HttpResponseStatus.CREATED, version);
        } else {
            sendError(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, "Method not allowed");
        }
    }

    private void handleVersionById(ChannelHandlerContext ctx, FullHttpRequest request, HttpMethod method, String uri) throws Exception {
        String name = uri.substring("/versions/".length());

        if (method.equals(HttpMethod.DELETE)) {
            List<Rule> rules = ruleManager.getRulesForVersion(name);
            if (!rules.isEmpty()) {
                sendError(ctx, HttpResponseStatus.CONFLICT, "Version has active rules, cannot delete");
                return;
            }
            if (versionManager.removeVersion(name)) {
                sendResponse(ctx, HttpResponseStatus.NO_CONTENT, "");
            } else {
                sendError(ctx, HttpResponseStatus.NOT_FOUND, "Version not found");
            }
        } else {
            sendError(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, "Method not allowed");
        }
    }

    private void handleRules(ChannelHandlerContext ctx, FullHttpRequest request, HttpMethod method) throws Exception {
        if (method.equals(HttpMethod.GET)) {
            sendJsonResponse(ctx, HttpResponseStatus.OK, ruleManager.getAllRules());
        } else if (method.equals(HttpMethod.POST)) {
            String body = request.content().toString(CharsetUtil.UTF_8);
            Rule rule = parseRuleFromBody(body);
            if (rule == null) {
                sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Invalid rule data");
                return;
            }
            if (rule.getId() == null || rule.getId().isEmpty()) {
                rule.setId(UUID.randomUUID().toString());
            }
            if (ruleManager.hasRule(rule.getId())) {
                sendError(ctx, HttpResponseStatus.CONFLICT, "Rule with this ID already exists");
                return;
            }
            if (!versionManager.hasVersion(rule.getTargetVersion())) {
                sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Target version does not exist");
                return;
            }
            ruleManager.addRule(rule);
            sendJsonResponse(ctx, HttpResponseStatus.CREATED, rule);
        } else {
            sendError(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, "Method not allowed");
        }
    }

    private void handleRuleById(ChannelHandlerContext ctx, FullHttpRequest request, HttpMethod method, String uri) throws Exception {
        String id = uri.substring("/rules/".length());

        if (method.equals(HttpMethod.PUT)) {
            String body = request.content().toString(CharsetUtil.UTF_8);
            Rule newRule = parseRuleFromBody(body);
            if (newRule == null) {
                sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Invalid rule data");
                return;
            }
            newRule.setId(id);
            if (!versionManager.hasVersion(newRule.getTargetVersion())) {
                sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Target version does not exist");
                return;
            }
            if (ruleManager.updateRule(id, newRule)) {
                sendJsonResponse(ctx, HttpResponseStatus.OK, newRule);
            } else {
                sendError(ctx, HttpResponseStatus.NOT_FOUND, "Rule not found");
            }
        } else if (method.equals(HttpMethod.DELETE)) {
            if (ruleManager.removeRule(id)) {
                sendResponse(ctx, HttpResponseStatus.NO_CONTENT, "");
            } else {
                sendError(ctx, HttpResponseStatus.NOT_FOUND, "Rule not found");
            }
        } else if (method.equals(HttpMethod.GET)) {
            Rule rule = ruleManager.getRule(id);
            if (rule != null) {
                sendJsonResponse(ctx, HttpResponseStatus.OK, rule);
            } else {
                sendError(ctx, HttpResponseStatus.NOT_FOUND, "Rule not found");
            }
        } else {
            sendError(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, "Method not allowed");
        }
    }

    private void handleStats(ChannelHandlerContext ctx, FullHttpRequest request, HttpMethod method) throws Exception {
        if (!method.equals(HttpMethod.GET)) {
            sendError(ctx, HttpResponseStatus.METHOD_NOT_ALLOWED, "Method not allowed");
            return;
        }

        Map<String, Map<String, Long>> result = new HashMap<>();
        for (Map.Entry<String, Stats> entry : statsManager.getAllStats().entrySet()) {
            Map<String, Long> data = new HashMap<>();
            data.put("requestCount", entry.getValue().getRequestCount());
            data.put("errorCount", entry.getValue().getErrorCount());
            result.put(entry.getKey(), data);
        }
        sendJsonResponse(ctx, HttpResponseStatus.OK, result);
    }

    private Version parseVersionFromBody(String body) {
        try {
            Map<String, Object> map = JsonUtil.fromJson(body, Map.class);
            String name = (String) map.get("name");
            String hostStr = (String) map.get("host");
            Integer weight = (Integer) map.get("weight");

            if (name == null || hostStr == null || weight == null) {
                return null;
            }

            String[] hostParts = hostStr.split(":");
            String host = hostParts[0];
            int port = hostParts.length > 1 ? Integer.parseInt(hostParts[1]) : 80;

            return new Version(name, host, port, weight);
        } catch (Exception e) {
            return null;
        }
    }

    private Rule parseRuleFromBody(String body) {
        try {
            Map<String, Object> map = JsonUtil.fromJson(body, Map.class);
            String id = (String) map.get("id");
            String typeStr = (String) map.get("type");
            String key = (String) map.get("key");
            String value = (String) map.get("value");
            String targetVersion = (String) map.get("targetVersion");

            if (typeStr == null || value == null || targetVersion == null) {
                return null;
            }

            Rule.RuleType type = Rule.RuleType.valueOf(typeStr.toUpperCase());
            return new Rule(id, type, key, value, targetVersion);
        } catch (Exception e) {
            return null;
        }
    }

    private void sendJsonResponse(ChannelHandlerContext ctx, HttpResponseStatus status, Object data) throws Exception {
        String json = JsonUtil.toJson(data);
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer(json, CharsetUtil.UTF_8));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    private void sendResponse(ChannelHandlerContext ctx, HttpResponseStatus status, String content) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer(content, CharsetUtil.UTF_8));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "text/plain; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    private void sendError(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        sendResponse(ctx, status, message);
    }
}
