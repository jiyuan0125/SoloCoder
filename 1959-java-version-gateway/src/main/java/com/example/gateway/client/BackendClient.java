package com.example.gateway.client;

import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.*;

import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

public class BackendClient {
    private final EventLoopGroup workerGroup = new NioEventLoopGroup();
    private final Map<String, Channel> channelPool = new ConcurrentHashMap<>();

    public CompletableFuture<BackendResponse> forward(String backendUrl,
                                                      String method,
                                                      String path,
                                                      Map<String, String> headers,
                                                      String body) {
        CompletableFuture<BackendResponse> future = new CompletableFuture<>();
        try {
            URI uri = new URI(backendUrl);
            String host = uri.getHost();
            int port = uri.getPort() > 0 ? uri.getPort() : 80;

            Bootstrap b = new Bootstrap();
            b.group(workerGroup)
                    .channel(NioSocketChannel.class)
                    .option(ChannelOption.SO_KEEPALIVE, true)
                    .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, 10000)
                    .handler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) {
                            ChannelPipeline p = ch.pipeline();
                            p.addLast(new HttpClientCodec());
                            p.addLast(new HttpObjectAggregator(1024 * 1024));
                            p.addLast(new BackendResponseHandler(future));
                        }
                    });

            ChannelFuture connectFuture = b.connect(host, port).sync();
            Channel ch = connectFuture.channel();

            FullHttpRequest request = buildRequest(method, path, host, headers, body);
            ch.writeAndFlush(request);

        } catch (Exception e) {
            future.completeExceptionally(e);
        }
        return future;
    }

    private FullHttpRequest buildRequest(String method, String path, String host,
                                         Map<String, String> headers, String body) {
        HttpMethod httpMethod = HttpMethod.valueOf(method.toUpperCase());
        byte[] content = body != null ? body.getBytes(StandardCharsets.UTF_8) : new byte[0];

        FullHttpRequest request = new DefaultFullHttpRequest(
                HttpVersion.HTTP_1_1,
                httpMethod,
                path,
                Unpooled.wrappedBuffer(content)
        );

        request.headers().set(HttpHeaderNames.HOST, host);
        request.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.CLOSE);
        request.headers().set(HttpHeaderNames.CONTENT_LENGTH, content.length);

        if (headers != null) {
            for (Map.Entry<String, String> entry : headers.entrySet()) {
                String name = entry.getKey();
                if (!HttpHeaderNames.CONTENT_LENGTH.contentEqualsIgnoreCase(name)
                        && !HttpHeaderNames.HOST.contentEqualsIgnoreCase(name)
                        && !HttpHeaderNames.CONNECTION.contentEqualsIgnoreCase(name)) {
                    request.headers().set(name, entry.getValue());
                }
            }
        }

        return request;
    }

    public void shutdown() {
        workerGroup.shutdownGracefully();
    }

    public static class BackendResponse {
        private final int status;
        private final String body;
        private final HttpHeaders headers;

        public BackendResponse(int status, String body, HttpHeaders headers) {
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

        public HttpHeaders getHeaders() {
            return headers;
        }
    }

    private static class BackendResponseHandler extends SimpleChannelInboundHandler<FullHttpResponse> {
        private final CompletableFuture<BackendResponse> future;

        BackendResponseHandler(CompletableFuture<BackendResponse> future) {
            this.future = future;
        }

        @Override
        protected void channelRead0(ChannelHandlerContext ctx, FullHttpResponse msg) {
            int status = msg.status().code();
            String body = msg.content().toString(StandardCharsets.UTF_8);
            HttpHeaders headers = msg.headers().copy();
            future.complete(new BackendResponse(status, body, headers));
            ctx.close();
        }

        @Override
        public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
            future.completeExceptionally(cause);
            ctx.close();
        }
    }
}
