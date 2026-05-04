#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <stdint.h>
#include <inttypes.h>
#include <sys/time.h>

#include "async_logger.h"
#include "log_formatter.h"

#define DEFAULT_THREAD_COUNT 4
#define DEFAULT_LOGS_PER_THREAD 10000
#define DEFAULT_BUFFER_SIZE (16 * 1024 * 1024)

typedef struct {
    int thread_id;
    int log_count;
    _Atomic uint64_t* sequence_counter;
    _Atomic int* threads_ready;
    pthread_barrier_t* barrier;
} ThreadConfig;

static uint64_t get_timestamp_us(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000000ULL + (uint64_t)tv.tv_usec;
}

static void* log_thread_func(void* arg) {
    ThreadConfig* config = (ThreadConfig*)arg;
    
    if (config->barrier) {
        pthread_barrier_wait(config->barrier);
    }
    
    for (int i = 0; i < config->log_count; i++) {
        uint64_t seq = atomic_fetch_add(config->sequence_counter, 1) + 1;
        
        LogLevel level;
        switch (i % 4) {
            case 0: level = LOG_LEVEL_DEBUG; break;
            case 1: level = LOG_LEVEL_INFO; break;
            case 2: level = LOG_LEVEL_WARN; break;
            default: level = LOG_LEVEL_ERROR; break;
        }

        const char* file_str = "worker.c";
        int line = 100 + (i % 100);

        char content[256];
        snprintf(content, sizeof(content),
                 "Thread %d, Log %d/%d, Sequence: %" PRIu64,
                 config->thread_id, i + 1, config->log_count, seq);

        async_logger_log(g_logger, level, file_str, line, "%s", content);
    }

    if (config->threads_ready) {
        atomic_fetch_sub(config->threads_ready, 1);
    }
    
    return NULL;
}

static void print_usage(const char* prog_name) {
    printf("Usage: %s [options]\n", prog_name);
    printf("Options:\n");
    printf("  -t <count>      Number of writer threads (default: %d)\n", DEFAULT_THREAD_COUNT);
    printf("  -n <count>      Logs per thread (default: %d)\n", DEFAULT_LOGS_PER_THREAD);
    printf("  -b <size>       Ring buffer size in MB (default: %d)\n", DEFAULT_BUFFER_SIZE / (1024 * 1024));
    printf("  -f <path>       Log file path (default: stdout)\n");
    printf("  -p <policy>     Buffer policy: 'drop' or 'block' (default: drop)\n");
    printf("  -s <strategy>   Flush strategy: 'immediate' or 'threshold' (default: threshold)\n");
    printf("  -l <level>      Min log level: debug/info/warn/error (default: debug)\n");
    printf("  -h              Show this help\n");
}

int main(int argc, char* argv[]) {
    int thread_count = DEFAULT_THREAD_COUNT;
    int logs_per_thread = DEFAULT_LOGS_PER_THREAD;
    size_t buffer_size = DEFAULT_BUFFER_SIZE;
    const char* log_file = NULL;
    RingBufferPolicy buffer_policy = RB_POLICY_DROP;
    FlushPolicy flush_policy = FLUSH_POLICY_THRESHOLD;
    LogLevel min_level = LOG_LEVEL_DEBUG;

    int opt;
    while ((opt = getopt(argc, argv, "t:n:b:f:p:s:l:h")) != -1) {
        switch (opt) {
            case 't':
                thread_count = atoi(optarg);
                if (thread_count < 1) thread_count = 1;
                break;
            case 'n':
                logs_per_thread = atoi(optarg);
                if (logs_per_thread < 1) logs_per_thread = 1;
                break;
            case 'b':
                buffer_size = (size_t)atoi(optarg) * 1024 * 1024;
                if (buffer_size < 1024 * 1024) buffer_size = 1024 * 1024;
                break;
            case 'f':
                log_file = optarg;
                break;
            case 'p':
                if (strcmp(optarg, "block") == 0) {
                    buffer_policy = RB_POLICY_BLOCK;
                } else {
                    buffer_policy = RB_POLICY_DROP;
                }
                break;
            case 's':
                if (strcmp(optarg, "immediate") == 0) {
                    flush_policy = FLUSH_POLICY_IMMEDIATE;
                } else {
                    flush_policy = FLUSH_POLICY_THRESHOLD;
                }
                break;
            case 'l':
                if (strcmp(optarg, "info") == 0) {
                    min_level = LOG_LEVEL_INFO;
                } else if (strcmp(optarg, "warn") == 0) {
                    min_level = LOG_LEVEL_WARN;
                } else if (strcmp(optarg, "error") == 0) {
                    min_level = LOG_LEVEL_ERROR;
                } else {
                    min_level = LOG_LEVEL_DEBUG;
                }
                break;
            case 'h':
            default:
                print_usage(argv[0]);
                return 0;
        }
    }

    printf("=== Async Logger Performance Test ===\n");
    printf("Configuration:\n");
    printf("  Threads: %d\n", thread_count);
    printf("  Logs per thread: %d\n", logs_per_thread);
    printf("  Total expected logs: %d\n", thread_count * logs_per_thread);
    printf("  Buffer size: %zu MB\n", buffer_size / (1024 * 1024));
    printf("  Buffer policy: %s\n", buffer_policy == RB_POLICY_DROP ? "DROP (overwrite)" : "BLOCK");
    printf("  Flush policy: %s\n", flush_policy == FLUSH_POLICY_IMMEDIATE ? "IMMEDIATE" : "THRESHOLD");
    printf("  Log level: %s\n", log_formatter_level_str(min_level));
    printf("  Output: %s\n", log_file ? log_file : "stdout");
    printf("\n");

    AsyncLoggerConfig config = {
        .buffer_size = buffer_size,
        .buffer_policy = buffer_policy,
        .flush_policy = flush_policy,
        .flush_threshold = ASYNC_LOGGER_FLUSH_THRESHOLD,
        .log_file_path = log_file,
        .min_log_level = min_level,
        .writer_thread_priority = -10
    };

    g_logger = async_logger_create(&config);
    if (!g_logger) {
        fprintf(stderr, "Failed to create async logger\n");
        return 1;
    }

    printf("Logger created. Starting benchmark...\n\n");

    pthread_t* threads = (pthread_t*)malloc(sizeof(pthread_t) * thread_count);
    ThreadConfig* thread_configs = (ThreadConfig*)malloc(sizeof(ThreadConfig) * thread_count);
    _Atomic uint64_t sequence_counter;
    _Atomic int threads_ready;
    pthread_barrier_t barrier;

    atomic_init(&sequence_counter, 0);
    atomic_init(&threads_ready, thread_count);
    pthread_barrier_init(&barrier, NULL, thread_count);

    uint64_t start_time = get_timestamp_us();

    for (int i = 0; i < thread_count; i++) {
        thread_configs[i].thread_id = i;
        thread_configs[i].log_count = logs_per_thread;
        thread_configs[i].sequence_counter = &sequence_counter;
        thread_configs[i].threads_ready = &threads_ready;
        thread_configs[i].barrier = &barrier;

        pthread_attr_t attr;
        pthread_attr_init(&attr);
        pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

        int ret = pthread_create(&threads[i], &attr, log_thread_func, &thread_configs[i]);
        pthread_attr_destroy(&attr);

        if (ret != 0) {
            fprintf(stderr, "Failed to create thread %d\n", i);
            async_logger_destroy(g_logger);
            free(threads);
            free(thread_configs);
            pthread_barrier_destroy(&barrier);
            return 1;
        }
    }

    for (int i = 0; i < thread_count; i++) {
        pthread_join(threads[i], NULL);
    }

    uint64_t end_time = get_timestamp_us();
    double elapsed_sec = (double)(end_time - start_time) / 1000000.0;

    uint64_t total_expected = (uint64_t)thread_count * logs_per_thread;
    uint64_t total_attempted = atomic_load(&sequence_counter);
    size_t dropped = atomic_load(&g_logger->dropped_count);

    printf("=== Test Results ===\n");
    printf("Elapsed time: %.3f seconds\n", elapsed_sec);
    printf("Total attempted: %" PRIu64 "\n", total_attempted);
    printf("Total expected: %" PRIu64 "\n", total_expected);
    printf("Dropped messages: %zu\n", dropped);
    printf("Total written to file: %zu bytes\n", (size_t)atomic_load(&g_logger->total_written));
    
    if (elapsed_sec > 0) {
        double throughput = (double)total_attempted / elapsed_sec;
        double avg_latency_us = elapsed_sec * 1000000.0 / (double)total_attempted;
        printf("\n=== Performance ===\n");
        printf("Throughput: %.2f logs/sec\n", throughput);
        printf("Average latency per log: %.3f us\n", avg_latency_us);
        printf("Per-thread throughput: %.2f logs/sec/thread\n", throughput / thread_count);
    }

    if (dropped > 0) {
        printf("\n=== Warning ===\n");
        printf("Some messages were dropped (%zu). Consider:\n", dropped);
        printf("  1. Increasing buffer size\n");
        printf("  2. Using BLOCK policy instead of DROP\n");
        printf("  3. Reducing log volume\n");
    } else {
        printf("\n=== All messages processed successfully! ===\n");
    }

    printf("\nFlushing remaining data and cleaning up...\n");
    
    async_logger_destroy(g_logger);
    g_logger = NULL;

    free(threads);
    free(thread_configs);
    pthread_barrier_destroy(&barrier);

    printf("Done.\n");
    return 0;
}
