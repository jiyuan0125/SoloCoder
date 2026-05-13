package com.solocoder.mq.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.concurrent.atomic.AtomicLong;

@Data
@NoArgsConstructor
@JsonIgnoreProperties(ignoreUnknown = true)
public class ConsumerGroup {

    @JsonProperty("group_id")
    private String groupId;

    @JsonIgnore
    private AtomicLong currentOffset = new AtomicLong(-1);

    @JsonProperty("current_offset")
    private long currentOffsetValue = -1;

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

    @JsonIgnore
    public long getCurrentOffsetAtomic() {
        return currentOffset.get();
    }

    @JsonIgnore
    public void setCurrentOffsetAtomic(long offset) {
        currentOffset.set(offset);
        currentOffsetValue = offset;
    }

    @JsonIgnore
    public long incrementAndGetOffset() {
        long offset = currentOffset.incrementAndGet();
        currentOffsetValue = offset;
        return offset;
    }

    public long getCurrentOffsetValue() {
        return currentOffset.get();
    }

    public void setCurrentOffsetValue(long offset) {
        currentOffset.set(offset);
        currentOffsetValue = offset;
    }
}
