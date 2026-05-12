package com.canary.router.netty;

import com.canary.router.config.ServerConfig;
import com.canary.router.core.RouterManager;
import com.canary.router.model.Version;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.stream.ChunkedWriteHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CanaryServer {
    private static final Logger logger = LoggerFactory.getLogger(CanaryServer.class);
    
    private final ServerConfig config;
    private final int port;
    
    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;
    private Channel serverChannel;
    
    public CanaryServer(ServerConfig config) {
        this.config = config;
        this.port = config.getPort();
    }
    
    public void start() throws InterruptedException {
        RouterManager routerManager = RouterManager.getInstance();
        
        Version stable = new Version("stable", config.getStableHost(), config.getStablePort(), 90);
        Version canary = new Version("canary", config.getCanaryHost(), config.getCanaryPort(), 10);
        routerManager.addVersion(stable);
        routerManager.addVersion(canary);
        
        bossGroup = new NioEventLoopGroup(1);
        workerGroup = new NioEventLoopGroup();
        
        try {
            ServerBootstrap b = new ServerBootstrap();
            b.group(bossGroup, workerGroup)
             .channel(NioServerSocketChannel.class)
             .childHandler(new ChannelInitializer<SocketChannel>() {
                 @Override
                 protected void initChannel(SocketChannel ch) {
                     ChannelPipeline p = ch.pipeline();
                     p.addLast(new HttpServerCodec());
                     p.addLast(new HttpObjectAggregator(1024 * 1024));
                     p.addLast(new ChunkedWriteHandler());
                     p.addLast(new FrontendHandler(workerGroup));
                 }
             })
             .option(ChannelOption.SO_BACKLOG, 1024)
             .childOption(ChannelOption.SO_KEEPALIVE, true);
            
            ChannelFuture f = b.bind(port).sync();
            serverChannel = f.channel();
            
            logger.info("Canary Router started on port {}", port);
            logger.info("Stable backend: {}:{}", config.getStableHost(), config.getStablePort());
            logger.info("Canary backend: {}:{}", config.getCanaryHost(), config.getCanaryPort());
            
            serverChannel.closeFuture().sync();
        } finally {
            stop();
        }
    }
    
    public void stop() {
        if (serverChannel != null) {
            serverChannel.close();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully();
        }
        if (bossGroup != null) {
            bossGroup.shutdownGracefully();
        }
        logger.info("Canary Router stopped");
    }
}
