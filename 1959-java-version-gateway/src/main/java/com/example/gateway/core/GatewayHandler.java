package com.example.gateway.core;

import com.example.gateway.api.ManagementApi;
import com.example.gateway.client.BackendClient;
import com.example.gateway.compatibility.CompatibilityService;
import com.example.gateway.routing.RouteRule;
import com.example.gateway.routing.RouteStore;
import com.example.gateway.routing.VersionExtractor;
import com.example.gateway.stats.StatsStore;
import com.example.gateway.util.JsonUtils;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;

import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

public class GatewayHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
    private final GatewayContext context;
    private final ManagementApi managementApi;

    public GatewayHandler(GatewayContext context) {
        this.context = context;
        this.managementApi = new ManagementApi(context);
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest req) {
        String method = req.method().name();
        String uri = req.uri();
        String path = extractPath(uri);
        String body = req.content().toString(StandardCharsets.UTF_8);

        ManagementApi.ApiResponse apiResp = managementApi.handle(method, path, body);
        if (apiResp != null) {
            writeResponse(ctx, req, apiResp.getStatus(), apiResp.getBody(), apiResp.getHeaders());
            return;
        }

        handleProxy(ctx, req, method, path, body);
    }

    private void handleProxy(ChannelHandlerContext ctx, FullHttpRequest req,
                             String method, String path, String body) {
        long startTime = System.currentTimeMillis();
        String defaultVersion = context.getConfig().getDefaultVersion();
        VersionExtractor.VersionPath vp = VersionExtractor.extract(path, defaultVersion);
        String version = vp.getVersion();
        String remainingPath = vp.getRemainingPath();

        RouteStore routeStore = context.getRouteStore();
        if (!routeStore.hasRoute(version)) {
            writeError(ctx, req, HttpResponseStatus.NOT_FOUND,
                    "No route configured for version: " + version);
            return;
        }

        RouteRule rule = routeStore.getRoute(version);
        String backend = routeStore.pickBackend(version);
        if (backend == null) {
            writeError(ctx, req, HttpResponseStatus.SERVICE_UNAVAILABLE,
                    "No backend available for version: " + version);
            return;
        }

        String transformedRequestBody = body;
        CompatibilityService compat = context.getCompatibilityService();
        if (rule.isCompatibilityEnabled() && compat.hasRule(version)) {
            transformedRequestBody = compat.transformRequest(version, body);
        }

        Map<String, String> headers = new HashMap<>();
        for (Map.Entry<String, String> entry : req.headers()) {
            headers.put(entry.getKey(), entry.getValue());
        }

        BackendClient client = context.getBackendClient();
        CompletableFuture<BackendClient.BackendResponse> future =
                client.forward(backend, method, remainingPath, headers, transformedRequestBody);

        future.whenComplete((resp, ex) -> {
            int status = 500;
            String responseBody = "";
            Map<String, String> respHeaders = new HashMap<>();

            if (ex != null) {
                status = 502;
                responseBody = errorJson("Backend error: " + ex.getMessage());
                respHeaders.put(HttpHeaderNames.CONTENT_TYPE.toString(), "application/json; charset=UTF-8");
            } else {
                status = resp.getStatus();
                String originalBody = resp.getBody();
                if (rule.isCompatibilityEnabled() && compat.hasRule(version)) {
                    responseBody = compat.transformResponse(version, originalBody);
                } else {
                    responseBody = originalBody;
                }
                if (resp.getHeaders() != null) {
                    for (Map.Entry<String, String> e : resp.getHeaders()) {
                        String name = e.getKey();
                        if (!HttpHeaderNames.CONTENT_LENGTH.contentEqualsIgnoreCase(name)
                                && !HttpHeaderNames.TRANSFER_ENCODING.contentEqualsIgnoreCase(name)
                                && !HttpHeaderNames.CONNECTION.contentEqualsIgnoreCase(name)) {
                            respHeaders.put(name, e.getValue());
                        }
                    }
                }
            }

            long responseTime = System.currentTimeMillis() - startTime;
            StatsStore stats = context.getStatsStore();
            stats.recordRequest(version, responseTime, status);

            writeResponse(ctx, req, status, responseBody, respHeaders);
        });
    }

    private String extractPath(String uri) {
        if (uri == null) {
            return "/";
        }
        int qidx = uri.indexOf('?');
        return qidx >= 0 ? uri.substring(0, qidx) : uri;
    }

    private void writeError(ChannelHandlerContext ctx, FullHttpRequest req,
                            HttpResponseStatus status, String message) {
        Map<String, String> headers = new HashMap<>();
        headers.put(HttpHeaderNames.CONTENT_TYPE.toString(), "application/json; charset=UTF-8");
        writeResponse(ctx, req, status.code(), errorJson(message), headers);
    }

    private String errorJson(String message) {
        Map<String, Object> error = new LinkedHashMap<>();
        error.put("error", message);
        return JsonUtils.toJson(error);
    }

    private void writeResponse(ChannelHandlerContext ctx, FullHttpRequest req,
                               int status, String body, Map<String, String> headers) {
        byte[] content = body != null ? body.getBytes(StandardCharsets.UTF_8) : new byte[0];
        FullHttpResponse resp = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1,
                HttpResponseStatus.valueOf(status),
                Unpooled.wrappedBuffer(content)
        );

        resp.headers().set(HttpHeaderNames.CONTENT_LENGTH, content.length);
        if (headers != null) {
            for (Map.Entry<String, String> entry : headers.entrySet()) {
                resp.headers().set(entry.getKey(), entry.getValue());
            }
        }

        boolean keepAlive = HttpUtil.isKeepAlive(req);
        if (keepAlive) {
            resp.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
            ctx.writeAndFlush(resp);
        } else {
            resp.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.CLOSE);
            ctx.writeAndFlush(resp).addListener(ChannelFutureListener.CLOSE);
        }
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        cause.printStackTrace();
        ctx.close();
    }
}
