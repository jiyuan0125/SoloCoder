#include "async_logger.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sched.h>
#include <sys/resource.h>
#include <stdarg.h>
#include <errno.h>

AsyncLogger* g_logger = NULL;

#define READ_BUFFER_SIZE (128 * 1024)

static void set_thread_priority(int nice_value) {
    setpriority(PRIO_PROCESS, 0, nice_value);
}

static void* writer_thread_func(void* arg) {
    AsyncLogger* logger = (AsyncLogger*)arg;
    uint8_t* read_buffer = (uint8_t*)malloc(READ_BUFFER_SIZE);
    char* output_buffer = (char*)malloc(READ_BUFFER_SIZE * 2);
    size_t unflushed_bytes = 0;

    if (!read_buffer || !output_buffer) {
        free(read_buffer);
        free(output_buffer);
        return NULL;
    }

    set_thread_priority(-10);

    while (atomic_load(&logger->is_running) || !ring_buffer_is_empty(logger->ring_buffer)) {
        size_t used = ring_buffer_used(logger->ring_buffer);
        size_t buffer_size = logger->ring_buffer->buffer_size;
        double usage = (double)used / (double)buffer_size;

        if (usage >= ASYNC_LOGGER_HIGH_WATERMARK && !atomic_load(&logger->writer_boosted)) {
            set_thread_priority(-20);
            atomic_store(&logger->writer_boosted, true);
        } else if (usage <= ASYNC_LOGGER_LOW_WATERMARK && atomic_load(&logger->writer_boosted)) {
            set_thread_priority(-10);
            atomic_store(&logger->writer_boosted, false);
        }

        size_t bytes_read = 0;
        bool has_data = ring_buffer_read(logger->ring_buffer, read_buffer, READ_BUFFER_SIZE, &bytes_read);

        if (!has_data || bytes_read == 0) {
            if (unflushed_bytes > 0) {
                fflush(logger->log_file);
                unflushed_bytes = 0;
            }
            
            struct timespec ts;
            ts.tv_sec = 0;
            ts.tv_nsec = 100000;
            nanosleep(&ts, NULL);
            continue;
        }

        size_t pos = 0;
        while (pos < bytes_read) {
            const LogMessageHeader* hdr = (const LogMessageHeader*)(read_buffer + pos);
            if (hdr->magic != LOG_MSG_MAGIC) {
                pos++;
                continue;
            }

            size_t output_len = log_formatter_format_output(
                output_buffer,
                READ_BUFFER_SIZE * 2,
                (const char*)hdr,
                bytes_read - pos
            );

            if (output_len > 0) {
                fwrite(output_buffer, 1, output_len, logger->log_file);
                unflushed_bytes += output_len;
                atomic_fetch_add(&logger->total_written, output_len);
            }

            pos += hdr->total_len;
        }

        if (logger->flush_policy == FLUSH_POLICY_IMMEDIATE || 
            unflushed_bytes >= logger->flush_threshold) {
            fflush(logger->log_file);
            unflushed_bytes = 0;
        }
    }

    if (unflushed_bytes > 0) {
        fflush(logger->log_file);
    }

    free(read_buffer);
    free(output_buffer);
    return NULL;
}

AsyncLogger* async_logger_create(const AsyncLoggerConfig* config) {
    if (!config) {
        return NULL;
    }

    AsyncLogger* logger = (AsyncLogger*)calloc(1, sizeof(AsyncLogger));
    if (!logger) {
        return NULL;
    }

    RingBufferConfig rb_config = {
        .buffer_size = config->buffer_size > 0 ? config->buffer_size : ASYNC_LOGGER_DEFAULT_BUFFER_SIZE,
        .policy = config->buffer_policy
    };

    logger->ring_buffer = ring_buffer_create(&rb_config);
    if (!logger->ring_buffer) {
        free(logger);
        return NULL;
    }

    if (config->log_file_path && config->log_file_path[0]) {
        logger->log_file = fopen(config->log_file_path, "a");
        if (!logger->log_file) {
            ring_buffer_destroy(logger->ring_buffer);
            free(logger);
            return NULL;
        }
    } else {
        logger->log_file = stdout;
    }

    logger->flush_policy = config->flush_policy;
    logger->flush_threshold = config->flush_threshold > 0 ? config->flush_threshold : ASYNC_LOGGER_FLUSH_THRESHOLD;
    logger->min_log_level = config->min_log_level;
    atomic_init(&logger->is_running, true);
    atomic_init(&logger->writer_boosted, false);
    atomic_init(&logger->dropped_count, 0);
    atomic_init(&logger->total_written, 0);

    pthread_mutex_init(&logger->flush_mutex, NULL);
    pthread_cond_init(&logger->flush_cond, NULL);

    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    int ret = pthread_create(&logger->writer_thread, &attr, writer_thread_func, logger);
    pthread_attr_destroy(&attr);

    if (ret != 0) {
        if (logger->log_file != stdout) {
            fclose(logger->log_file);
        }
        ring_buffer_destroy(logger->ring_buffer);
        pthread_mutex_destroy(&logger->flush_mutex);
        pthread_cond_destroy(&logger->flush_cond);
        free(logger);
        return NULL;
    }

    return logger;
}

void async_logger_destroy(AsyncLogger* logger) {
    if (!logger) {
        return;
    }

    atomic_store(&logger->is_running, false);
    pthread_join(logger->writer_thread, NULL);

    async_logger_flush_all(logger);

    if (logger->log_file != stdout) {
        fclose(logger->log_file);
    }

    ring_buffer_destroy(logger->ring_buffer);
    pthread_mutex_destroy(&logger->flush_mutex);
    pthread_cond_destroy(&logger->flush_cond);
    free(logger);
}

void async_logger_log(
    AsyncLogger* logger,
    LogLevel level,
    const char* file,
    int line,
    const char* fmt,
    ...
) {
    if (!logger || level < logger->min_log_level) {
        return;
    }

    if (!atomic_load(&logger->is_running)) {
        return;
    }

    char buffer[LOG_FORMATTER_MAX_LINE];
    va_list args;
    va_start(args, fmt);

    size_t msg_len = log_formatter_vformat(
        buffer,
        sizeof(buffer),
        level,
        pthread_self(),
        file,
        line,
        fmt,
        args
    );

    va_end(args);

    if (msg_len == 0) {
        return;
    }

    bool success = ring_buffer_write(logger->ring_buffer, buffer, msg_len);
    if (!success) {
        atomic_fetch_add(&logger->dropped_count, 1);
    }
}

void async_logger_set_level(AsyncLogger* logger, LogLevel level) {
    if (logger) {
        logger->min_log_level = level;
    }
}

void async_logger_flush(AsyncLogger* logger) {
    if (!logger || !logger->log_file) {
        return;
    }
    fflush(logger->log_file);
}

void async_logger_flush_all(AsyncLogger* logger) {
    if (!logger) {
        return;
    }

    uint8_t* read_buffer = (uint8_t*)malloc(READ_BUFFER_SIZE);
    char* output_buffer = (char*)malloc(READ_BUFFER_SIZE * 2);

    if (!read_buffer || !output_buffer) {
        free(read_buffer);
        free(output_buffer);
        return;
    }

    while (!ring_buffer_is_empty(logger->ring_buffer)) {
        size_t bytes_read = 0;
        bool has_data = ring_buffer_read(logger->ring_buffer, read_buffer, READ_BUFFER_SIZE, &bytes_read);

        if (!has_data || bytes_read == 0) {
            break;
        }

        size_t pos = 0;
        while (pos < bytes_read) {
            const LogMessageHeader* hdr = (const LogMessageHeader*)(read_buffer + pos);
            if (hdr->magic != LOG_MSG_MAGIC) {
                pos++;
                continue;
            }

            size_t output_len = log_formatter_format_output(
                output_buffer,
                READ_BUFFER_SIZE * 2,
                (const char*)hdr,
                bytes_read - pos
            );

            if (output_len > 0) {
                fwrite(output_buffer, 1, output_len, logger->log_file);
            }

            pos += hdr->total_len;
        }
    }

    fflush(logger->log_file);
    free(read_buffer);
    free(output_buffer);
}
