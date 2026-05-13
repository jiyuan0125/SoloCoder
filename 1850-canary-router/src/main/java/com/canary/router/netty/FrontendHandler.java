package com.canary.router.netty;

import com.canary.router.core.RouterManager;
import com.canary.router.model.Stats;
import com.canary.router.model.Version;
import io.netty.bootstrap.Bootstrap;
import io.netty.channel.*;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.*;
import io.netty.util.ReferenceCountUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;
import java.util.concurrent.atomic.AtomicBoolean;

public class FrontendHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
    private static final Logger logger = LoggerFactory.getLogger(FrontendHandler.class);
    
    private final ApiHandler apiHandler = new ApiHandler();
    private final RouterManager routerManager = RouterManager.getInstance();
    private final EventLoopGroup workerGroup;
    
    public FrontendHandler(EventLoopGroup workerGroup) {
        this.workerGroup = workerGroup;
    }
    
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) {
        InetSocketAddress clientAddress = (InetSocketAddress) ctx.channel().remoteAddress();
        
        FullHttpResponse apiResponse = apiHandler.handle(request);
        if (apiResponse != null) {
            writeResponse(ctx, request, apiResponse);
            return;
        }
        
        Version targetVersion = routerManager.route(request, clientAddress);
        
        if (targetVersion == null) {
            writeResponse(ctx, request, ApiHandler.errorResponse(
                HttpResponseStatus.SERVICE_UNAVAILABLE, "No backend available"));
            return;
        }
        
        final String versionName = targetVersion.getName();
        Stats stats = routerManager.getStats(versionName);
        stats.incrementRequest();
        
        final FullHttpRequest retainedRequest = request.retain();
        final Channel inboundChannel = ctx.channel();
        final AtomicBoolean requestReleased = new AtomicBoolean(false);
        
        Bootstrap b = new Bootstrap();
        b.group(workerGroup)
         .channel(NioSocketChannel.class)
         .handler(new ChannelInitializer<SocketChannel>() {
             @Override
             protected void initChannel(SocketChannel ch) {
                 ChannelPipeline p = ch.pipeline();
                 p.addLast(new HttpClientCodec());
                 p.addLast(new HttpObjectAggregator(1024 * 1024));
                 p.addLast(new BackendHandler(inboundChannel, versionName, retainedRequest, requestReleased));
             }
         });
        
        ChannelFuture connectFuture = b.connect(targetVersion.getHost(), targetVersion.getPort());
        final Channel outboundChannel = connectFuture.channel();
        
        connectFuture.addListener((ChannelFutureListener) future -> {
            if (future.isSuccess()) {
                outboundChannel.writeAndFlush(retainedRequest);
            } else {
                logger.error("Failed to connect to backend", future.cause());
                if (requestReleased.compareAndSet(false, true)) {
                    ReferenceCountUtil.release(retainedRequest);
                }
                
                Stats s = routerManager.getStats(versionName);
                s.incrementError();
                
                inboundChannel.writeAndFlush(ApiHandler.errorResponse(
                    HttpResponseStatus.BAD_GATEWAY, "Backend connection failed"));
                inboundChannel.close();
            }
        });
        
        ctx.channel().closeFuture().addListener((ChannelFutureListener) f -> {
            if (outboundChannel.isActive()) {
                outboundChannel.close();
            }
        });
    }
    
    private void writeResponse(ChannelHandlerContext ctx, FullHttpRequest request, FullHttpResponse response) {
        boolean keepAlive = HttpUtil.isKeepAlive(request);
        
        if (!keepAlive) {
            ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
        } else {
            response.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
            ctx.writeAndFlush(response);
        }
    }
    
    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Frontend handler error", cause);
        ctx.close();
    }
    
    static class BackendHandler extends ChannelInboundHandlerAdapter {
        private final Channel inboundChannel;
        private final String versionName;
        private final FullHttpRequest originalRequest;
        private final AtomicBoolean requestReleased;
        
        BackendHandler(Channel inboundChannel, String versionName, 
                      FullHttpRequest originalRequest, AtomicBoolean requestReleased) {
            this.inboundChannel = inboundChannel;
            this.versionName = versionName;
            this.originalRequest = originalRequest;
            this.requestReleased = requestReleased;
        }
        
        @Override
        public void channelRead(ChannelHandlerContext ctx, Object msg) {
            try {
                if (msg instanceof FullHttpResponse) {
                    FullHttpResponse response = (FullHttpResponse) msg;
                    
                    if (response.status().code() >= 500) {
                        Stats stats = RouterManager.getInstance().getStats(versionName);
                        stats.incrementError();
                    }
                    
                    final Channel outboundChannel = ctx.channel();
                    final boolean responseKeepAlive = HttpUtil.isKeepAlive(response);
                    final boolean requestKeepAlive = HttpUtil.isKeepAlive(originalRequest);
                    
                    FullHttpResponse retainedResponse = response.retain();
                    
                    inboundChannel.writeAndFlush(retainedResponse).addListener((ChannelFutureListener) future -> {
                        if (!future.isSuccess()) {
                            logger.error("Failed to write response to client", future.cause());
                            future.channel().close();
                        }
                        
                        if (!responseKeepAlive) {
                            outboundChannel.close();
                        }
                        if (!requestKeepAlive) {
                            inboundChannel.close();
                        }
                    });
                    
                    if (requestReleased.compareAndSet(false, true)) {
                        ReferenceCountUtil.release(originalRequest);
                    }
                } else {
                    ReferenceCountUtil.release(msg);
                }
            } catch (Exception e) {
                logger.error("Error processing backend response", e);
                ctx.close();
                if (inboundChannel.isActive()) {
                    inboundChannel.close();
                }
            }
        }
        
        @Override
        public void channelInactive(ChannelHandlerContext ctx) {
            if (requestReleased.compareAndSet(false, true)) {
                ReferenceCountUtil.release(originalRequest);
            }
            if (inboundChannel.isActive()) {
                inboundChannel.close();
            }
        }
        
        @Override
        public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
            logger.error("Backend handler error", cause);
            ctx.close();
            
            if (requestReleased.compareAndSet(false, true)) {
                ReferenceCountUtil.release(originalRequest);
            }
            
            if (inboundChannel.isActive()) {
                Stats stats = RouterManager.getInstance().getStats(versionName);
                stats.incrementError();
                inboundChannel.writeAndFlush(ApiHandler.errorResponse(
                    HttpResponseStatus.BAD_GATEWAY, "Backend error"));
                inboundChannel.close();
            }
        }
    }
}
