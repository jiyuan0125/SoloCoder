package com.gateway;

import com.gateway.engine.RuleMatchingEngine;
import com.gateway.engine.TrafficDistributionEngine;
import com.gateway.handler.ApiHandler;
import com.gateway.handler.GatewayHandler;
import com.gateway.manager.RuleManager;
import com.gateway.manager.StatsManager;
import com.gateway.manager.VersionManager;
import com.gateway.model.Version;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class TrafficTintGateway {
    private static final Logger logger = LoggerFactory.getLogger(TrafficTintGateway.class);

    private final int port;
    private final VersionManager versionManager;
    private final RuleManager ruleManager;
    private final StatsManager statsManager;
    private final RuleMatchingEngine ruleMatchingEngine;
    private final TrafficDistributionEngine trafficDistributionEngine;

    public TrafficTintGateway(int port) {
        this.port = port;
        this.versionManager = new VersionManager();
        this.ruleManager = new RuleManager();
        this.statsManager = new StatsManager();
        this.ruleMatchingEngine = new RuleMatchingEngine(ruleManager);
        this.trafficDistributionEngine = new TrafficDistributionEngine(versionManager);
    }

    public void loadConfig(String configPath) throws IOException {
        Properties props = new Properties();
        try (InputStream is = new FileInputStream(configPath)) {
            props.load(is);
        }

        for (String key : props.stringPropertyNames()) {
            if (key.startsWith("version.") && key.endsWith(".host")) {
                String versionName = key.substring("version.".length(), key.length() - ".host".length());
                String hostPort = props.getProperty(key);
                String weightStr = props.getProperty("version." + versionName + ".weight", "0");

                String[] parts = hostPort.split(":");
                String host = parts[0];
                int port = parts.length > 1 ? Integer.parseInt(parts[1]) : 80;
                int weight = Integer.parseInt(weightStr);

                Version version = new Version(versionName, host, port, weight);
                versionManager.addVersion(version);
                logger.info("Loaded version: {} -> {}:{} (weight: {})",
                        versionName, host, port, weight);
            }
        }
    }

    public void start() throws InterruptedException {
        EventLoopGroup bossGroup = new NioEventLoopGroup(1);
        EventLoopGroup workerGroup = new NioEventLoopGroup();

        try {
            ServerBootstrap b = new ServerBootstrap();
            b.group(bossGroup, workerGroup)
             .channel(NioServerSocketChannel.class)
             .option(ChannelOption.SO_BACKLOG, 128)
             .childOption(ChannelOption.SO_KEEPALIVE, true)
             .childHandler(new ChannelInitializer<SocketChannel>() {
                 @Override
                 protected void initChannel(SocketChannel ch) {
                     ChannelPipeline p = ch.pipeline();
                     p.addLast(new HttpServerCodec());
                     p.addLast(new HttpObjectAggregator(1024 * 1024));
                     p.addLast(new GatewayHandler(versionManager, ruleMatchingEngine,
                             trafficDistributionEngine, statsManager, workerGroup));
                     p.addLast(new ApiHandler(versionManager, ruleManager, statsManager));
                 }
             });

            logger.info("Starting Traffic Tint Gateway on port: {}", port);
            ChannelFuture f = b.bind(port).sync();
            logger.info("Traffic Tint Gateway started successfully");

            f.channel().closeFuture().sync();
        } finally {
            workerGroup.shutdownGracefully();
            bossGroup.shutdownGracefully();
        }
    }

    public static void main(String[] args) throws Exception {
        int port = 8080;
        String configPath = "config/gateway.properties";

        if (args.length > 0) {
            try {
                port = Integer.parseInt(args[0]);
            } catch (NumberFormatException e) {
                logger.error("Invalid port number: {}", args[0]);
                System.exit(1);
            }
        }

        if (args.length > 1) {
            configPath = args[1];
        }

        TrafficTintGateway gateway = new TrafficTintGateway(port);

        try {
            gateway.loadConfig(configPath);
        } catch (IOException e) {
            logger.warn("Could not load config from {}, using empty config: {}", configPath, e.getMessage());
        }

        gateway.start();
    }
}
