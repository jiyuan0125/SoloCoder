#ifndef LOG_FORMATTER_H
#define LOG_FORMATTER_H

#include <stddef.h>
#include <stdarg.h>
#include <stdint.h>
#include <sys/time.h>
#include <pthread.h>

typedef enum {
    LOG_LEVEL_DEBUG = 0,
    LOG_LEVEL_INFO,
    LOG_LEVEL_WARN,
    LOG_LEVEL_ERROR
} LogLevel;

#define LOG_MSG_MAGIC 0x4C4F474D

typedef struct {
    uint32_t magic;
    uint32_t total_len;
    uint64_t timestamp_us;
    LogLevel level;
    pthread_t thread_id;
    uint32_t file_len;
    uint32_t line;
    uint32_t content_len;
} LogMessageHeader;

#define LOG_FORMATTER_MAX_LINE 4096

uint64_t log_formatter_timestamp_us(void);

const char* log_formatter_level_str(LogLevel level);

size_t log_formatter_format(
    char* buffer,
    size_t buffer_size,
    LogLevel level,
    pthread_t thread_id,
    const char* file,
    int line,
    const char* fmt,
    ...
);

size_t log_formatter_vformat(
    char* buffer,
    size_t buffer_size,
    LogLevel level,
    pthread_t thread_id,
    const char* file,
    int line,
    const char* fmt,
    va_list args
);

size_t log_formatter_format_output(
    char* output,
    size_t output_size,
    const char* raw_message,
    size_t raw_len
);

#endif
