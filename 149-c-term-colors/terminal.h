#ifndef TERMINAL_H
#define TERMINAL_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    TERM_OUTPUT_AUTO,
    TERM_OUTPUT_FORCE_COLOR,
    TERM_OUTPUT_FORCE_NO_COLOR,
    TERM_OUTPUT_STRIP
} term_output_mode_t;

typedef enum {
    TERM_COLORS_UNKNOWN = 0,
    TERM_COLORS_8       = 8,
    TERM_COLORS_16      = 16,
    TERM_COLORS_256     = 256,
    TERM_COLORS_TRUE    = -1
} term_colors_t;

typedef struct {
    int is_tty;
    term_colors_t colors;
    term_output_mode_t mode;
    int supports_truecolor;
    int supports_256color;
} terminal_info_t;

int terminal_is_tty(int fd);

const char* terminal_getenv(const char* name);

int terminal_check_no_color(void);
int terminal_check_force_color(void);

terminal_info_t terminal_detect(int fd);

void terminal_set_mode(term_output_mode_t mode);
term_output_mode_t terminal_get_mode(void);

int terminal_should_use_color(const terminal_info_t *info);

size_t terminal_strip_ansi_codes(char *dest, size_t dest_size,
                                  const char *src, size_t src_len);

size_t terminal_strip_ansi_codes_inplace(char *str, size_t len);

term_colors_t terminal_get_color_capability(void);

int terminal_supports_truecolor(void);
int terminal_supports_256color(void);

#ifdef __cplusplus
}
#endif

#endif
