package com.trace.collector.service;

import com.trace.collector.model.Span;
import com.trace.collector.model.ServiceRegistration;
import com.trace.collector.model.TraceNode;
import com.trace.collector.store.RetryQueue;
import com.trace.collector.store.ServiceRegistry;
import com.trace.collector.store.SpanStore;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

public class SpanService {

    private static final Logger logger = LoggerFactory.getLogger(SpanService.class);

    private final ServiceRegistry serviceRegistry = new ServiceRegistry();
    private final SpanStore spanStore = new SpanStore();
    private final RetryQueue retryQueue = new RetryQueue();

    private final AtomicBoolean running = new AtomicBoolean(false);
    private Thread retryThread;

    public void start() {
        if (running.compareAndSet(false, true)) {
            retryThread = new Thread(this::runRetryLoop, "retry-thread");
            retryThread.setDaemon(true);
            retryThread.start();
            logger.info("SpanService started");
        }
    }

    public void stop() {
        running.compareAndSet(true, false);
        logger.info("SpanService stopped");
    }

    public void acceptSpan(Span span) {
        if (span.getTraceId() == null || span.getTraceId().isEmpty()) {
            logger.warn("Rejected span: missing traceId");
            return;
        }

        if (span.getServiceName() == null || span.getServiceName().isEmpty()) {
            logger.warn("Rejected span: missing serviceName, traceId={}", span.getTraceId());
            return;
        }

        if (!serviceRegistry.isRegistered(span.getServiceName())) {
            logger.warn("Discarding span from unregistered service: serviceName={}, traceId={}",
                    span.getServiceName(), span.getTraceId());
            return;
        }

        storeSpan(span);
    }

    private void storeSpan(Span span) {
        spanStore.addSpan(span);
        logger.debug("Stored span: traceId={}, serviceName={}, spanId={}",
                span.getTraceId(), span.getServiceName(), span.getSpanId());
    }

    private void runRetryLoop() {
        while (running.get()) {
            try {
                Thread.sleep(RetryQueue.getRetryIntervalMs());
                processRetryQueue();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            } catch (Exception e) {
                logger.error("Error in retry loop", e);
            }
        }
    }

    private void processRetryQueue() {
        if (retryQueue.isEmpty()) {
            return;
        }
        logger.info("Processing retry queue, size={}", retryQueue.size());
    }

    public List<TraceNode> getTraceTree(String traceId) {
        return spanStore.getTraceTree(traceId);
    }

    public List<Span> searchSpans(String serviceName, Long from, Long to) {
        return spanStore.searchSpans(serviceName, from, to);
    }

    public void registerService(ServiceRegistration registration) {
        serviceRegistry.register(registration);
    }

    public RetryQueue getRetryQueue() {
        return retryQueue;
    }
}
