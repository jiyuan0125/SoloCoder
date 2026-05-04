#ifndef TERMINAL_H
#define TERMINAL_H

#include <termios.h>
#include <stddef.h>
#include <stdbool.h>

typedef struct {
    struct termios orig_termios;
    int is_raw;
} TerminalState;

void terminal_enable_raw_mode(void);
void terminal_disable_raw_mode(void);
int terminal_get_width(void);
int terminal_get_cursor_row(void);
void terminal_move_cursor(int cols);
void terminal_move_cursor_to(int col);
void terminal_clear_from_cursor(void);
void terminal_clear_screen(void);
void terminal_write(const char *s, size_t len);
void terminal_write_char(char c);
int terminal_read_key(void);
void terminal_sigint_handler(int sig);
void terminal_set_sigint_callback(void (*callback)(void));

#define ARROW_UP    1000
#define ARROW_DOWN  1001
#define ARROW_LEFT  1002
#define ARROW_RIGHT 1003
#define HOME_KEY    1004
#define END_KEY     1005
#define DELETE_KEY  1006
#define TAB_KEY     '\t'
#define CTRL_A      1
#define CTRL_B      2
#define CTRL_C      3
#define CTRL_D      4
#define CTRL_E      5
#define CTRL_F      6
#define CTRL_K      11
#define CTRL_U      21
#define CTRL_W      23
#define BACKSPACE   127

#endif
