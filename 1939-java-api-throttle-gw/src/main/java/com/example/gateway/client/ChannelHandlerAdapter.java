package com.example.gateway.client;

import io.netty.channel.Channel;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.timeout.ReadTimeoutException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public abstract class ChannelHandlerAdapter extends SimpleChannelInboundHandler<FullHttpResponse> {
    private static final Logger logger = LoggerFactory.getLogger(ChannelHandlerAdapter.class);

    private Channel clientChannel;
    private FullHttpRequest forwardRequest;
    private boolean hasResponse = false;

    public void setClientChannel(Channel clientChannel) {
        this.clientChannel = clientChannel;
    }

    public Channel getClientChannel() {
        return clientChannel;
    }

    public void setForwardRequest(FullHttpRequest forwardRequest) {
        this.forwardRequest = forwardRequest;
    }

    public boolean hasResponse() {
        return hasResponse;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        if (forwardRequest != null && clientChannel != null && clientChannel.isActive()) {
            logger.debug("Channel active, sending forward request");
            sendRequest();
        }
        super.channelActive(ctx);
    }

    public void onChannelReady() {
        if (forwardRequest != null && clientChannel != null && clientChannel.isActive()) {
            logger.debug("Channel ready, sending forward request");
            sendRequest();
        }
    }

    private void sendRequest() {
        if (forwardRequest == null) {
            logger.warn("No forward request to send");
            return;
        }

        ChannelFuture future = clientChannel.writeAndFlush(forwardRequest);
        future.addListener((ChannelFutureListener) f -> {
            if (!f.isSuccess()) {
                logger.error("Failed to send request", f.cause());
                onError(f.cause());
            }
        });
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpResponse response) {
        hasResponse = true;
        onResponse(response);
        ctx.pipeline().remove(this);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Client channel error", cause);
        if (cause instanceof ReadTimeoutException) {
            onTimeout();
        } else {
            onError(cause);
        }
        ctx.close();
    }

    public void onConnectionFailure(Throwable cause) {
        logger.error("Connection failed", cause);
        onError(cause);
    }

    public abstract void onResponse(FullHttpResponse response);

    public abstract void onTimeout();

    public abstract void onError(Throwable cause);
}
