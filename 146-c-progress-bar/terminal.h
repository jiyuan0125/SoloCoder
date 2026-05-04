#ifndef TERMINAL_H
#define TERMINAL_H

#include "progress.h"

void terminal_init(void);
void terminal_cleanup(void);
int terminal_is_supported(void);
int terminal_get_width(void);

void terminal_draw_progress(const ProgressState *state, int index, int total,
                             int terminal_supported, int terminal_width);
void terminal_draw_all(const ProgressManager *manager);
void terminal_clear_line(void);
void terminal_move_up(int lines);
void terminal_move_down(int lines);
void terminal_hide_cursor(void);
void terminal_show_cursor(void);

void bytes_format_unit(uint64_t bytes, char *buf, size_t buf_size);

#endif
