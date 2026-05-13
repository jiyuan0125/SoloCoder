package com.trace.collector.store;

import com.trace.collector.model.Span;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.ReentrantLock;

public class RetryQueue {

    private static final Logger logger = LoggerFactory.getLogger(RetryQueue.class);

    private static final int MAX_SIZE = 1000;
    private static final int RETRY_INTERVAL_MS = 5000;

    private final Queue<Span> queue = new ConcurrentLinkedQueue<>();
    private final AtomicInteger size = new AtomicInteger(0);
    private final ReentrantLock offerLock = new ReentrantLock();

    public boolean offer(Span span) {
        offerLock.lock();
        try {
            int currentSize = size.get();
            if (currentSize >= MAX_SIZE) {
                Span oldest = queue.poll();
                if (oldest != null) {
                    size.decrementAndGet();
                    logger.warn("Retry queue is full, discarding oldest span. TraceID: {}",
                            oldest.getTraceId());
                }
            }
            boolean offered = queue.offer(span);
            if (offered) {
                size.incrementAndGet();
            }
            return offered;
        } finally {
            offerLock.unlock();
        }
    }

    public Span poll() {
        Span span = queue.poll();
        if (span != null) {
            size.decrementAndGet();
        }
        return span;
    }

    public boolean isEmpty() {
        return queue.isEmpty();
    }

    public int size() {
        return size.get();
    }

    public static int getMaxSize() {
        return MAX_SIZE;
    }

    public static int getRetryIntervalMs() {
        return RETRY_INTERVAL_MS;
    }
}
