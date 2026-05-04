#ifndef LOGGER_H
#define LOGGER_H

#include <stdint.h>
#include <stddef.h>
#include <pthread.h>
#include <sys/time.h>

#include "ansi_color.h"
#include "terminal.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    LOG_LEVEL_ALL   = 0,
    LOG_LEVEL_DEBUG = 1,
    LOG_LEVEL_INFO  = 2,
    LOG_LEVEL_WARN  = 3,
    LOG_LEVEL_ERROR = 4,
    LOG_LEVEL_FATAL = 5,
    LOG_LEVEL_OFF   = 6
} log_level_t;

typedef struct {
    log_level_t level;
    const char* name;
    int name_len;
    const char* fg_color;
    int bright;
    int blink;
    int bold;
} log_level_config_t;

typedef struct {
    log_level_t min_level;
    int use_color;
    int show_debug;
    int show_thread_id;
    int highlight_special;
    terminal_info_t term_info;
} logger_config_t;

#define LOG_BUFFER_SIZE 4096

#define LOG_TIME_FORMAT_SIZE 16
#define LOG_LEVEL_TEXT_SIZE 256

typedef struct {
    char buffer[LOG_BUFFER_SIZE];
    size_t used;
    char time_buf[LOG_TIME_FORMAT_SIZE];
    char level_buf[LOG_LEVEL_TEXT_SIZE];
    char temp_code[ANSI_COLOR_CODE_MAX_LEN * 4];
} log_thread_buffer_t;

const log_level_config_t* log_get_level_config(log_level_t level);

log_level_t log_level_from_name(const char* name);

const char* log_level_to_name(log_level_t level);

int log_level_to_color_code(log_level_t level,
                             char* fg_code, size_t fg_size,
                             char* attr_code, size_t attr_size);

void logger_config_init(logger_config_t* config);

void logger_config_set_level(logger_config_t* config, log_level_t level);

void logger_config_set_color(logger_config_t* config, int enabled);

void logger_config_detect_terminal(logger_config_t* config, int fd);

int logger_should_log(const logger_config_t* config, log_level_t level);

size_t log_format_time(char* buf, size_t buf_size, const struct timeval* tv);

size_t log_format_thread_id(char* buf, size_t buf_size, pthread_t tid);

size_t log_format_level(char* buf, size_t buf_size, log_level_t level,
                         const logger_config_t* config);

size_t log_format_message_highlight(char* dest, size_t dest_size,
                                     const char* message, size_t msg_len,
                                     const logger_config_t* config);

size_t log_format_full(char* dest, size_t dest_size,
                       log_level_t level,
                       const struct timeval* tv,
                       pthread_t tid,
                       const char* message,
                       const logger_config_t* config);

log_thread_buffer_t* log_get_thread_buffer(void);

size_t log_printf(log_level_t level, const char* format, ...);

void log_set_default_config(const logger_config_t* config);
logger_config_t* log_get_default_config(void);

#define LOG_DEBUG(fmt, ...) log_printf(LOG_LEVEL_DEBUG, fmt, ##__VA_ARGS__)
#define LOG_INFO(fmt, ...)  log_printf(LOG_LEVEL_INFO,  fmt, ##__VA_ARGS__)
#define LOG_WARN(fmt, ...)  log_printf(LOG_LEVEL_WARN,  fmt, ##__VA_ARGS__)
#define LOG_ERROR(fmt, ...) log_printf(LOG_LEVEL_ERROR, fmt, ##__VA_ARGS__)
#define LOG_FATAL(fmt, ...) log_printf(LOG_LEVEL_FATAL, fmt, ##__VA_ARGS__)

#ifdef __cplusplus
}
#endif

#endif
