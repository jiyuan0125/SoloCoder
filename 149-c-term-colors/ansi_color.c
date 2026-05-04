#include "ansi_color.h"
#include <string.h>
#include <stdio.h>

static ansi_color_mode_t g_color_mode = ANSI_COLOR_MODE_TRUECOLOR;

size_t ansi_format_attr(char *buf, size_t buf_size, ansi_attr_t attr)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "\033[%dm", (int)attr);
}

size_t ansi_format_fg_16(char *buf, size_t buf_size, ansi_color_t color, int bright)
{
    if (buf == NULL || buf_size == 0) return 0;
    int base = bright ? ANSI_FG_COLOR_BRIGHT_BASE : ANSI_FG_COLOR_BASE;
    return (size_t)snprintf(buf, buf_size, "\033[%dm", base + (int)color);
}

size_t ansi_format_bg_16(char *buf, size_t buf_size, ansi_color_t color, int bright)
{
    if (buf == NULL || buf_size == 0) return 0;
    int base = bright ? ANSI_BG_COLOR_BRIGHT_BASE : ANSI_BG_COLOR_BASE;
    return (size_t)snprintf(buf, buf_size, "\033[%dm", base + (int)color);
}

size_t ansi_format_fg_256(char *buf, size_t buf_size, uint8_t color)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "\033[38;5;%um", color);
}

size_t ansi_format_bg_256(char *buf, size_t buf_size, uint8_t color)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "\033[48;5;%um", color);
}

size_t ansi_format_fg_rgb(char *buf, size_t buf_size, uint8_t r, uint8_t g, uint8_t b)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "\033[38;2;%u;%u;%um", r, g, b);
}

size_t ansi_format_bg_rgb(char *buf, size_t buf_size, uint8_t r, uint8_t g, uint8_t b)
{
    if (buf == NULL || buf_size == 0) return 0;
    return (size_t)snprintf(buf, buf_size, "\033[48;2;%u;%u;%um", r, g, b);
}

size_t ansi_format_reset(char *buf, size_t buf_size)
{
    if (buf == NULL || buf_size == 0) return 0;
    const char *reset = ANSI_RESET;
    size_t len = strlen(reset);
    if (len + 1 > buf_size) return 0;
    memcpy(buf, reset, len + 1);
    return len;
}

size_t ansi_format_combine(char *buf, size_t buf_size,
                           const char *fg_code, const char *bg_code,
                           const char *attr_codes[], size_t attr_count)
{
    if (buf == NULL || buf_size == 0) return 0;

    size_t total = 0;
    char *ptr = buf;
    size_t remaining = buf_size;

    if (remaining < 3) return 0;
    memcpy(ptr, "\033[", 2);
    ptr += 2;
    remaining -= 2;
    total += 2;

    int first = 1;

    for (size_t i = 0; i < attr_count && attr_codes[i] != NULL; i++) {
        const char *code = attr_codes[i];
        if (strncmp(code, "\033[", 2) == 0) {
            const char *num_start = code + 2;
            const char *num_end = num_start;
            while (*num_end >= '0' && *num_end <= '9') num_end++;

            size_t num_len = (size_t)(num_end - num_start);
            if (!first) {
                if (remaining < 2) break;
                *ptr++ = ';';
                remaining--;
                total++;
            }
            if (num_len >= remaining) break;
            memcpy(ptr, num_start, num_len);
            ptr += num_len;
            remaining -= num_len;
            total += num_len;
            first = 0;
        }
    }

    if (fg_code != NULL && strncmp(fg_code, "\033[", 2) == 0) {
        const char *num_start = fg_code + 2;
        const char *num_end = num_start;
        while (*num_end >= '0' && *num_end <= '9') num_end++;

        size_t num_len = (size_t)(num_end - num_start);
        if (num_len > 0) {
            if (!first) {
                if (remaining < 2) goto end;
                *ptr++ = ';';
                remaining--;
                total++;
            }
            if (num_len >= remaining) goto end;
            memcpy(ptr, num_start, num_len);
            ptr += num_len;
            remaining -= num_len;
            total += num_len;
            first = 0;
        }
    }

    if (bg_code != NULL && strncmp(bg_code, "\033[", 2) == 0) {
        const char *num_start = bg_code + 2;
        const char *num_end = num_start;
        while (*num_end >= '0' && *num_end <= '9') num_end++;

        size_t num_len = (size_t)(num_end - num_start);
        if (num_len > 0) {
            if (!first) {
                if (remaining < 2) goto end;
                *ptr++ = ';';
                remaining--;
                total++;
            }
            if (num_len >= remaining) goto end;
            memcpy(ptr, num_start, num_len);
            ptr += num_len;
            remaining -= num_len;
            total += num_len;
            first = 0;
        }
    }

end:
    if (remaining < 2) {
        buf[0] = '\0';
        return 0;
    }
    *ptr++ = 'm';
    remaining--;
    total++;
    *ptr = '\0';

    return total;
}

const char* ansi_get_fg_name(ansi_color_t color)
{
    static const char *names[] = {
        "BLACK", "RED", "GREEN", "YELLOW",
        "BLUE", "MAGENTA", "CYAN", "WHITE"
    };
    if (color >= ANSI_COLOR_BLACK && color <= ANSI_COLOR_WHITE) {
        return names[color];
    }
    return "DEFAULT";
}

const char* ansi_get_attr_name(ansi_attr_t attr)
{
    static const char *names[] = {
        "RESET", "BOLD", "DIM", "ITALIC", "UNDERLINE",
        "BLINK", "BLINK_FAST", "REVERSE", "HIDDEN", "STRIKETHROUGH"
    };
    if (attr >= ANSI_ATTR_RESET && attr <= ANSI_ATTR_STRIKETHROUGH) {
        return names[attr];
    }
    return "UNKNOWN";
}

ansi_color_mode_t ansi_get_color_mode(void)
{
    return g_color_mode;
}

void ansi_set_color_mode(ansi_color_mode_t mode)
{
    g_color_mode = mode;
}

int ansi_is_valid_256_color(uint8_t color)
{
    (void)color;
    return 1;
}
