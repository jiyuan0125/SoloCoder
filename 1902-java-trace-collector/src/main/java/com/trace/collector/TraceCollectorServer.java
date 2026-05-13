package com.trace.collector;

import com.trace.collector.handler.HttpRequestHandler;
import com.trace.collector.server.HttpServerInitializer;
import com.trace.collector.service.SpanService;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelOption;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TraceCollectorServer {

    private static final Logger logger = LoggerFactory.getLogger(TraceCollectorServer.class);

    private static final int DEFAULT_PORT = 9100;

    private final int port;
    private final SpanService spanService;

    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;
    private Channel channel;

    public TraceCollectorServer(int port) {
        this.port = port;
        this.spanService = new SpanService();
    }

    public void start() throws Exception {
        spanService.start();

        bossGroup = new NioEventLoopGroup(1);
        workerGroup = new NioEventLoopGroup();

        try {
            HttpRequestHandler httpRequestHandler = new HttpRequestHandler(spanService);

            ServerBootstrap b = new ServerBootstrap();
            b.group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class)
                    .childHandler(new HttpServerInitializer(httpRequestHandler))
                    .option(ChannelOption.SO_BACKLOG, 128)
                    .childOption(ChannelOption.SO_KEEPALIVE, true);

            channel = b.bind(port).sync().channel();
            logger.info("Trace Collector Server started on port {}", port);
        } catch (Exception e) {
            logger.error("Failed to start server", e);
            stop();
            throw e;
        }
    }

    public void awaitTermination() throws InterruptedException {
        if (channel != null) {
            channel.closeFuture().sync();
        }
    }

    public void stop() {
        logger.info("Shutting down Trace Collector Server...");

        if (channel != null) {
            channel.close();
        }

        if (bossGroup != null) {
            bossGroup.shutdownGracefully();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully();
        }

        if (spanService != null) {
            spanService.stop();
        }

        logger.info("Trace Collector Server shut down");
    }

    public static void main(String[] args) {
        int port = DEFAULT_PORT;
        String portEnv = System.getenv("PORT");
        if (portEnv != null && !portEnv.isEmpty()) {
            try {
                port = Integer.parseInt(portEnv);
            } catch (NumberFormatException e) {
                logger.warn("Invalid PORT environment variable: {}, using default {}", portEnv, DEFAULT_PORT);
            }
        }

        TraceCollectorServer server = new TraceCollectorServer(port);

        Runtime.getRuntime().addShutdownHook(new Thread(server::stop, "shutdown-hook"));

        try {
            server.start();
            server.awaitTermination();
        } catch (Exception e) {
            logger.error("Server error", e);
            System.exit(1);
        }
    }
}
