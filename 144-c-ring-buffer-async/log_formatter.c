#include "log_formatter.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <stdarg.h>
#include <time.h>

uint64_t log_formatter_timestamp_us(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000000ULL + (uint64_t)tv.tv_usec;
}

const char* log_formatter_level_str(LogLevel level) {
    switch (level) {
        case LOG_LEVEL_DEBUG: return "DEBUG";
        case LOG_LEVEL_INFO:  return "INFO ";
        case LOG_LEVEL_WARN:  return "WARN ";
        case LOG_LEVEL_ERROR: return "ERROR";
        default: return "UNKNOWN";
    }
}

size_t log_formatter_vformat(
    char* buffer,
    size_t buffer_size,
    LogLevel level,
    pthread_t thread_id,
    const char* file,
    int line,
    const char* fmt,
    va_list args
) {
    if (!buffer || buffer_size < sizeof(LogMessageHeader)) {
        return 0;
    }

    LogMessageHeader* hdr = (LogMessageHeader*)buffer;
    hdr->magic = LOG_MSG_MAGIC;
    hdr->level = level;
    hdr->thread_id = thread_id;
    hdr->line = (uint32_t)line;
    hdr->timestamp_us = log_formatter_timestamp_us();

    const char* filename = strrchr(file, '/');
    if (filename) {
        filename++;
    } else {
        filename = file;
    }
    size_t file_len = strlen(filename);
    hdr->file_len = (uint32_t)file_len;

    size_t pos = sizeof(LogMessageHeader);
    size_t remaining = buffer_size - sizeof(LogMessageHeader);

    if (file_len >= remaining) {
        return 0;
    }
    memcpy(buffer + pos, filename, file_len);
    pos += file_len;
    remaining -= file_len;

    char content_buf[LOG_FORMATTER_MAX_LINE];
    int content_len = vsnprintf(content_buf, sizeof(content_buf), fmt, args);
    if (content_len < 0 || (size_t)content_len >= sizeof(content_buf)) {
        hdr->content_len = 0;
    } else {
        hdr->content_len = (uint32_t)content_len;
    }

    if (hdr->content_len >= remaining) {
        hdr->content_len = (uint32_t)(remaining - 1);
    }
    if (hdr->content_len > 0) {
        memcpy(buffer + pos, content_buf, hdr->content_len);
        pos += hdr->content_len;
    }

    hdr->total_len = (uint32_t)pos;
    return pos;
}

size_t log_formatter_format(
    char* buffer,
    size_t buffer_size,
    LogLevel level,
    pthread_t thread_id,
    const char* file,
    int line,
    const char* fmt,
    ...
) {
    va_list args;
    va_start(args, fmt);
    size_t result = log_formatter_vformat(buffer, buffer_size, level, thread_id, file, line, fmt, args);
    va_end(args);
    return result;
}

size_t log_formatter_format_output(
    char* output,
    size_t output_size,
    const char* raw_message,
    size_t raw_len
) {
    if (!output || output_size == 0 || !raw_message) {
        return 0;
    }

    const LogMessageHeader* hdr = (const LogMessageHeader*)raw_message;
    if (raw_len < sizeof(LogMessageHeader) || hdr->magic != LOG_MSG_MAGIC) {
        return 0;
    }

    time_t sec = (time_t)(hdr->timestamp_us / 1000000ULL);
    uint32_t usec = (uint32_t)(hdr->timestamp_us % 1000000ULL);
    struct tm tm;
    localtime_r(&sec, &tm);

    const char* file_ptr = raw_message + sizeof(LogMessageHeader);
    const char* content_ptr = file_ptr + hdr->file_len;

    int written = snprintf(
        output, output_size,
        "%04d-%02d-%02d %02d:%02d:%02d.%06u [%s] [TID:%lu] %.*s:%u - %.*s\n",
        tm.tm_year + 1900,
        tm.tm_mon + 1,
        tm.tm_mday,
        tm.tm_hour,
        tm.tm_min,
        tm.tm_sec,
        usec,
        log_formatter_level_str(hdr->level),
        (unsigned long)hdr->thread_id,
        (int)hdr->file_len,
        (hdr->file_len > 0) ? file_ptr : "",
        hdr->line,
        (int)hdr->content_len,
        (hdr->content_len > 0) ? content_ptr : ""
    );

    if (written < 0) {
        return 0;
    }

    return (size_t)written;
}
