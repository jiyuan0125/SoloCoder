package com.configcenter.push;

import com.configcenter.db.ChangeRecord;
import com.configcenter.db.Database;
import com.configcenter.model.Watch;
import com.configcenter.websocket.WebSocketManager;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import io.netty.channel.Channel;
import io.netty.handler.codec.http.websocketx.TextWebSocketFrame;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;

public class ChangeNotifier {

    private static final Logger logger = LoggerFactory.getLogger(ChangeNotifier.class);
    private static final ObjectMapper mapper = new ObjectMapper();
    private static final HttpClient httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private final WebSocketManager webSocketManager;
    private final RetryQueue retryQueue;

    public ChangeNotifier(WebSocketManager webSocketManager, RetryQueue retryQueue) {
        this.webSocketManager = webSocketManager;
        this.retryQueue = retryQueue;
    }

    public void notifyChange(String key, String action, String oldValue, String newValue, boolean isSecret) {
        Database database = new Database();
        List<Watch> watches = database.getAllWatches();

        for (Watch watch : watches) {
            if (watch.getMode().equals("websocket")) {
                notifyWebSocket(watch, key, action, oldValue, newValue, isSecret);
            } else if (watch.getMode().equals("http")) {
                notifyHttp(watch, key, action, oldValue, newValue, isSecret);
            }
        }
    }

    private void notifyWebSocket(Watch watch, String key, String action, String oldValue, String newValue, boolean isSecret) {
        Channel channel = webSocketManager.getChannel(watch.getClientId());

        if (channel == null || !channel.isActive()) {
            logger.info("WebSocket client {} not connected, adding to retry queue", watch.getClientId());
            retryQueue.addRetry(watch.getClientId(), "websocket", null, key, action, oldValue, newValue, isSecret);
            return;
        }

        try {
            String message = createNotificationMessage(key, action, oldValue, newValue, isSecret);
            channel.writeAndFlush(new TextWebSocketFrame(message));
            logger.info("Notified WebSocket client {} about change to {}", watch.getClientId(), key);
        } catch (Exception e) {
            logger.error("Failed to notify WebSocket client {}", watch.getClientId(), e);
            retryQueue.addRetry(watch.getClientId(), "websocket", null, key, action, oldValue, newValue, isSecret);
        }
    }

    private void notifyHttp(Watch watch, String key, String action, String oldValue, String newValue, boolean isSecret) {
        try {
            String message = createNotificationMessage(key, action, oldValue, newValue, isSecret);

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(watch.getCallbackUrl()))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(5))
                    .POST(HttpRequest.BodyPublishers.ofString(message))
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                logger.info("Notified HTTP client {} about change to {}", watch.getClientId(), key);
            } else {
                logger.warn("HTTP callback returned status {} for client {}", response.statusCode(), watch.getClientId());
                retryQueue.addRetry(watch.getClientId(), "http", watch.getCallbackUrl(), key, action, oldValue, newValue, isSecret);
            }
        } catch (Exception e) {
            logger.error("Failed to notify HTTP client {}", watch.getClientId(), e);
            retryQueue.addRetry(watch.getClientId(), "http", watch.getCallbackUrl(), key, action, oldValue, newValue, isSecret);
        }
    }

    private String createNotificationMessage(String key, String action, String oldValue, String newValue, boolean isSecret) throws Exception {
        ObjectNode node = mapper.createObjectNode();
        node.put("key", key);
        node.put("action", action);
        node.put("secret", isSecret);
        if (oldValue != null) {
            node.put("old_value", oldValue);
        }
        if (newValue != null) {
            node.put("new_value", newValue);
        }
        node.put("timestamp", System.currentTimeMillis());
        return node.toString();
    }

    public void resendPendingChanges(Watch watch) {
        Database database = new Database();
        List<ChangeRecord> pendingChanges = database.getPendingChanges(watch.getClientId());

        logger.info("Resending {} pending changes to client {}", pendingChanges.size(), watch.getClientId());

        for (ChangeRecord change : pendingChanges) {
            if (watch.getMode().equals("websocket")) {
                notifyWebSocket(watch, change.getKey(), change.getAction(), change.getOldValue(), change.getNewValue(), change.isSecret());
            } else if (watch.getMode().equals("http")) {
                notifyHttp(watch, change.getKey(), change.getAction(), change.getOldValue(), change.getNewValue(), change.isSecret());
            }
        }

        database.clearPendingChanges(watch.getClientId());
    }
}
