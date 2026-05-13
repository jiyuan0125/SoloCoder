package com.configcenter.websocket;

import com.configcenter.crypto.EncryptionService;
import com.configcenter.db.ChangeRecord;
import com.configcenter.db.Database;
import com.configcenter.model.Watch;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.websocketx.TextWebSocketFrame;
import io.netty.handler.codec.http.websocketx.WebSocketServerProtocolHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.Optional;

public class WebSocketServerHandler extends SimpleChannelInboundHandler<TextWebSocketFrame> {

    private static final Logger logger = LoggerFactory.getLogger(WebSocketServerHandler.class);
    private static final ObjectMapper mapper = new ObjectMapper();

    private final WebSocketManager webSocketManager;
    private final EncryptionService encryptionService;
    private final Database database;

    public WebSocketServerHandler(WebSocketManager webSocketManager,
                                  EncryptionService encryptionService,
                                  Database database) {
        this.webSocketManager = webSocketManager;
        this.encryptionService = encryptionService;
        this.database = database;
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof WebSocketServerProtocolHandler.HandshakeComplete) {
            logger.info("WebSocket handshake complete");
        }
        super.userEventTriggered(ctx, evt);
    }

    @Override
    public void channelRead0(ChannelHandlerContext ctx, TextWebSocketFrame frame) throws Exception {
        String message = frame.text();
        logger.info("Received WebSocket message: {}", message);

        try {
            JsonNode json = mapper.readTree(message);
            String type = json.has("type") ? json.get("type").asText() : null;

            if ("register".equals(type)) {
                handleRegister(ctx, json);
            } else if ("ping".equals(type)) {
                sendPong(ctx);
            } else {
                sendError(ctx, "Unknown message type");
            }
        } catch (Exception e) {
            logger.error("Error handling WebSocket message", e);
            sendError(ctx, "Invalid message format");
        }
    }

    private void handleRegister(ChannelHandlerContext ctx, JsonNode json) throws Exception {
        String clientId = json.has("client_id") ? json.get("client_id").asText() : null;

        if (clientId == null) {
            sendError(ctx, "client_id is required");
            return;
        }

        webSocketManager.registerChannel(clientId, ctx.channel());

        Optional<Watch> watchOpt = database.getWatch(clientId);
        if (watchOpt.isPresent()) {
            List<ChangeRecord> pendingChanges = database.getPendingChanges(clientId);
            if (!pendingChanges.isEmpty()) {
                logger.info("Client {} reconnected, sending {} pending changes", clientId, pendingChanges.size());

                for (ChangeRecord change : pendingChanges) {
                    ObjectNode notification = mapper.createObjectNode();
                    notification.put("type", "change");
                    notification.put("key", change.getKey());
                    notification.put("action", change.getAction());
                    notification.put("secret", change.isSecret());

                    if (change.getOldValue() != null) {
                        notification.put("old_value", change.isSecret() ? "[ENCRYPTED]" : change.getOldValue());
                    }
                    if (change.getNewValue() != null) {
                        notification.put("new_value", change.isSecret() ? "[ENCRYPTED]" : change.getNewValue());
                    }
                    notification.put("timestamp", change.getTimestamp().toEpochMilli());

                    ctx.channel().writeAndFlush(new TextWebSocketFrame(notification.toString()));
                }

                database.clearPendingChanges(clientId);
            }
        }

        ObjectNode response = mapper.createObjectNode();
        response.put("type", "registered");
        response.put("client_id", clientId);
        ctx.channel().writeAndFlush(new TextWebSocketFrame(response.toString()));
    }

    private void sendPong(ChannelHandlerContext ctx) throws Exception {
        ObjectNode response = mapper.createObjectNode();
        response.put("type", "pong");
        response.put("timestamp", System.currentTimeMillis());
        ctx.channel().writeAndFlush(new TextWebSocketFrame(response.toString()));
    }

    private void sendError(ChannelHandlerContext ctx, String error) throws Exception {
        ObjectNode response = mapper.createObjectNode();
        response.put("type", "error");
        response.put("message", error);
        ctx.channel().writeAndFlush(new TextWebSocketFrame(response.toString()));
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        webSocketManager.removeChannel(ctx.channel());
        super.channelInactive(ctx);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("WebSocket exception", cause);
        ctx.close();
    }
}
