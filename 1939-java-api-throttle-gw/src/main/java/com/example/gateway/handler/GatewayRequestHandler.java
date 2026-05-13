package com.example.gateway.handler;

import com.example.gateway.client.ChannelHandlerAdapter;
import com.example.gateway.client.HttpClientPool;
import com.example.gateway.conversion.ProtocolConverter;
import com.example.gateway.manager.AdminApiManager;
import com.example.gateway.model.RouteMatchResult;
import com.example.gateway.ratelimit.RateLimitManager;
import com.example.gateway.route.RouteManager;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.DefaultFullHttpRequest;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpHeaderValues;
import io.netty.handler.codec.http.HttpHeaders;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpVersion;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.charset.StandardCharsets;
import java.util.Map;

public class GatewayRequestHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
    private static final Logger logger = LoggerFactory.getLogger(GatewayRequestHandler.class);

    private final RouteManager routeManager;
    private final RateLimitManager rateLimitManager;
    private final ProtocolConverter converter;
    private final HttpClientPool httpClientPool;
    private final AdminApiManager adminApiManager;

    public GatewayRequestHandler(RouteManager routeManager, RateLimitManager rateLimitManager,
                                 ProtocolConverter converter, HttpClientPool httpClientPool,
                                 AdminApiManager adminApiManager) {
        this.routeManager = routeManager;
        this.rateLimitManager = rateLimitManager;
        this.converter = converter;
        this.httpClientPool = httpClientPool;
        this.adminApiManager = adminApiManager;
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
        String uri = request.uri();
        HttpMethod method = request.method();

        if (adminApiManager.isAdminRequest(uri, method)) {
            handleAdminRequest(ctx, request);
            return;
        }

        RouteMatchResult matchResult = routeManager.matchRoute(uri);

        if (!matchResult.isMatched()) {
            sendError(ctx, HttpResponseStatus.NOT_FOUND, "No route matched for: " + uri);
            return;
        }

        String routePath = matchResult.getRoute().getPath();
        if (!rateLimitManager.tryAcquire(routePath)) {
            double waitTime = rateLimitManager.getWaitTimeSeconds(routePath);
            sendRateLimited(ctx, waitTime);
            return;
        }

        handleForwardRequest(ctx, request, matchResult);
    }

    private void handleAdminRequest(ChannelHandlerContext ctx, FullHttpRequest request) {
        FullHttpResponse response = adminApiManager.handleRequest(request);
        ctx.writeAndFlush(response);
    }

    private void handleForwardRequest(ChannelHandlerContext ctx, FullHttpRequest request, RouteMatchResult matchResult) {
        Map<String, String> pathParams = matchResult.getPathParams();
        String targetPath = matchResult.getTargetPath();
        Map<String, String> targetHost = matchResult.getRoute().getTargetHost();

        String query = extractQuery(request.uri());
        String finalTargetPath = query != null ? targetPath + "?" + query : targetPath;

        String contentType = request.headers().get(HttpHeaderNames.CONTENT_TYPE);
        String accept = request.headers().get(HttpHeaderNames.ACCEPT);
        HttpMethod method = request.method();

        String requestBody = request.content().toString(StandardCharsets.UTF_8);
        String convertedRequestBody = requestBody;

        if (!method.equals(HttpMethod.GET) && contentType != null) {
            if (contentType.contains("application/xml")) {
                ProtocolConverter.ConversionResult result = converter.xmlToJson(requestBody);
                if (!result.isSuccess()) {
                    sendError(ctx, HttpResponseStatus.BAD_REQUEST, "Request conversion failed: " + result.getError());
                    return;
                }
                convertedRequestBody = result.getContent();
                contentType = "application/json";
            }
        }

        String host = targetHost.get("host");
        int port = Integer.parseInt(targetHost.get("port"));

        FullHttpRequest forwardRequest = new DefaultFullHttpRequest(
                HttpVersion.HTTP_1_1,
                method,
                finalTargetPath,
                Unpooled.copiedBuffer(convertedRequestBody, StandardCharsets.UTF_8)
        );

        HttpHeaders headers = forwardRequest.headers();
        headers.set(HttpHeaderNames.HOST, host + ":" + port);
        headers.set(HttpHeaderNames.CONTENT_LENGTH, forwardRequest.content().readableBytes());
        headers.set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);

        if (contentType != null) {
            headers.set(HttpHeaderNames.CONTENT_TYPE, contentType);
        }
        if (accept != null) {
            headers.set(HttpHeaderNames.ACCEPT, "application/json");
        }

        for (Map.Entry<String, String> entry : pathParams.entrySet()) {
            headers.set("X-Path-Param-" + entry.getKey(), entry.getValue());
        }

        final String finalAccept = accept;
        ChannelHandlerAdapter clientHandler = new ChannelHandlerAdapter() {
            @Override
            public void onResponse(FullHttpResponse backendResponse) {
                handleBackendResponse(ctx, backendResponse, finalAccept);
            }

            @Override
            public void onTimeout() {
                sendError(ctx, HttpResponseStatus.GATEWAY_TIMEOUT, "Backend request timed out");
            }

            @Override
            public void onError(Throwable cause) {
                sendError(ctx, HttpResponseStatus.BAD_GATEWAY, "Backend error: " + cause.getMessage());
            }
        };

        clientHandler.setForwardRequest(forwardRequest);
        httpClientPool.connect(host, port, clientHandler);
    }

    private void handleBackendResponse(ChannelHandlerContext ctx, FullHttpResponse backendResponse, String accept) {
        String responseBody = backendResponse.content().toString(StandardCharsets.UTF_8);
        String convertedResponseBody = responseBody;
        String responseContentType = backendResponse.headers().get(HttpHeaderNames.CONTENT_TYPE);

        if (accept != null && accept.contains("application/xml") &&
                (responseContentType == null || responseContentType.contains("application/json"))) {
            ProtocolConverter.ConversionResult result = converter.jsonToXml(responseBody);
            if (result.isSuccess()) {
                convertedResponseBody = result.getContent();
                responseContentType = "application/xml";
            } else {
                logger.warn("Response conversion failed, returning original: {}", result.getError());
            }
        }

        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1,
                backendResponse.status(),
                Unpooled.copiedBuffer(convertedResponseBody, StandardCharsets.UTF_8)
        );

        HttpHeaders headers = response.headers();
        headers.set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        if (responseContentType != null) {
            headers.set(HttpHeaderNames.CONTENT_TYPE, responseContentType);
        }
        headers.set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);

        ctx.writeAndFlush(response);
    }

    private String extractQuery(String uri) {
        int idx = uri.indexOf('?');
        if (idx >= 0) {
            return uri.substring(idx + 1);
        }
        return null;
    }

    private void sendError(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        String body = "{\"error\": \"" + status.code() + "\", \"message\": \"" + message + "\"}";
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1,
                status,
                Unpooled.copiedBuffer(body, CharsetUtil.UTF_8)
        );
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response);
    }

    private void sendRateLimited(ChannelHandlerContext ctx, double waitTimeSeconds) {
        int retryAfter = (int) Math.ceil(waitTimeSeconds);
        if (retryAfter < 1) retryAfter = 1;

        String body = "{\"error\": \"429\", \"message\": \"Rate limit exceeded\", \"retryAfter\": " + retryAfter + "}";
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1,
                HttpResponseStatus.TOO_MANY_REQUESTS,
                Unpooled.copiedBuffer(body, CharsetUtil.UTF_8)
        );
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        response.headers().set("Retry-After", String.valueOf(retryAfter));
        ctx.writeAndFlush(response);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Gateway handler error", cause);
        sendError(ctx, HttpResponseStatus.INTERNAL_SERVER_ERROR, cause.getMessage());
        ctx.close();
    }
}
