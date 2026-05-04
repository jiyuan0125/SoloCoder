#ifndef ANSI_COLOR_H
#define ANSI_COLOR_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    ANSI_ATTR_RESET        = 0,
    ANSI_ATTR_BOLD         = 1,
    ANSI_ATTR_DIM          = 2,
    ANSI_ATTR_ITALIC       = 3,
    ANSI_ATTR_UNDERLINE    = 4,
    ANSI_ATTR_BLINK        = 5,
    ANSI_ATTR_BLINK_FAST   = 6,
    ANSI_ATTR_REVERSE      = 7,
    ANSI_ATTR_HIDDEN       = 8,
    ANSI_ATTR_STRIKETHROUGH = 9
} ansi_attr_t;

typedef enum {
    ANSI_COLOR_BLACK   = 0,
    ANSI_COLOR_RED     = 1,
    ANSI_COLOR_GREEN   = 2,
    ANSI_COLOR_YELLOW  = 3,
    ANSI_COLOR_BLUE    = 4,
    ANSI_COLOR_MAGENTA = 5,
    ANSI_COLOR_CYAN    = 6,
    ANSI_COLOR_WHITE   = 7,
    ANSI_COLOR_DEFAULT = 9
} ansi_color_t;

typedef enum {
    ANSI_COLOR_MODE_16,
    ANSI_COLOR_MODE_256,
    ANSI_COLOR_MODE_TRUECOLOR
} ansi_color_mode_t;

#define ANSI_FG_COLOR_BASE 30
#define ANSI_BG_COLOR_BASE 40
#define ANSI_FG_COLOR_BRIGHT_BASE 90
#define ANSI_BG_COLOR_BRIGHT_BASE 100

#define ANSI_ESC_START "\033["
#define ANSI_ESC_END "m"
#define ANSI_RESET "\033[0m"

#define ANSI_FG_BLACK   "\033[30m"
#define ANSI_FG_RED     "\033[31m"
#define ANSI_FG_GREEN   "\033[32m"
#define ANSI_FG_YELLOW  "\033[33m"
#define ANSI_FG_BLUE    "\033[34m"
#define ANSI_FG_MAGENTA "\033[35m"
#define ANSI_FG_CYAN    "\033[36m"
#define ANSI_FG_WHITE   "\033[37m"
#define ANSI_FG_DEFAULT "\033[39m"

#define ANSI_BG_BLACK   "\033[40m"
#define ANSI_BG_RED     "\033[41m"
#define ANSI_BG_GREEN   "\033[42m"
#define ANSI_BG_YELLOW  "\033[43m"
#define ANSI_BG_BLUE    "\033[44m"
#define ANSI_BG_MAGENTA "\033[45m"
#define ANSI_BG_CYAN    "\033[46m"
#define ANSI_BG_WHITE   "\033[47m"
#define ANSI_BG_DEFAULT "\033[49m"

#define ANSI_FG_GRAY    "\033[90m"

#define ANSI_ATTR_STR_BOLD      "\033[1m"
#define ANSI_ATTR_STR_DIM       "\033[2m"
#define ANSI_ATTR_STR_ITALIC    "\033[3m"
#define ANSI_ATTR_STR_UNDERLINE "\033[4m"
#define ANSI_ATTR_STR_BLINK     "\033[5m"

#define ANSI_COLOR_CODE_MAX_LEN 32

size_t ansi_format_attr(char *buf, size_t buf_size, ansi_attr_t attr);

size_t ansi_format_fg_16(char *buf, size_t buf_size, ansi_color_t color, int bright);
size_t ansi_format_bg_16(char *buf, size_t buf_size, ansi_color_t color, int bright);

size_t ansi_format_fg_256(char *buf, size_t buf_size, uint8_t color);
size_t ansi_format_bg_256(char *buf, size_t buf_size, uint8_t color);

size_t ansi_format_fg_rgb(char *buf, size_t buf_size, uint8_t r, uint8_t g, uint8_t b);
size_t ansi_format_bg_rgb(char *buf, size_t buf_size, uint8_t r, uint8_t g, uint8_t b);

size_t ansi_format_reset(char *buf, size_t buf_size);

size_t ansi_format_combine(char *buf, size_t buf_size,
                           const char *fg_code, const char *bg_code,
                           const char *attr_codes[], size_t attr_count);

const char* ansi_get_fg_name(ansi_color_t color);
const char* ansi_get_attr_name(ansi_attr_t attr);

ansi_color_mode_t ansi_get_color_mode(void);
void ansi_set_color_mode(ansi_color_mode_t mode);

int ansi_is_valid_256_color(uint8_t color);

#ifdef __cplusplus
}
#endif

#endif
