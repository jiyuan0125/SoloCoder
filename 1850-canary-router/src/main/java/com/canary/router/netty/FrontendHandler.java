package com.canary.router.netty;

import com.canary.router.core.RouterManager;
import com.canary.router.model.Stats;
import com.canary.router.model.Version;
import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.*;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;

public class FrontendHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
    private static final Logger logger = LoggerFactory.getLogger(FrontendHandler.class);
    
    private final ApiHandler apiHandler = new ApiHandler();
    private final RouterManager routerManager = RouterManager.getInstance();
    private final EventLoopGroup workerGroup;
    
    private Channel outboundChannel;
    private FullHttpRequest currentRequest;
    private String currentVersion;
    
    public FrontendHandler(EventLoopGroup workerGroup) {
        this.workerGroup = workerGroup;
    }
    
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
        InetSocketAddress clientAddress = (InetSocketAddress) ctx.channel().remoteAddress();
        
        FullHttpResponse apiResponse = apiHandler.handle(request);
        if (apiResponse != null) {
            writeResponse(ctx, apiResponse);
            return;
        }
        
        Version targetVersion = routerManager.route(request, clientAddress);
        
        if (targetVersion == null) {
            writeResponse(ctx, ApiHandler.errorResponse(
                HttpResponseStatus.SERVICE_UNAVAILABLE, "No backend available"));
            return;
        }
        
        this.currentRequest = request.retain();
        this.currentVersion = targetVersion.getName();
        
        Stats stats = routerManager.getStats(currentVersion);
        stats.incrementRequest();
        
        forwardRequest(ctx, targetVersion);
    }
    
    private void forwardRequest(ChannelHandlerContext ctx, Version version) {
        Bootstrap b = new Bootstrap();
        b.group(workerGroup)
         .channel(NioSocketChannel.class)
         .handler(new ChannelInitializer<SocketChannel>() {
             @Override
             protected void initChannel(SocketChannel ch) {
                 ChannelPipeline p = ch.pipeline();
                 p.addLast(new HttpClientCodec());
                 p.addLast(new HttpObjectAggregator(1024 * 1024));
                 p.addLast(new BackendHandler(ctx.channel()));
             }
         });
        
        ChannelFuture connectFuture = b.connect(version.getHost(), version.getPort());
        this.outboundChannel = connectFuture.channel();
        
        connectFuture.addListener((ChannelFutureListener) future -> {
            if (future.isSuccess()) {
                outboundChannel.writeAndFlush(currentRequest);
            } else {
                logger.error("Failed to connect to backend", future.cause());
                Stats stats = routerManager.getStats(currentVersion);
                stats.incrementError();
                
                ctx.writeAndFlush(ApiHandler.errorResponse(
                    HttpResponseStatus.BAD_GATEWAY, "Backend connection failed"));
                ctx.close();
            }
        });
    }
    
    private void writeResponse(ChannelHandlerContext ctx, FullHttpResponse response) {
        boolean keepAlive = HttpUtil.isKeepAlive(currentRequest != null ? currentRequest : 
            new DefaultFullHttpRequest(HttpVersion.HTTP_1_1, HttpMethod.GET, "/"));
        
        if (!keepAlive) {
            ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
        } else {
            response.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
            ctx.writeAndFlush(response);
        }
    }
    
    class BackendHandler extends SimpleChannelInboundHandler<FullHttpResponse> {
        private final Channel inboundChannel;
        
        BackendHandler(Channel inboundChannel) {
            this.inboundChannel = inboundChannel;
        }
        
        @Override
        protected void channelRead0(ChannelHandlerContext ctx, FullHttpResponse response) {
            if (response.status().code() >= 500) {
                Stats stats = routerManager.getStats(currentVersion);
                stats.incrementError();
            }
            
            inboundChannel.writeAndFlush(response).addListener((ChannelFutureListener) future -> {
                if (future.isSuccess()) {
                    if (outboundChannel != null) {
                        outboundChannel.close();
                    }
                }
            });
        }
        
        @Override
        public void channelInactive(ChannelHandlerContext ctx) {
            if (inboundChannel.isActive()) {
                inboundChannel.close();
            }
        }
        
        @Override
        public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
            logger.error("Backend handler error", cause);
            ctx.close();
            
            if (inboundChannel.isActive()) {
                Stats stats = routerManager.getStats(currentVersion);
                stats.incrementError();
                inboundChannel.writeAndFlush(ApiHandler.errorResponse(
                    HttpResponseStatus.BAD_GATEWAY, "Backend error"));
                inboundChannel.close();
            }
        }
    }
    
    @Override
    public void channelInactive(ChannelHandlerContext ctx) {
        if (outboundChannel != null && outboundChannel.isActive()) {
            outboundChannel.close();
        }
    }
    
    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Frontend handler error", cause);
        ctx.close();
    }
}
