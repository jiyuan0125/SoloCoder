#define _GNU_SOURCE
#include "logger.h"
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include <time.h>
#include <unistd.h>
#include <stdlib.h>

static const log_level_config_t g_level_configs[] = {
    { LOG_LEVEL_DEBUG, "DEBUG", 5, ANSI_FG_GRAY, 0, 0, 0 },
    { LOG_LEVEL_INFO,  "INFO ", 5, NULL,           0, 0, 0 },
    { LOG_LEVEL_WARN,  "WARN ", 5, ANSI_FG_YELLOW, 0, 0, 0 },
    { LOG_LEVEL_ERROR, "ERROR", 5, ANSI_FG_RED,    0, 0, 0 },
    { LOG_LEVEL_FATAL, "FATAL", 5, ANSI_FG_RED,    0, 1, 1 }
};

static const int g_level_count = sizeof(g_level_configs) / sizeof(g_level_configs[0]);

static logger_config_t g_default_config;
static int g_default_initialized = 0;

static pthread_once_t g_init_once = PTHREAD_ONCE_INIT;
static __thread log_thread_buffer_t* g_thread_buffer = NULL;

static void init_default_config(void)
{
    if (!g_default_initialized) {
        logger_config_init(&g_default_config);
        g_default_initialized = 1;
    }
}

const log_level_config_t* log_get_level_config(log_level_t level)
{
    for (int i = 0; i < g_level_count; i++) {
        if (g_level_configs[i].level == level) {
            return &g_level_configs[i];
        }
    }
    return NULL;
}

log_level_t log_level_from_name(const char* name)
{
    if (name == NULL) return LOG_LEVEL_OFF;

    for (int i = 0; i < g_level_count; i++) {
        if (strcasecmp(name, g_level_configs[i].name) == 0) {
            return g_level_configs[i].level;
        }
    }
    return LOG_LEVEL_OFF;
}

const char* log_level_to_name(log_level_t level)
{
    const log_level_config_t* cfg = log_get_level_config(level);
    if (cfg != NULL) {
        return cfg->name;
    }
    return "UNKNOWN";
}

int log_level_to_color_code(log_level_t level,
                             char* fg_code, size_t fg_size,
                             char* attr_code, size_t attr_size)
{
    const log_level_config_t* cfg = log_get_level_config(level);
    if (cfg == NULL) {
        if (fg_code != NULL && fg_size > 0) fg_code[0] = '\0';
        if (attr_code != NULL && attr_size > 0) attr_code[0] = '\0';
        return 0;
    }

    size_t written = 0;

    if (fg_code != NULL && fg_size > 0) {
        if (cfg->fg_color != NULL) {
            size_t len = strlen(cfg->fg_color);
            if (len + 1 <= fg_size) {
                memcpy(fg_code, cfg->fg_color, len + 1);
                written += (int)len;
            } else {
                fg_code[0] = '\0';
            }
        } else {
            if (ANSI_FG_DEFAULT && fg_size > 5) {
                memcpy(fg_code, ANSI_FG_DEFAULT, 6);
                written += 5;
            } else {
                fg_code[0] = '\0';
            }
        }
    }

    if (attr_code != NULL && attr_size > 0) {
        attr_code[0] = '\0';

        int attr_count = 0;
        int attrs[3];

        if (cfg->bold) attrs[attr_count++] = 1;
        if (cfg->blink) attrs[attr_count++] = 5;
        if (cfg->bright) attrs[attr_count++] = 2;

        if (attr_count > 0) {
            char* ptr = attr_code;
            size_t remaining = attr_size;

            if (remaining < 5) {
                attr_code[0] = '\0';
            } else {
                memcpy(ptr, "\033[", 2);
                ptr += 2;
                remaining -= 2;

                for (int i = 0; i < attr_count; i++) {
                    if (i > 0 && remaining > 1) {
                        *ptr++ = ';';
                        remaining--;
                    }
                    if (remaining > 0) {
                        int n = attrs[i];
                        if (n >= 10) {
                            if (remaining > 1) {
                                *ptr++ = (char)('0' + (n / 10));
                                remaining--;
                            }
                        }
                        if (remaining > 0) {
                            *ptr++ = (char)('0' + (n % 10));
                            remaining--;
                            written += 1;
                        }
                    }
                }

                if (remaining > 0) {
                    *ptr++ = 'm';
                    remaining--;
                }
                *ptr = '\0';

                written += 3;
            }
        }
    }

    return (int)written;
}

void logger_config_init(logger_config_t* config)
{
    if (config == NULL) return;

    memset(config, 0, sizeof(*config));
    config->min_level = LOG_LEVEL_DEBUG;
    config->use_color = -1;
    config->show_debug = 1;
    config->show_thread_id = 1;
    config->highlight_special = 1;

    logger_config_detect_terminal(config, STDOUT_FILENO);
}

void logger_config_set_level(logger_config_t* config, log_level_t level)
{
    if (config != NULL) {
        config->min_level = level;
    }
}

void logger_config_set_color(logger_config_t* config, int enabled)
{
    if (config != NULL) {
        config->use_color = enabled;
    }
}

void logger_config_detect_terminal(logger_config_t* config, int fd)
{
    if (config == NULL) return;

    config->term_info = terminal_detect(fd);

    if (config->use_color == -1) {
        config->use_color = terminal_should_use_color(&config->term_info);
    }
}

int logger_should_log(const logger_config_t* config, log_level_t level)
{
    if (config == NULL) {
        pthread_once(&g_init_once, init_default_config);
        config = &g_default_config;
    }
    return level >= config->min_level;
}

size_t log_format_time(char* buf, size_t buf_size, const struct timeval* tv)
{
    if (buf == NULL || buf_size == 0) return 0;

    struct tm tm_time;
    time_t sec = (time_t)tv->tv_sec;

    if (localtime_r(&sec, &tm_time) == NULL) {
        buf[0] = '\0';
        return 0;
    }

    int ms = (int)(tv->tv_usec / 1000);

    return (size_t)snprintf(buf, buf_size, "%02d:%02d:%02d.%03d",
                             tm_time.tm_hour,
                             tm_time.tm_min,
                             tm_time.tm_sec,
                             ms);
}

size_t log_format_thread_id(char* buf, size_t buf_size, pthread_t tid)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "%lu", (unsigned long)tid);
}

size_t log_format_level(char* buf, size_t buf_size, log_level_t level,
                         const logger_config_t* config)
{
    if (buf == NULL || buf_size == 0) return 0;

    const log_level_config_t* cfg = log_get_level_config(level);
    if (cfg == NULL) {
        buf[0] = '\0';
        return 0;
    }

    int use_color = 0;
    if (config != NULL) {
        use_color = (config->use_color != 0);
    } else {
        pthread_once(&g_init_once, init_default_config);
        use_color = (g_default_config.use_color != 0);
    }

    char* ptr = buf;
    size_t remaining = buf_size;
    size_t total = 0;

    if (use_color && (cfg->fg_color != NULL || cfg->bold || cfg->blink)) {
        char fg_code[64] = {0};
        char attr_code[64] = {0};

        log_level_to_color_code(level, fg_code, sizeof(fg_code),
                                 attr_code, sizeof(attr_code));

        if (attr_code[0] != '\0') {
            size_t len = strlen(attr_code);
            if (len <= remaining) {
                memcpy(ptr, attr_code, len);
                ptr += len;
                remaining -= len;
                total += len;
            }
        }
        if (fg_code[0] != '\0') {
            size_t len = strlen(fg_code);
            if (len <= remaining) {
                memcpy(ptr, fg_code, len);
                ptr += len;
                remaining -= len;
                total += len;
            }
        }
    }

    size_t name_len = (size_t)cfg->name_len;
    if (name_len + 1 <= remaining) {
        memcpy(ptr, cfg->name, name_len);
        ptr += name_len;
        remaining -= name_len;
        total += name_len;
    }

    if (use_color) {
        const char* reset = ANSI_RESET;
        size_t reset_len = strlen(reset);
        if (reset_len <= remaining) {
            memcpy(ptr, reset, reset_len);
            ptr += reset_len;
            remaining -= reset_len;
            total += reset_len;
        }
    }

    *ptr = '\0';
    return total;
}

static int is_special_marker(const char* start, const char* end,
                              int* out_type, size_t* out_len)
{
    if (start + 7 <= end &&
        memcmp(start, "[ERROR]", 7) == 0) {
        *out_type = 1;
        *out_len = 7;
        return 1;
    }
    if (start + 6 <= end &&
        memcmp(start, "[WARN]", 6) == 0) {
        *out_type = 2;
        *out_len = 6;
        return 1;
    }
    if (start + 7 <= end &&
        memcmp(start, "[DEBUG]", 7) == 0) {
        *out_type = 3;
        *out_len = 7;
        return 1;
    }
    if (start + 6 <= end &&
        memcmp(start, "[INFO]", 6) == 0) {
        *out_type = 4;
        *out_len = 6;
        return 1;
    }

    const char* p = start;
    if (p < end && *p != '[') {
        while (p < end && ((*p >= 'a' && *p <= 'z') ||
                          (*p >= 'A' && *p <= 'Z') ||
                          (*p >= '0' && *p <= '9') ||
                          *p == '_' || *p == '-' || *p == '.' ||
                          *p == '/' || *p == '\\')) {
            p++;
        }

        if (p < end && *p == ':') {
            p++;
            const char* line_start = p;
            while (p < end && *p >= '0' && *p <= '9') {
                p++;
            }

            if (p > line_start) {
                if (p == end || *p == ' ' || *p == '\t' || *p == '\n' || *p == ')') {
                    *out_type = 5;
                    *out_len = (size_t)(p - start);
                    return 1;
                }
                if (p < end && *p == ':') {
                    p++;
                    while (p < end && *p >= '0' && *p <= '9') {
                        p++;
                    }
                    *out_type = 5;
                    *out_len = (size_t)(p - start);
                    return 1;
                }
            }
        }
    }

    return 0;
}

size_t log_format_message_highlight(char* dest, size_t dest_size,
                                     const char* message, size_t msg_len,
                                     const logger_config_t* config)
{
    if (dest == NULL || dest_size == 0) return 0;
    if (message == NULL || msg_len == 0) {
        dest[0] = '\0';
        return 0;
    }

    int use_color = 0;
    int highlight = 0;

    if (config != NULL) {
        use_color = (config->use_color != 0);
        highlight = config->highlight_special;
    } else {
        pthread_once(&g_init_once, init_default_config);
        use_color = (g_default_config.use_color != 0);
        highlight = g_default_config.highlight_special;
    }

    if (!use_color || !highlight) {
        size_t copy_len = (msg_len < dest_size - 1) ? msg_len : (dest_size - 1);
        memcpy(dest, message, copy_len);
        dest[copy_len] = '\0';
        return copy_len;
    }

    const char* src = message;
    const char* src_end = message + msg_len;
    char* dst = dest;
    size_t dst_remaining = dest_size;
    size_t total = 0;

    const char* normal_start = src;

    while (src < src_end) {
        int type = 0;
        size_t match_len = 0;

        if (is_special_marker(src, src_end, &type, &match_len)) {
            if (src > normal_start) {
                size_t nlen = (size_t)(src - normal_start);
                if (nlen + 1 <= dst_remaining) {
                    memcpy(dst, normal_start, nlen);
                    dst += nlen;
                    dst_remaining -= nlen;
                    total += nlen;
                }
            }

            const char* color_start = NULL;
            const char* color_end = ANSI_RESET;

            switch (type) {
                case 1:
                    color_start = ANSI_FG_RED ANSI_ATTR_STR_BOLD;
                    break;
                case 2:
                    color_start = ANSI_FG_YELLOW ANSI_ATTR_STR_BOLD;
                    break;
                case 3:
                    color_start = ANSI_FG_GRAY;
                    break;
                case 4:
                    color_start = ANSI_FG_CYAN;
                    break;
                case 5:
                    color_start = ANSI_FG_MAGENTA ANSI_ATTR_STR_UNDERLINE;
                    break;
                default:
                    color_start = NULL;
                    break;
            }

            if (color_start != NULL) {
                size_t len = strlen(color_start);
                if (len + 1 <= dst_remaining) {
                    memcpy(dst, color_start, len);
                    dst += len;
                    dst_remaining -= len;
                    total += len;
                }
            }

            if (match_len + 1 <= dst_remaining) {
                memcpy(dst, src, match_len);
                dst += match_len;
                dst_remaining -= match_len;
                total += match_len;
            }

            if (color_start != NULL) {
                size_t len = strlen(color_end);
                if (len + 1 <= dst_remaining) {
                    memcpy(dst, color_end, len);
                    dst += len;
                    dst_remaining -= len;
                    total += len;
                }
            }

            src += match_len;
            normal_start = src;
        } else {
            src++;
        }
    }

    if (src > normal_start) {
        size_t nlen = (size_t)(src - normal_start);
        if (nlen + 1 <= dst_remaining) {
            memcpy(dst, normal_start, nlen);
            dst += nlen;
            dst_remaining -= nlen;
            total += nlen;
        }
    }

    *dst = '\0';
    return total;
}

size_t log_format_full(char* dest, size_t dest_size,
                       log_level_t level,
                       const struct timeval* tv,
                       pthread_t tid,
                       const char* message,
                       const logger_config_t* config)
{
    if (dest == NULL || dest_size == 0) return 0;
    if (message == NULL) message = "";

    char* ptr = dest;
    size_t remaining = dest_size;
    size_t total = 0;

    struct timeval tv_local;
    if (tv == NULL) {
        gettimeofday(&tv_local, NULL);
        tv = &tv_local;
    }

    pthread_t tid_local = 0;
    if (config == NULL || config->show_thread_id) {
        tid_local = pthread_self();
        if (tid == 0) tid = tid_local;
    }

    char time_buf[32];
    size_t time_len = log_format_time(time_buf, sizeof(time_buf), tv);

    size_t needed = time_len + 4;
    if (needed <= remaining) {
        *ptr++ = '[';
        remaining--;
        total++;

        memcpy(ptr, time_buf, time_len);
        ptr += time_len;
        remaining -= time_len;
        total += time_len;

        *ptr++ = ']';
        *ptr++ = ' ';
        remaining -= 2;
        total += 2;
    }

    log_thread_buffer_t* tbuf = log_get_thread_buffer();
    size_t level_len = log_format_level(tbuf->level_buf, sizeof(tbuf->level_buf),
                                         level, config);

    needed = level_len + 2;
    if (level_len > 0 && needed <= remaining) {
        *ptr++ = '[';
        remaining--;
        total++;

        memcpy(ptr, tbuf->level_buf, level_len);
        ptr += level_len;
        remaining -= level_len;
        total += level_len;

        *ptr++ = ']';
        remaining--;
        total++;

        if (remaining > 1) {
            *ptr++ = ' ';
            remaining--;
            total++;
        }
    }

    if (config == NULL || config->show_thread_id) {
        char tid_buf[32];
        size_t tid_len = log_format_thread_id(tid_buf, sizeof(tid_buf), tid);

        needed = tid_len + 4;
        if (tid_len > 0 && needed <= remaining) {
            *ptr++ = '[';
            remaining--;
            total++;

            memcpy(ptr, tid_buf, tid_len);
            ptr += tid_len;
            remaining -= tid_len;
            total += tid_len;

            *ptr++ = ']';
            remaining--;
            total++;

            if (remaining > 1) {
                *ptr++ = ' ';
                remaining--;
                total++;
            }
        }
    }

    size_t msg_len = strlen(message);
    if (msg_len > 0 && remaining > 1) {
        size_t hl_len = log_format_message_highlight(ptr, remaining, message, msg_len, config);
        ptr += hl_len;
        remaining -= hl_len;
        total += hl_len;
    }

    if (remaining > 1) {
        *ptr++ = '\n';
        remaining--;
        total++;
    }

    *ptr = '\0';
    return total;
}

log_thread_buffer_t* log_get_thread_buffer(void)
{
    if (g_thread_buffer == NULL) {
        g_thread_buffer = (log_thread_buffer_t*)calloc(1, sizeof(log_thread_buffer_t));
        if (g_thread_buffer == NULL) {
            static log_thread_buffer_t fallback;
            memset(&fallback, 0, sizeof(fallback));
            return &fallback;
        }
    }
    return g_thread_buffer;
}

static void ensure_default_config(void)
{
    pthread_once(&g_init_once, init_default_config);
}

size_t log_printf(log_level_t level, const char* format, ...)
{
    ensure_default_config();

    if (!logger_should_log(NULL, level)) {
        return 0;
    }

    if (level == LOG_LEVEL_DEBUG && !g_default_config.show_debug) {
        return 0;
    }

    log_thread_buffer_t* tbuf = log_get_thread_buffer();

    va_list args;
    va_start(args, format);
    int msg_len = vsnprintf(tbuf->buffer, LOG_BUFFER_SIZE, format, args);
    va_end(args);

    if (msg_len < 0) return 0;
    if ((size_t)msg_len >= LOG_BUFFER_SIZE) {
        tbuf->buffer[LOG_BUFFER_SIZE - 1] = '\0';
        msg_len = (int)(LOG_BUFFER_SIZE - 1);
    }

    struct timeval tv;
    gettimeofday(&tv, NULL);

    size_t log_len = log_format_full(tbuf->temp_code, sizeof(tbuf->temp_code),
                                      level, &tv, 0,
                                      tbuf->buffer, NULL);

    if (log_len == 0) return 0;

    if (g_default_config.use_color == 0 ||
        g_default_config.term_info.mode == TERM_OUTPUT_STRIP) {
        char stripped[LOG_BUFFER_SIZE];
        size_t stripped_len = terminal_strip_ansi_codes(stripped, sizeof(stripped),
                                                         tbuf->temp_code, log_len);
        if (stripped_len > 0) {
            fwrite(stripped, 1, stripped_len, stdout);
            return stripped_len;
        }
    }

    fwrite(tbuf->temp_code, 1, log_len, stdout);
    fflush(stdout);

    return log_len;
}

void log_set_default_config(const logger_config_t* config)
{
    if (config != NULL) {
        pthread_once(&g_init_once, init_default_config);
        g_default_config = *config;
    }
}

logger_config_t* log_get_default_config(void)
{
    pthread_once(&g_init_once, init_default_config);
    return &g_default_config;
}
