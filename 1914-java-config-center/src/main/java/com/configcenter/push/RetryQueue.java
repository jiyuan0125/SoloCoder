package com.configcenter.push;

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
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class RetryQueue {

    private static final Logger logger = LoggerFactory.getLogger(RetryQueue.class);
    private static final ObjectMapper mapper = new ObjectMapper();
    private static final HttpClient httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private static final int[] RETRY_INTERVALS = {5, 15, 30};
    private static final int MAX_RETRIES = 3;

    private final BlockingQueue<RetryTask> queue = new LinkedBlockingQueue<>();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
    private final ExecutorService executor = Executors.newCachedThreadPool();

    public RetryQueue() {
        startProcessor();
    }

    private void startProcessor() {
        executor.submit(() -> {
            while (!Thread.currentThread().isInterrupted()) {
                try {
                    RetryTask task = queue.take();
                    processTask(task);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                } catch (Exception e) {
                    logger.error("Error processing retry task", e);
                }
            }
        });
    }

    public void addRetry(String clientId, String mode, String callbackUrl,
                         String key, String action, String oldValue, String newValue, boolean isSecret) {
        RetryTask task = new RetryTask(clientId, mode, callbackUrl, key, action, oldValue, newValue, isSecret);
        queue.offer(task);
        logger.info("Added retry task for client {}, key: {}, attempt: {}", clientId, key, task.getAttemptCount() + 1);
    }

    private void processTask(RetryTask task) {
        if (task.getAttemptCount() >= MAX_RETRIES) {
            logger.warn("Max retries reached for client {}, key: {}. Giving up.",
                    task.getClientId(), task.getKey());
            return;
        }

        boolean success = attemptDelivery(task);

        if (success) {
            logger.info("Successfully delivered after {} attempts for client {}, key: {}",
                    task.getAttemptCount() + 1, task.getClientId(), task.getKey());
        } else {
            task.incrementAttempt();
            if (task.getAttemptCount() < MAX_RETRIES) {
                int delay = RETRY_INTERVALS[task.getAttemptCount() - 1];
                logger.info("Scheduling retry {} for client {}, key: {} in {} seconds",
                        task.getAttemptCount() + 1, task.getClientId(), task.getKey(), delay);
                scheduler.schedule(() -> queue.offer(task), delay, TimeUnit.SECONDS);
            } else {
                logger.warn("Max retries reached for client {}, key: {}. Giving up.",
                        task.getClientId(), task.getKey());
            }
        }
    }

    private boolean attemptDelivery(RetryTask task) {
        try {
            String message = createMessage(task);

            if (task.getMode().equals("websocket")) {
                return deliverWebSocket(task, message);
            } else if (task.getMode().equals("http")) {
                return deliverHttp(task, message);
            }
            return false;
        } catch (Exception e) {
            logger.error("Delivery attempt failed for client {}, key: {}",
                    task.getClientId(), task.getKey(), e);
            return false;
        }
    }

    private boolean deliverWebSocket(RetryTask task, String message) {
        return false;
    }

    private boolean deliverHttp(RetryTask task, String message) {
        try {
            if (task.getCallbackUrl() == null) {
                return false;
            }

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(task.getCallbackUrl()))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(5))
                    .POST(HttpRequest.BodyPublishers.ofString(message))
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            return response.statusCode() >= 200 && response.statusCode() < 300;
        } catch (Exception e) {
            return false;
        }
    }

    private String createMessage(RetryTask task) throws Exception {
        ObjectNode node = mapper.createObjectNode();
        node.put("key", task.getKey());
        node.put("action", task.getAction());
        node.put("secret", task.isSecret());
        if (task.getOldValue() != null) {
            node.put("old_value", task.getOldValue());
        }
        if (task.getNewValue() != null) {
            node.put("new_value", task.getNewValue());
        }
        node.put("retry_attempt", task.getAttemptCount() + 1);
        node.put("timestamp", System.currentTimeMillis());
        return node.toString();
    }

    public void shutdown() {
        executor.shutdown();
        scheduler.shutdown();
        try {
            if (!executor.awaitTermination(5, TimeUnit.SECONDS)) {
                executor.shutdownNow();
            }
            if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            executor.shutdownNow();
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }

    private static class RetryTask {
        private final String clientId;
        private final String mode;
        private final String callbackUrl;
        private final String key;
        private final String action;
        private final String oldValue;
        private final String newValue;
        private final boolean secret;
        private final AtomicInteger attemptCount = new AtomicInteger(0);

        public RetryTask(String clientId, String mode, String callbackUrl,
                         String key, String action, String oldValue, String newValue, boolean secret) {
            this.clientId = clientId;
            this.mode = mode;
            this.callbackUrl = callbackUrl;
            this.key = key;
            this.action = action;
            this.oldValue = oldValue;
            this.newValue = newValue;
            this.secret = secret;
        }

        public String getClientId() { return clientId; }
        public String getMode() { return mode; }
        public String getCallbackUrl() { return callbackUrl; }
        public String getKey() { return key; }
        public String getAction() { return action; }
        public String getOldValue() { return oldValue; }
        public String getNewValue() { return newValue; }
        public boolean isSecret() { return secret; }
        public int getAttemptCount() { return attemptCount.get(); }
        public void incrementAttempt() { attemptCount.incrementAndGet(); }
    }
}
