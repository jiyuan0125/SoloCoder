#ifndef ASYNC_LOGGER_H
#define ASYNC_LOGGER_H

#include "log_formatter.h"
#include "ring_buffer.h"
#include <stdio.h>
#include <pthread.h>

#define ASYNC_LOGGER_DEFAULT_BUFFER_SIZE (64 * 1024 * 1024)
#define ASYNC_LOGGER_FLUSH_THRESHOLD (32 * 1024)
#define ASYNC_LOGGER_HIGH_WATERMARK 0.85
#define ASYNC_LOGGER_LOW_WATERMARK 0.50

typedef enum {
    FLUSH_POLICY_IMMEDIATE,
    FLUSH_POLICY_THRESHOLD
} FlushPolicy;

typedef struct {
    size_t buffer_size;
    RingBufferPolicy buffer_policy;
    FlushPolicy flush_policy;
    size_t flush_threshold;
    const char* log_file_path;
    LogLevel min_log_level;
    int writer_thread_priority;
} AsyncLoggerConfig;

typedef struct {
    RingBuffer* ring_buffer;
    FILE* log_file;
    pthread_t writer_thread;
    _Atomic bool is_running;
    _Atomic bool writer_boosted;
    FlushPolicy flush_policy;
    size_t flush_threshold;
    LogLevel min_log_level;
    pthread_mutex_t flush_mutex;
    pthread_cond_t flush_cond;
    _Atomic size_t dropped_count;
    _Atomic uint64_t total_written;
} AsyncLogger;

extern AsyncLogger* g_logger;

AsyncLogger* async_logger_create(const AsyncLoggerConfig* config);
void async_logger_destroy(AsyncLogger* logger);

void async_logger_log(
    AsyncLogger* logger,
    LogLevel level,
    const char* file,
    int line,
    const char* fmt,
    ...
);

void async_logger_set_level(AsyncLogger* logger, LogLevel level);
void async_logger_flush(AsyncLogger* logger);
void async_logger_flush_all(AsyncLogger* logger);

#define LOG_DEBUG(fmt, ...) \
    async_logger_log(g_logger, LOG_LEVEL_DEBUG, __FILE__, __LINE__, fmt, ##__VA_ARGS__)

#define LOG_INFO(fmt, ...) \
    async_logger_log(g_logger, LOG_LEVEL_INFO, __FILE__, __LINE__, fmt, ##__VA_ARGS__)

#define LOG_WARN(fmt, ...) \
    async_logger_log(g_logger, LOG_LEVEL_WARN, __FILE__, __LINE__, fmt, ##__VA_ARGS__)

#define LOG_ERROR(fmt, ...) \
    async_logger_log(g_logger, LOG_LEVEL_ERROR, __FILE__, __LINE__, fmt, ##__VA_ARGS__)

#endif
