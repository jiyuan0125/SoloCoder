package com.solocoder.mq.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
public class Message {

    @JsonProperty("id")
    private String id;

    @JsonProperty("offset")
    private long offset;

    @JsonProperty("topic")
    private String topic;

    @JsonProperty("content")
    private String content;

    @JsonProperty("status")
    private MessageStatus status;

    @JsonProperty("delay_ms")
    private long delayMs;

    @JsonProperty("created_at")
    private long createdAt;

    @JsonProperty("ready_at")
    private long readyAt;

    @JsonProperty("retry_count")
    private int retryCount;

    @JsonProperty("last_consuming_at")
    private long lastConsumingAt;

    public static Message create(String topic, String content, long delayMs) {
        long now = System.currentTimeMillis();
        Message msg = new Message();
        msg.setId(UUID.randomUUID().toString());
        msg.setTopic(topic);
        msg.setContent(content);
        msg.setDelayMs(delayMs);
        msg.setStatus(delayMs == 0 ? MessageStatus.READY : MessageStatus.PENDING);
        msg.setCreatedAt(now);
        msg.setReadyAt(now + delayMs);
        msg.setRetryCount(0);
        msg.setLastConsumingAt(0);
        return msg;
    }
}
