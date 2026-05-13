package com.configcenter.http;

import com.configcenter.crypto.EncryptionService;
import com.configcenter.db.Database;
import com.configcenter.push.ChangeNotifier;
import com.configcenter.push.RetryQueue;
import com.configcenter.websocket.WebSocketManager;
import com.configcenter.websocket.WebSocketServerHandler;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.socket.SocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.codec.http.websocketx.WebSocketServerProtocolHandler;
import io.netty.handler.ssl.SslContext;

public class HttpServerInitializer extends ChannelInitializer<SocketChannel> {

    private final SslContext sslCtx;
    private final EncryptionService encryptionService;
    private final Database database;
    private final WebSocketManager webSocketManager;
    private final ChangeNotifier changeNotifier;
    private final RetryQueue retryQueue;

    public HttpServerInitializer(EncryptionService encryptionService,
                                 Database database,
                                 WebSocketManager webSocketManager,
                                 ChangeNotifier changeNotifier,
                                 RetryQueue retryQueue) {
        this(null, encryptionService, database, webSocketManager, changeNotifier, retryQueue);
    }

    public HttpServerInitializer(SslContext sslCtx,
                                 EncryptionService encryptionService,
                                 Database database,
                                 WebSocketManager webSocketManager,
                                 ChangeNotifier changeNotifier,
                                 RetryQueue retryQueue) {
        this.sslCtx = sslCtx;
        this.encryptionService = encryptionService;
        this.database = database;
        this.webSocketManager = webSocketManager;
        this.changeNotifier = changeNotifier;
        this.retryQueue = retryQueue;
    }

    @Override
    public void initChannel(SocketChannel ch) {
        ChannelPipeline pipeline = ch.pipeline();

        if (sslCtx != null) {
            pipeline.addLast(sslCtx.newHandler(ch.alloc()));
        }

        pipeline.addLast(new HttpServerCodec());
        pipeline.addLast(new HttpObjectAggregator(65536));
        pipeline.addLast(new HttpRequestHandler(
                encryptionService,
                database,
                webSocketManager,
                changeNotifier,
                retryQueue
        ));
        pipeline.addLast(new WebSocketServerProtocolHandler("/ws"));
        pipeline.addLast(new WebSocketServerHandler(
                webSocketManager,
                encryptionService,
                database
        ));
    }
}
