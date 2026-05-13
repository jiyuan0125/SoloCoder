package com.example.gateway;

import com.example.gateway.client.HttpClientPool;
import com.example.gateway.config.ConfigLoader;
import com.example.gateway.conversion.ProtocolConverter;
import com.example.gateway.handler.GatewayRequestHandler;
import com.example.gateway.manager.AdminApiManager;
import com.example.gateway.model.GatewayConfig;
import com.example.gateway.ratelimit.RateLimitManager;
import com.example.gateway.route.RouteManager;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelOption;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.stream.ChunkedWriteHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ApiGatewayServer {
    private static final Logger logger = LoggerFactory.getLogger(ApiGatewayServer.class);

    private final GatewayConfig config;
    private final RouteManager routeManager;
    private final RateLimitManager rateLimitManager;
    private final ProtocolConverter converter;
    private final HttpClientPool httpClientPool;
    private final AdminApiManager adminApiManager;

    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;

    public ApiGatewayServer() {
        this.config = ConfigLoader.load();
        this.routeManager = new RouteManager();
        this.rateLimitManager = new RateLimitManager();
        this.converter = new ProtocolConverter();
        this.httpClientPool = new HttpClientPool(config.getClientTimeoutSeconds());
        this.adminApiManager = new AdminApiManager(routeManager, rateLimitManager);
    }

    public void start() throws InterruptedException {
        logger.info("Starting API Gateway Server on port {}", config.getServerPort());

        bossGroup = new NioEventLoopGroup(1);
        workerGroup = new NioEventLoopGroup();

        try {
            ServerBootstrap bootstrap = new ServerBootstrap()
                    .group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class)
                    .option(ChannelOption.SO_BACKLOG, 1024)
                    .childOption(ChannelOption.SO_KEEPALIVE, true)
                    .childHandler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) {
                            ch.pipeline().addLast(
                                    new HttpServerCodec(),
                                    new HttpObjectAggregator(1024 * 1024 * 10),
                                    new ChunkedWriteHandler(),
                                    new GatewayRequestHandler(
                                            routeManager,
                                            rateLimitManager,
                                            converter,
                                            httpClientPool,
                                            adminApiManager
                                    )
                            );
                        }
                    });

            ChannelFuture future = bootstrap.bind(config.getServerPort()).sync();
            logger.info("API Gateway Server started successfully on port {}", config.getServerPort());
            logger.info("Admin endpoints:");
            logger.info("  GET  /_admin/health");
            logger.info("  GET  /_admin/routes");
            logger.info("  POST /_admin/routes");
            logger.info("  DELETE /_admin/routes/{path}");
            logger.info("  GET  /_admin/ratelimit");
            logger.info("  POST /_admin/ratelimit");

            future.channel().closeFuture().sync();
        } finally {
            shutdown();
        }
    }

    public void shutdown() {
        logger.info("Shutting down API Gateway Server...");
        if (httpClientPool != null) {
            httpClientPool.shutdown();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully();
        }
        if (bossGroup != null) {
            bossGroup.shutdownGracefully();
        }
        logger.info("API Gateway Server stopped");
    }

    public RouteManager getRouteManager() {
        return routeManager;
    }

    public RateLimitManager getRateLimitManager() {
        return rateLimitManager;
    }

    public static void main(String[] args) {
        ApiGatewayServer server = new ApiGatewayServer();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            server.shutdown();
        }));

        try {
            server.start();
        } catch (InterruptedException e) {
            logger.error("Server interrupted", e);
            Thread.currentThread().interrupt();
        }
    }
}
