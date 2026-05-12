package com.example.memorymq.engine;

import com.example.memorymq.model.Message;

import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.Delayed;
import java.util.concurrent.TimeUnit;

public class InFlightMessage implements Delayed {
    private final String messageId;
    private final Message message;
    private final String consumerId;
    private final int retryCount;
    private final long availableAtNanos;
    private final Instant createdAt;

    public InFlightMessage(String messageId, Message message, String consumerId,
                       int retryCount, long delayMillis) {
        this.messageId = messageId;
        this.message = message;
        this.consumerId = consumerId;
        this.retryCount = retryCount;
        this.createdAt = Instant.now();
        this.availableAtNanos = System.nanoTime() + TimeUnit.MILLISECONDS.toNanos(delayMillis);
    }

    @Override
    public long getDelay(TimeUnit unit) {
        return unit.convert(availableAtNanos - System.nanoTime(), TimeUnit.NANOSECONDS);
    }

    @Override
    public int compareTo(Delayed other) {
        if (this == other) return 0;
        long diff = this.getDelay(TimeUnit.NANOSECONDS) - other.getDelay(TimeUnit.NANOSECONDS);
        return (diff < 0) ? -1 : (diff > 0) ? 1 : 0;
    }

    public String getMessageId() { return messageId; }

    public Message getMessage() { return message; }

    public String getConsumerId() { return consumerId; }

    public int getRetryCount() { return retryCount; }

    public Instant getCreatedAt() { return createdAt; }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof InFlightMessage)) return false;
        InFlightMessage that = (InFlightMessage) o;
        return messageId.equals(that.messageId);
    }

    @Override
    public int hashCode() {
        return messageId.hashCode();
    }
}
