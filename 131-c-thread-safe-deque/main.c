#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include "log_channel.h"
#include "log_collector.h"

#define NUM_WORKER_THREADS  5
#define MESSAGES_PER_THREAD  20
#define MAX_TOTAL_MESSAGES   1000
#define MAX_THREAD_MESSAGES  200

typedef struct {
    log_channel_t*     channel;
    int                thread_id;
    volatile int*      running;
    unsigned int       rand_seed;
} worker_context_t;

static volatile sig_atomic_t g_shutdown_requested = 0;

static void signal_handler(int sig) {
    (void)sig;
    g_shutdown_requested = 1;
}

static void* worker_thread_func(void* arg) {
    worker_context_t* ctx = (worker_context_t*)arg;
    if (ctx == NULL) {
        return NULL;
    }
    
    log_writer_t* writer = log_channel_register_writer(ctx->channel);
    if (writer == NULL) {
        fprintf(stderr, "Worker %d: Failed to register writer\n", ctx->thread_id);
        return NULL;
    }
    
    printf("Worker %d: Started, thread_id=%lu\n", 
           ctx->thread_id, (unsigned long)pthread_self());
    
    for (int i = 0; i < MESSAGES_PER_THREAD && !g_shutdown_requested; i++) {
        log_level_t level;
        const char* msg_type;
        
        switch (i % 4) {
            case 0:
                level = LOG_LEVEL_DEBUG;
                msg_type = "debug";
                break;
            case 1:
                level = LOG_LEVEL_INFO;
                msg_type = "info";
                break;
            case 2:
                level = LOG_LEVEL_WARN;
                msg_type = "warning";
                break;
            default:
                level = LOG_LEVEL_ERROR;
                msg_type = "error";
                break;
        }
        
        int ret = log_channel_write(ctx->channel, writer, level,
            "Worker %d: This is %s message #%d, value=%d, pi=%.5f",
            ctx->thread_id, msg_type, i + 1, i * 10, 3.14159);
        
        if (ret != 0) {
            fprintf(stderr, "Worker %d: Failed to write message #%d (ret=%d)\n",
                    ctx->thread_id, i + 1, ret);
        }
        
        usleep(10000 + (rand_r(&ctx->rand_seed) % 50000));
    }
    
    log_channel_write(ctx->channel, writer, LOG_LEVEL_INFO,
        "Worker %d: Finished writing all messages", ctx->thread_id);
    
    printf("Worker %d: Finished, unregistering writer\n", ctx->thread_id);
    
    log_channel_unregister_writer(ctx->channel, writer);
    
    return NULL;
}

int main(int argc, char* argv[]) {
    (void)argc;
    (void)argv;
    
    printf("=== Multi-Threaded Log Collector Demo ===\n");
    printf("Number of worker threads: %d\n", NUM_WORKER_THREADS);
    printf("Messages per thread: %d\n", MESSAGES_PER_THREAD);
    printf("Max total messages: %zu\n", (size_t)MAX_TOTAL_MESSAGES);
    printf("Overflow policy: DROP_OLDEST\n");
    printf("Output: STDOUT (sorted by timestamp)\n");
    printf("==========================================\n\n");
    
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    log_channel_config_t channel_config = {
        .max_total_messages = MAX_TOTAL_MESSAGES,
        .max_thread_messages = MAX_THREAD_MESSAGES,
        .overflow_policy = LOG_POLICY_DROP_OLDEST,
        .max_threads = LOG_MAX_THREADS
    };
    
    log_channel_t* channel = log_channel_create(&channel_config);
    if (channel == NULL) {
        fprintf(stderr, "Failed to create log channel\n");
        return 1;
    }
    
    log_collector_config_t collector_config = {
        .output_type = LOG_OUTPUT_STDOUT,
        .output_file = NULL,
        .append_to_file = false,
        .flush_interval_ms = 200,
        .sort_by_timestamp = true,
        .batch_size_hint = 100
    };
    
    log_collector_t* collector = log_collector_create(channel, &collector_config);
    if (collector == NULL) {
        fprintf(stderr, "Failed to create log collector\n");
        log_channel_destroy(channel);
        return 1;
    }
    
    if (log_collector_start(collector) != 0) {
        fprintf(stderr, "Failed to start log collector\n");
        log_collector_destroy(collector);
        log_channel_destroy(channel);
        return 1;
    }
    
    printf("Main: Log collector started\n\n");
    
    pthread_t worker_threads[NUM_WORKER_THREADS];
    worker_context_t contexts[NUM_WORKER_THREADS];
    int running = 1;
    
    unsigned int base_seed = (unsigned int)time(NULL);
    
    for (int i = 0; i < NUM_WORKER_THREADS; i++) {
        contexts[i].channel = channel;
        contexts[i].thread_id = i + 1;
        contexts[i].running = &running;
        contexts[i].rand_seed = base_seed + (unsigned int)(i * 12345);
        
        int ret = pthread_create(&worker_threads[i], NULL, 
                                  worker_thread_func, &contexts[i]);
        if (ret != 0) {
            fprintf(stderr, "Failed to create worker thread %d\n", i + 1);
            for (int j = 0; j < i; j++) {
                pthread_join(worker_threads[j], NULL);
            }
            log_collector_stop(collector, true);
            log_collector_destroy(collector);
            log_channel_destroy(channel);
            return 1;
        }
    }
    
    printf("Main: All %d worker threads started\n\n", NUM_WORKER_THREADS);
    
    for (int i = 0; i < NUM_WORKER_THREADS; i++) {
        pthread_join(worker_threads[i], NULL);
    }
    
    running = 0;
    printf("\nMain: All workers finished, waiting for collector to drain...\n");
    
    usleep(500000);
    
    size_t dropped = 0, written = 0;
    size_t current = log_channel_get_stats(channel, &dropped, &written);
    printf("\n=== Statistics ===\n");
    printf("Total messages written: %zu\n", written);
    printf("Total messages dropped: %zu\n", dropped);
    printf("Current queue size: %zu\n", current);
    
    uint64_t flushed = 0;
    uint64_t processed = log_collector_get_stats(collector, &flushed);
    printf("Messages processed by collector: %lu\n", (unsigned long)processed);
    printf("Flush count: %lu\n", (unsigned long)flushed);
    
    log_collector_stop(collector, true);
    
    printf("\nMain: Collector stopped, cleaning up...\n");
    
    log_collector_destroy(collector);
    log_channel_destroy(channel);
    
    printf("\n=== Demo completed successfully ===\n");
    
    return 0;
}
