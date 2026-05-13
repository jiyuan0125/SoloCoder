package com.example.gateway.client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelOption;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.HttpClientCodec;
import io.netty.handler.codec.http.HttpContentDecompressor;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.timeout.ReadTimeoutHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

public class HttpClientPool {
    private static final Logger logger = LoggerFactory.getLogger(HttpClientPool.class);

    private final EventLoopGroup group = new NioEventLoopGroup();
    private final Bootstrap bootstrap;
    private final int timeoutSeconds;
    private final Map<String, Channel> connections = new ConcurrentHashMap<>();

    public HttpClientPool(int timeoutSeconds) {
        this.timeoutSeconds = timeoutSeconds;
        this.bootstrap = new Bootstrap()
                .group(group)
                .channel(NioSocketChannel.class)
                .option(ChannelOption.SO_KEEPALIVE, true)
                .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, timeoutSeconds * 1000);
    }

    public void connect(String host, int port, ChannelHandlerAdapter clientHandler) {
        String key = host + ":" + port;
        Channel existing = connections.get(key);

        if (existing != null && existing.isActive()) {
            clientHandler.setClientChannel(existing);
            existing.pipeline().addLast(clientHandler);
            clientHandler.onChannelReady();
            return;
        }

        ChannelInitializer<SocketChannel> initializer = new ChannelInitializer<SocketChannel>() {
            @Override
            protected void initChannel(SocketChannel ch) {
                ch.pipeline().addLast(
                        new ReadTimeoutHandler(timeoutSeconds, TimeUnit.SECONDS),
                        new HttpClientCodec(),
                        new HttpObjectAggregator(1024 * 1024 * 10),
                        new HttpContentDecompressor()
                );
            }
        };

        bootstrap.handler(initializer);

        ChannelFuture future = bootstrap.connect(host, port);
        future.addListener((ChannelFutureListener) f -> {
            if (f.isSuccess()) {
                Channel channel = f.channel();
                connections.put(key, channel);
                channel.closeFuture().addListener((ChannelFutureListener) closeFuture -> {
                    connections.remove(key);
                });
                clientHandler.setClientChannel(channel);
                channel.pipeline().addLast(clientHandler);
            } else {
                clientHandler.onConnectionFailure(f.cause());
            }
        });
    }

    public void shutdown() {
        group.shutdownGracefully();
    }

    public int getTimeoutSeconds() {
        return timeoutSeconds;
    }
}
