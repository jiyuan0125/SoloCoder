package com.configcenter;

import com.configcenter.config.AppConfig;
import com.configcenter.crypto.EncryptionService;
import com.configcenter.db.Database;
import com.configcenter.http.HttpServerInitializer;
import com.configcenter.push.ChangeNotifier;
import com.configcenter.push.RetryQueue;
import com.configcenter.websocket.WebSocketManager;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelOption;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ConfigCenterServer {

    private static final Logger logger = LoggerFactory.getLogger(ConfigCenterServer.class);

    private final AppConfig config;
    private final Database database;
    private final EncryptionService encryptionService;
    private final WebSocketManager webSocketManager;
    private final ChangeNotifier changeNotifier;
    private final RetryQueue retryQueue;

    public ConfigCenterServer(AppConfig config) {
        this.config = config;
        this.encryptionService = new EncryptionService(config.getEncryptionKey());
        this.database = new Database();
        this.database.init();
        this.webSocketManager = new WebSocketManager();
        this.retryQueue = new RetryQueue();
        this.changeNotifier = new ChangeNotifier(webSocketManager, retryQueue);
    }

    public void start() throws Exception {
        logger.info("Starting Config Center on port: {}", config.getPort());

        EventLoopGroup bossGroup = new NioEventLoopGroup(1);
        EventLoopGroup workerGroup = new NioEventLoopGroup();

        try {
            ServerBootstrap bootstrap = new ServerBootstrap();
            bootstrap.group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class)
                    .childHandler(new HttpServerInitializer(
                            encryptionService,
                            database,
                            webSocketManager,
                            changeNotifier,
                            retryQueue
                    ))
                    .option(ChannelOption.SO_BACKLOG, 128)
                    .childOption(ChannelOption.SO_KEEPALIVE, true);

            ChannelFuture future = bootstrap.bind(config.getPort()).sync();
            logger.info("Config Center started successfully on port: {}", config.getPort());

            future.channel().closeFuture().sync();
        } finally {
            bossGroup.shutdownGracefully();
            workerGroup.shutdownGracefully();
        }
    }

    public static void main(String[] args) throws Exception {
        AppConfig config = AppConfig.parse(args);
        new ConfigCenterServer(config).start();
    }
}
