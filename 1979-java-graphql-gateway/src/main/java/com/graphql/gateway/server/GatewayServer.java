package com.graphql.gateway.server;

import com.graphql.gateway.config.GatewayConfig;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;

public class GatewayServer {
    private final GatewayConfig config;
    private Channel channel;
    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;

    public GatewayServer() {
        this.config = GatewayConfig.getInstance();
    }

    public void start() throws InterruptedException {
        bossGroup = new NioEventLoopGroup(1);
        workerGroup = new NioEventLoopGroup();

        try {
            ServerBootstrap b = new ServerBootstrap();
            b.group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class)
                    .handler(new LoggingHandler(LogLevel.INFO))
                    .childHandler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) {
                            ChannelPipeline p = ch.pipeline();
                            p.addLast(new HttpServerCodec());
                            p.addLast(new HttpObjectAggregator(1024 * 1024));
                            p.addLast(new HttpServerHandler());
                        }
                    });

            int port = config.getServerPort();
            String host = config.getServerHost();

            channel = b.bind(host, port).sync().channel();
            System.out.println("GraphQL Gateway started on " + host + ":" + port);
            System.out.println("Available endpoints:");
            System.out.println("  POST /graphql  - Execute GraphQL queries");
            System.out.println("  POST /schemas  - Register backend schema");
            System.out.println("  GET  /schema   - View merged schema");
            System.out.println("  GET  /stats    - View query statistics");
            System.out.println("  GET  /health   - Health check");

            channel.closeFuture().sync();
        } finally {
            stop();
        }
    }

    public void stop() {
        if (channel != null) {
            channel.close();
        }
        if (bossGroup != null) {
            bossGroup.shutdownGracefully();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully();
        }
        System.out.println("GraphQL Gateway stopped");
    }

    public static void main(String[] args) throws InterruptedException {
        new GatewayServer().start();
    }
}
