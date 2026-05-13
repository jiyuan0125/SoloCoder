package com.solocoder.mq.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.concurrent.atomic.AtomicLong;

@Data
@NoArgsConstructor
public class ConsumerGroup {

    @JsonProperty("group_id")
    private String groupId;

    @JsonProperty("current_offset")
    private AtomicLong currentOffset = new AtomicLong(-1);

    @JsonProperty("last_consuming_message_id")
    private volatile String lastConsumingMessageId;

    @JsonProperty("last_consuming_at")
    private volatile long lastConsumingAt;

    @JsonProperty("consumer_active")
    private volatile boolean consumerActive;

    @JsonProperty("consumer_id")
    private volatile String consumerId;

    @JsonProperty("last_heartbeat")
    private volatile long lastHeartbeat;

    public static ConsumerGroup create(String groupId) {
        ConsumerGroup group = new ConsumerGroup();
        group.setGroupId(groupId);
        return group;
    }

    public long getCurrentOffsetValue() {
        return currentOffset.get();
    }

    public void setCurrentOffsetValue(long offset) {
        currentOffset.set(offset);
    }

    public long incrementAndGetOffset() {
        return currentOffset.incrementAndGet();
    }
}
