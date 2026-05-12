package com.gateway.handler;

import com.gateway.engine.RuleMatchingEngine;
import com.gateway.engine.TrafficDistributionEngine;
import com.gateway.manager.StatsManager;
import com.gateway.manager.VersionManager;
import com.gateway.model.Version;
import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.*;
import io.netty.util.CharsetUtil;

import java.net.InetSocketAddress;
import java.util.HashMap;
import java.util.Map;

public class GatewayHandler extends SimpleChannelInboundHandler<FullHttpRequest> {

    private final VersionManager versionManager;
    private final RuleMatchingEngine ruleMatchingEngine;
    private final TrafficDistributionEngine trafficDistributionEngine;
    private final StatsManager statsManager;
    private final EventLoopGroup workerGroup;

    public GatewayHandler(VersionManager versionManager,
                          RuleMatchingEngine ruleMatchingEngine,
                          TrafficDistributionEngine trafficDistributionEngine,
                          StatsManager statsManager,
                          EventLoopGroup workerGroup) {
        this.versionManager = versionManager;
        this.ruleMatchingEngine = ruleMatchingEngine;
        this.trafficDistributionEngine = trafficDistributionEngine;
        this.statsManager = statsManager;
        this.workerGroup = workerGroup;
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) throws Exception {
        String uri = request.uri();

        if (isManagementApi(uri)) {
            ctx.fireChannelRead(request.retain());
            return;
        }

        Map<String, String> headers = extractHeaders(request);
        Map<String, String> cookies = extractCookies(request);
        String clientIp = getClientIp(ctx, request);

        String selectedVersion = ruleMatchingEngine.matchVersion(headers, cookies, clientIp);
        if (selectedVersion == null) {
            selectedVersion = trafficDistributionEngine.selectVersionByWeight();
        }

        if (selectedVersion == null || !versionManager.hasVersion(selectedVersion)) {
            sendError(ctx, HttpResponseStatus.SERVICE_UNAVAILABLE, "No available backend version");
            return;
        }

        Version version = versionManager.getVersion(selectedVersion);
        statsManager.incrementRequestCount(selectedVersion);

        HttpRequest newRequest = new DefaultFullHttpRequest(
                request.protocolVersion(), request.method(), request.uri(),
                request.content().retain());
        newRequest.headers().set(request.headers());
        newRequest.headers().set("X-Gateway-Version", selectedVersion);

        forwardRequest(ctx, newRequest, version, selectedVersion);
    }

    private boolean isManagementApi(String uri) {
        return uri.equals("/versions") || uri.startsWith("/versions/") ||
               uri.equals("/rules") || uri.startsWith("/rules/") ||
               uri.equals("/stats");
    }

    private Map<String, String> extractHeaders(HttpRequest request) {
        Map<String, String> headers = new HashMap<>();
        for (Map.Entry<String, String> entry : request.headers()) {
            headers.put(entry.getKey(), entry.getValue());
        }
        return headers;
    }

    private Map<String, String> extractCookies(HttpRequest request) {
        Map<String, String> cookies = new HashMap<>();
        String cookieHeader = request.headers().get(HttpHeaderNames.COOKIE);
        if (cookieHeader != null) {
            for (String cookie : cookieHeader.split(";")) {
                cookie = cookie.trim();
                int idx = cookie.indexOf('=');
                if (idx > 0) {
                    cookies.put(cookie.substring(0, idx), cookie.substring(idx + 1));
                }
            }
        }
        return cookies;
    }

    private String getClientIp(ChannelHandlerContext ctx, HttpRequest request) {
        String xForwardedFor = request.headers().get("X-Forwarded-For");
        if (xForwardedFor != null && !xForwardedFor.isEmpty()) {
            return xForwardedFor.split(",")[0].trim();
        }
        String xRealIp = request.headers().get("X-Real-IP");
        if (xRealIp != null && !xRealIp.isEmpty()) {
            return xRealIp;
        }
        InetSocketAddress remoteAddress = (InetSocketAddress) ctx.channel().remoteAddress();
        return remoteAddress.getAddress().getHostAddress();
    }

    private void forwardRequest(ChannelHandlerContext clientCtx, HttpRequest request,
                                Version version, String versionName) {
        Bootstrap b = new Bootstrap();
        b.group(workerGroup)
         .channel(NioSocketChannel.class)
         .handler(new ChannelInitializer<SocketChannel>() {
             @Override
             protected void initChannel(SocketChannel ch) {
                 ChannelPipeline p = ch.pipeline();
                 p.addLast(new HttpClientCodec());
                 p.addLast(new HttpObjectAggregator(1024 * 1024));
                 p.addLast(new BackendResponseHandler(clientCtx, versionName, statsManager));
             }
         });

        ChannelFuture connectFuture = b.connect(version.getHost(), version.getPort());
        connectFuture.addListener((ChannelFutureListener) future -> {
            if (future.isSuccess()) {
                future.channel().writeAndFlush(request);
            } else {
                statsManager.incrementErrorCount(versionName);
                sendError(clientCtx, HttpResponseStatus.BAD_GATEWAY, "Failed to connect to backend");
            }
        });
    }

    private void sendError(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        sendErrorStatic(ctx, status, message);
    }

    private static void sendErrorStatic(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer(message, CharsetUtil.UTF_8));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "text/plain; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    private static class BackendResponseHandler extends SimpleChannelInboundHandler<FullHttpResponse> {
        private final ChannelHandlerContext clientCtx;
        private final String versionName;
        private final StatsManager statsManager;

        public BackendResponseHandler(ChannelHandlerContext clientCtx, String versionName, StatsManager statsManager) {
            this.clientCtx = clientCtx;
            this.versionName = versionName;
            this.statsManager = statsManager;
        }

        @Override
        protected void channelRead0(ChannelHandlerContext ctx, FullHttpResponse response) {
            if (response.status().code() >= 500) {
                statsManager.incrementErrorCount(versionName);
            }

            FullHttpResponse newResponse = new DefaultFullHttpResponse(
                    response.protocolVersion(), response.status(),
                    response.content().retain());
            newResponse.headers().set(response.headers());
            newResponse.headers().set("X-Gateway-Version", versionName);

            clientCtx.writeAndFlush(newResponse).addListener(ChannelFutureListener.CLOSE);
            ctx.close();
        }

        @Override
        public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
            statsManager.incrementErrorCount(versionName);
            ctx.close();
            sendErrorStatic(clientCtx, HttpResponseStatus.BAD_GATEWAY, "Backend error: " + cause.getMessage());
        }
    }
}
