package com.trace.collector.handler;

import com.trace.collector.model.Span;
import com.trace.collector.model.ServiceRegistration;
import com.trace.collector.model.TraceNode;
import com.trace.collector.service.SpanService;
import com.trace.collector.util.JsonUtil;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpHeaderValues;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpVersion;
import io.netty.handler.codec.http.QueryStringDecoder;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import io.netty.channel.ChannelHandler;

@ChannelHandler.Sharable
public class HttpRequestHandler extends SimpleChannelInboundHandler<FullHttpRequest> {

    private static final Logger logger = LoggerFactory.getLogger(HttpRequestHandler.class);

    private static final Pattern TRACE_BY_ID_PATTERN = Pattern.compile("^/traces/([^/?]+)$");

    private final SpanService spanService;

    public HttpRequestHandler(SpanService spanService) {
        this.spanService = spanService;
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
        try {
            String uri = request.uri();
            HttpMethod method = request.method();

            QueryStringDecoder decoder = new QueryStringDecoder(uri);
            String path = decoder.path();
            Map<String, List<String>> params = decoder.parameters();

            if (method == HttpMethod.POST && "/spans".equals(path)) {
                handlePostSpans(ctx, request);
            } else if (method == HttpMethod.GET && path.equals("/traces")) {
                handleGetTraces(ctx, params);
            } else {
                Matcher matcher = TRACE_BY_ID_PATTERN.matcher(path);
                if (method == HttpMethod.GET && matcher.matches()) {
                    handleGetTraceById(ctx, matcher.group(1));
                } else if (method == HttpMethod.POST && "/services".equals(path)) {
                    handlePostServices(ctx, request);
                } else {
                    sendResponse(ctx, HttpResponseStatus.NOT_FOUND, "Not Found");
                }
            }
        } catch (Exception e) {
            logger.error("Error processing request", e);
            sendResponse(ctx, HttpResponseStatus.INTERNAL_SERVER_ERROR, "Internal Server Error");
        }
    }

    private void handlePostSpans(ChannelHandlerContext ctx, FullHttpRequest request) {
        try {
            String body = request.content().toString(CharsetUtil.UTF_8);
            Span span = JsonUtil.fromJson(body, Span.class);

            spanService.acceptSpan(span);
            sendJsonResponse(ctx, HttpResponseStatus.ACCEPTED, "{\"status\":\"accepted\"}");
        } catch (Exception e) {
            logger.error("Error processing span", e);
            sendResponse(ctx, HttpResponseStatus.BAD_REQUEST, "Bad Request: " + e.getMessage());
        }
    }

    private void handleGetTraceById(ChannelHandlerContext ctx, String traceId) {
        List<TraceNode> traceTree = spanService.getTraceTree(traceId);
        if (traceTree.isEmpty()) {
            sendResponse(ctx, HttpResponseStatus.NOT_FOUND, "Trace not found: " + traceId);
        } else {
            String json = JsonUtil.toJson(traceTree);
            sendJsonResponse(ctx, HttpResponseStatus.OK, json);
        }
    }

    private void handleGetTraces(ChannelHandlerContext ctx, Map<String, List<String>> params) {
        String serviceName = getParam(params, "service");
        Long from = getLongParam(params, "from");
        Long to = getLongParam(params, "to");

        List<Span> spans = spanService.searchSpans(serviceName, from, to);
        String json = JsonUtil.toJson(spans);
        sendJsonResponse(ctx, HttpResponseStatus.OK, json);
    }

    private void handlePostServices(ChannelHandlerContext ctx, FullHttpRequest request) {
        try {
            String body = request.content().toString(CharsetUtil.UTF_8);
            ServiceRegistration registration = JsonUtil.fromJson(body, ServiceRegistration.class);

            if (registration.getName() == null || registration.getName().isEmpty()) {
                sendResponse(ctx, HttpResponseStatus.BAD_REQUEST, "Service name is required");
                return;
            }

            spanService.registerService(registration);
            sendJsonResponse(ctx, HttpResponseStatus.OK, "{\"status\":\"registered\"}");
        } catch (Exception e) {
            logger.error("Error registering service", e);
            sendResponse(ctx, HttpResponseStatus.BAD_REQUEST, "Bad Request: " + e.getMessage());
        }
    }

    private String getParam(Map<String, List<String>> params, String key) {
        List<String> values = params.get(key);
        if (values != null && !values.isEmpty()) {
            return values.get(0);
        }
        return null;
    }

    private Long getLongParam(Map<String, List<String>> params, String key) {
        String value = getParam(params, key);
        if (value != null && !value.isEmpty()) {
            try {
                return Long.parseLong(value);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }

    private void sendResponse(ChannelHandlerContext ctx, HttpResponseStatus status, String body) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer(body, CharsetUtil.UTF_8));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "text/plain; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    private void sendJsonResponse(ChannelHandlerContext ctx, HttpResponseStatus status, String json) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer(json, CharsetUtil.UTF_8));
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json; charset=UTF-8");
        response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Channel exception", cause);
        ctx.close();
    }
}
