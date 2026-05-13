package com.configcenter.http;

import com.configcenter.crypto.EncryptionService;
import com.configcenter.db.ChangeRecord;
import com.configcenter.db.Database;
import com.configcenter.model.ConfigEntry;
import com.configcenter.model.Watch;
import com.configcenter.push.ChangeNotifier;
import com.configcenter.push.RetryQueue;
import com.configcenter.websocket.WebSocketManager;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import static io.netty.handler.codec.http.HttpHeaderNames.*;
import static io.netty.handler.codec.http.HttpResponseStatus.*;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;

public class HttpRequestHandler extends SimpleChannelInboundHandler<FullHttpRequest> {

    private static final Logger logger = LoggerFactory.getLogger(HttpRequestHandler.class);
    private static final int MAX_VALUE_SIZE = 64 * 1024;
    private static final ObjectMapper mapper = new ObjectMapper();

    private final EncryptionService encryptionService;
    private final Database database;
    private final WebSocketManager webSocketManager;
    private final ChangeNotifier changeNotifier;
    private final RetryQueue retryQueue;

    public HttpRequestHandler(EncryptionService encryptionService,
                              Database database,
                              WebSocketManager webSocketManager,
                              ChangeNotifier changeNotifier,
                              RetryQueue retryQueue) {
        this.encryptionService = encryptionService;
        this.database = database;
        this.webSocketManager = webSocketManager;
        this.changeNotifier = changeNotifier;
        this.retryQueue = retryQueue;
    }

    @Override
    public void channelRead0(ChannelHandlerContext ctx, FullHttpRequest req) throws Exception {
        if (!req.decoderResult().isSuccess()) {
            sendResponse(ctx, BAD_REQUEST, createErrorResponse("Bad request"));
            return;
        }

        String uri = req.uri();
        HttpMethod method = req.method();

        logger.info("Request: {} {}", method, uri);

        try {
            if (uri.equals("/configs/export") && method == HttpMethod.GET) {
                handleExportConfigs(ctx);
            } else if (uri.equals("/configs") && method == HttpMethod.GET) {
                handleListConfigs(ctx);
            } else if (uri.startsWith("/configs/") && method == HttpMethod.GET) {
                handleGetConfig(ctx, req);
            } else if (uri.startsWith("/configs/") && (method == HttpMethod.POST || method == HttpMethod.PUT)) {
                handleSaveConfig(ctx, req);
            } else if (uri.startsWith("/configs/") && method == HttpMethod.DELETE) {
                handleDeleteConfig(ctx, req);
            } else if (uri.equals("/watches") && method == HttpMethod.POST) {
                handleRegisterWatch(ctx, req);
            } else if (uri.equals("/ws")) {
                ctx.fireChannelRead(req.retain());
            } else {
                sendResponse(ctx, NOT_FOUND, createErrorResponse("Not found"));
            }
        } catch (Exception e) {
            logger.error("Error handling request", e);
            sendResponse(ctx, INTERNAL_SERVER_ERROR, createErrorResponse("Internal server error"));
        }
    }

    private void handleGetConfig(ChannelHandlerContext ctx, FullHttpRequest req) throws Exception {
        String key = extractKey(req.uri());
        Optional<ConfigEntry> entryOpt = database.getConfig(key);

        if (entryOpt.isPresent()) {
            ConfigEntry entry = entryOpt.get();
            ObjectNode response = mapper.createObjectNode();
            response.put("key", entry.getKey());

            if (entry.isSecret()) {
                response.put("value", encryptionService.decrypt(entry.getValue()));
            } else {
                response.put("value", entry.getValue());
            }
            response.put("secret", entry.isSecret());

            sendResponse(ctx, OK, response.toString());
        } else {
            sendResponse(ctx, NOT_FOUND, createErrorResponse("Config not found"));
        }
    }

    private void handleSaveConfig(ChannelHandlerContext ctx, FullHttpRequest req) throws Exception {
        String key = extractKey(req.uri());
        String body = req.content().toString(CharsetUtil.UTF_8);

        JsonNode json = mapper.readTree(body);
        String value = json.has("value") ? json.get("value").asText() : null;
        boolean secret = json.has("secret") && json.get("secret").asBoolean();

        if (value == null) {
            sendResponse(ctx, BAD_REQUEST, createErrorResponse("Value is required"));
            return;
        }

        if (value.getBytes(CharsetUtil.UTF_8).length > MAX_VALUE_SIZE) {
            sendResponse(ctx, UNPROCESSABLE_ENTITY, createErrorResponse("Value exceeds 64KB limit"));
            return;
        }

        Optional<ConfigEntry> existingOpt = database.getConfig(key);
        String action = existingOpt.isPresent() ? "updated" : "created";
        String oldValue = existingOpt.map(ConfigEntry::getValue).orElse(null);
        boolean oldSecret = existingOpt.map(ConfigEntry::isSecret).orElse(false);

        String storedValue;
        if (secret) {
            storedValue = encryptionService.encrypt(value);
        } else {
            storedValue = value;
        }

        if (existingOpt.isPresent()) {
            ConfigEntry existing = existingOpt.get();
            if (oldSecret && !secret) {
                storedValue = encryptionService.decrypt(existing.getValue());
                if (json.has("value")) {
                    storedValue = value;
                }
            } else if (!oldSecret && secret) {
                if (json.has("value")) {
                    storedValue = encryptionService.encrypt(value);
                } else {
                    storedValue = encryptionService.encrypt(existing.getValue());
                }
            }
        }

        ConfigEntry entry = new ConfigEntry(key, storedValue, secret,
                existingOpt.map(ConfigEntry::getCreatedAt).orElse(Instant.now()),
                Instant.now());

        database.saveConfig(entry);

        String newLogValue = secret ? "[ENCRYPTED]" : storedValue;
        String oldLogValue = oldSecret ? "[ENCRYPTED]" : oldValue;
        database.logChange(key, action, oldLogValue, newLogValue, secret);

        changeNotifier.notifyChange(key, action, oldValue, storedValue, secret);

        ObjectNode response = mapper.createObjectNode();
        response.put("key", key);
        response.put("value", value);
        response.put("secret", secret);
        response.put("action", action);

        sendResponse(ctx, existingOpt.isPresent() ? OK : CREATED, response.toString());
    }

    private void handleDeleteConfig(ChannelHandlerContext ctx, FullHttpRequest req) throws Exception {
        String key = extractKey(req.uri());
        Optional<ConfigEntry> existingOpt = database.getConfig(key);

        if (existingOpt.isPresent()) {
            ConfigEntry existing = existingOpt.get();
            database.deleteConfig(key);

            String oldLogValue = existing.isSecret() ? "[ENCRYPTED]" : existing.getValue();
            database.logChange(key, "deleted", oldLogValue, null, existing.isSecret());

            changeNotifier.notifyChange(key, "deleted", existing.getValue(), null, existing.isSecret());

            ObjectNode response = mapper.createObjectNode();
            response.put("key", key);
            response.put("action", "deleted");
            sendResponse(ctx, OK, response.toString());
        } else {
            sendResponse(ctx, NOT_FOUND, createErrorResponse("Config not found"));
        }
    }

    private void handleListConfigs(ChannelHandlerContext ctx) throws Exception {
        List<ConfigEntry> configs = database.getAllConfigs();
        ArrayNode array = mapper.createArrayNode();

        for (ConfigEntry entry : configs) {
            ObjectNode node = mapper.createObjectNode();
            node.put("key", entry.getKey());
            if (entry.isSecret()) {
                node.put("value", encryptionService.decrypt(entry.getValue()));
            } else {
                node.put("value", entry.getValue());
            }
            node.put("secret", entry.isSecret());
            array.add(node);
        }

        sendResponse(ctx, OK, array.toString());
    }

    private void handleExportConfigs(ChannelHandlerContext ctx) throws Exception {
        List<ConfigEntry> configs = database.getAllConfigs();
        List<String> skippedKeys = new ArrayList<>();

        ObjectNode response = mapper.createObjectNode();
        ArrayNode configsArray = mapper.createArrayNode();
        ArrayNode skippedArray = mapper.createArrayNode();

        for (ConfigEntry entry : configs) {
            if (entry.isSecret()) {
                skippedKeys.add(entry.getKey());
                skippedArray.add(entry.getKey());
            } else {
                ObjectNode node = mapper.createObjectNode();
                node.put("key", entry.getKey());
                node.put("value", entry.getValue());
                configsArray.add(node);
            }
        }

        response.set("configs", configsArray);
        response.set("skipped_keys", skippedArray);
        response.put("skipped_count", skippedKeys.size());

        sendResponse(ctx, OK, response.toString());
    }

    private void handleRegisterWatch(ChannelHandlerContext ctx, FullHttpRequest req) throws Exception {
        String body = req.content().toString(CharsetUtil.UTF_8);
        JsonNode json = mapper.readTree(body);

        String clientId = json.has("client_id") ? json.get("client_id").asText() : null;
        String mode = json.has("mode") ? json.get("mode").asText() : null;
        String callbackUrl = json.has("callback_url") ? json.get("callback_url").asText() : null;

        if (clientId == null || mode == null) {
            sendResponse(ctx, BAD_REQUEST, createErrorResponse("client_id and mode are required"));
            return;
        }

        if (!mode.equals("websocket") && !mode.equals("http")) {
            sendResponse(ctx, BAD_REQUEST, createErrorResponse("mode must be 'websocket' or 'http'"));
            return;
        }

        if (mode.equals("http") && callbackUrl == null) {
            sendResponse(ctx, BAD_REQUEST, createErrorResponse("callback_url is required for http mode"));
            return;
        }

        Watch watch = new Watch(clientId, mode, callbackUrl);
        database.saveWatch(watch);

        List<ChangeRecord> pendingChanges = database.getPendingChanges(clientId);
        if (!pendingChanges.isEmpty()) {
            logger.info("Client {} reconnected, {} pending changes to send", clientId, pendingChanges.size());
            for (ChangeRecord change : pendingChanges) {
                retryQueue.addRetry(clientId, mode, callbackUrl,
                        change.getKey(), change.getAction(),
                        change.getOldValue(), change.getNewValue(),
                        change.isSecret());
            }
        }

        ObjectNode response = mapper.createObjectNode();
        response.put("client_id", clientId);
        response.put("mode", mode);
        if (callbackUrl != null) {
            response.put("callback_url", callbackUrl);
        }
        response.put("pending_changes", pendingChanges.size());

        sendResponse(ctx, OK, response.toString());
    }

    private String extractKey(String uri) {
        int idx = uri.indexOf('?');
        String path = idx > 0 ? uri.substring(0, idx) : uri;
        return path.substring("/configs/".length());
    }

    private String createErrorResponse(String message) throws Exception {
        ObjectNode node = mapper.createObjectNode();
        node.put("error", message);
        return node.toString();
    }

    private void sendResponse(ChannelHandlerContext ctx, HttpResponseStatus status, String content) {
        ByteBuf buffer = Unpooled.copiedBuffer(content, CharsetUtil.UTF_8);
        FullHttpResponse response = new DefaultFullHttpResponse(HTTP_1_1, status, buffer);
        response.headers().set(CONTENT_TYPE, "application/json; charset=UTF-8");
        response.headers().set(CONTENT_LENGTH, response.content().readableBytes());
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Exception caught", cause);
        ctx.close();
    }
}
