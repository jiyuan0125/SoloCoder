#include "terminal.h"
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

static term_output_mode_t g_output_mode = TERM_OUTPUT_AUTO;

int terminal_is_tty(int fd)
{
    return isatty(fd);
}

const char* terminal_getenv(const char* name)
{
    return getenv(name);
}

int terminal_check_no_color(void)
{
    const char* no_color = getenv("NO_COLOR");
    if (no_color == NULL) return 0;
    if (no_color[0] == '\0') return 0;
    if (strcmp(no_color, "0") == 0) return 0;
    return 1;
}

int terminal_check_force_color(void)
{
    const char* force_color = getenv("FORCE_COLOR");
    if (force_color == NULL) return 0;
    if (force_color[0] == '\0') return 1;
    if (strcmp(force_color, "0") == 0) return 0;
    return 1;
}

static term_colors_t term_colors_from_name(const char* name)
{
    if (name == NULL) return TERM_COLORS_UNKNOWN;

    if (strstr(name, "truecolor") || strstr(name, "24bit") ||
        strstr(name, "directcolor")) {
        return TERM_COLORS_TRUE;
    }
    if (strstr(name, "256color")) {
        return TERM_COLORS_256;
    }
    if (strstr(name, "16color")) {
        return TERM_COLORS_16;
    }
    if (strstr(name, "88color")) {
        return TERM_COLORS_256;
    }
    if (strstr(name, "xterm") || strstr(name, "rxvt") ||
        strstr(name, "screen") || strstr(name, "tmux")) {
        return TERM_COLORS_16;
    }
    if (strstr(name, "linux")) {
        return TERM_COLORS_8;
    }

    return TERM_COLORS_UNKNOWN;
}

int terminal_supports_256color(void)
{
    const char* term = getenv("TERM");
    const char* colorterm = getenv("COLORTERM");

    if (colorterm != NULL &&
        (strstr(colorterm, "256") || strstr(colorterm, "truecolor"))) {
        return 1;
    }

    if (term == NULL) return 0;

    if (strstr(term, "256color") || strstr(term, "88color") ||
        strstr(term, "truecolor") || strstr(term, "24bit")) {
        return 1;
    }

    return 0;
}

int terminal_supports_truecolor(void)
{
    const char* colorterm = getenv("COLORTERM");
    const char* term = getenv("TERM");

    if (colorterm != NULL &&
        (strstr(colorterm, "truecolor") || strstr(colorterm, "24bit"))) {
        return 1;
    }

    if (term == NULL) return 0;

    if (strstr(term, "truecolor") || strstr(term, "24bit") ||
        strstr(term, "directcolor")) {
        return 1;
    }

    if (strstr(term, "iterm2") || strstr(term, "kitty") ||
        strstr(term, "konsole") || strstr(term, "alacritty") ||
        strstr(term, "wezterm")) {
        return 1;
    }

    return 0;
}

term_colors_t terminal_get_color_capability(void)
{
    if (terminal_supports_truecolor()) {
        return TERM_COLORS_TRUE;
    }
    if (terminal_supports_256color()) {
        return TERM_COLORS_256;
    }

    const char* term = getenv("TERM");
    if (term != NULL) {
        return term_colors_from_name(term);
    }

    return TERM_COLORS_UNKNOWN;
}

terminal_info_t terminal_detect(int fd)
{
    terminal_info_t info;
    memset(&info, 0, sizeof(info));

    info.is_tty = terminal_is_tty(fd);
    info.colors = terminal_get_color_capability();
    info.mode = g_output_mode;
    info.supports_truecolor = terminal_supports_truecolor();
    info.supports_256color = terminal_supports_256color();

    return info;
}

void terminal_set_mode(term_output_mode_t mode)
{
    g_output_mode = mode;
}

term_output_mode_t terminal_get_mode(void)
{
    return g_output_mode;
}

int terminal_should_use_color(const terminal_info_t *info)
{
    if (info == NULL) return 0;

    switch (info->mode) {
        case TERM_OUTPUT_FORCE_COLOR:
            return 1;
        case TERM_OUTPUT_FORCE_NO_COLOR:
        case TERM_OUTPUT_STRIP:
            return 0;
        case TERM_OUTPUT_AUTO:
        default:
            break;
    }

    if (terminal_check_no_color()) {
        return 0;
    }
    if (terminal_check_force_color()) {
        return 1;
    }

    return info->is_tty;
}

size_t terminal_strip_ansi_codes(char *dest, size_t dest_size,
                                  const char *src, size_t src_len)
{
    if (dest == NULL || dest_size == 0) return 0;
    if (src == NULL || src_len == 0) {
        dest[0] = '\0';
        return 0;
    }

    size_t i = 0;
    size_t j = 0;
    int in_escape = 0;

    while (i < src_len && j < dest_size - 1) {
        if (in_escape) {
            if (src[i] >= '@' && src[i] <= '~' && src[i] != '[') {
                in_escape = 0;
            }
            i++;
        } else {
            if (src[i] == '\033') {
                in_escape = 1;
                i++;
            } else {
                dest[j++] = src[i++];
            }
        }
    }

    dest[j] = '\0';
    return j;
}

size_t terminal_strip_ansi_codes_inplace(char *str, size_t len)
{
    if (str == NULL || len == 0) return 0;

    size_t i = 0;
    size_t j = 0;
    int in_escape = 0;

    while (i < len) {
        if (in_escape) {
            if (str[i] >= '@' && str[i] <= '~' && str[i] != '[') {
                in_escape = 0;
            }
            i++;
        } else {
            if (str[i] == '\033') {
                in_escape = 1;
                i++;
            } else {
                if (i != j) {
                    str[j] = str[i];
                }
                j++;
                i++;
            }
        }
    }

    if (j < len) {
        str[j] = '\0';
    }

    return j;
}
