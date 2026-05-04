#include "lineedit.h"
#include "terminal.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

#define INITIAL_BUF_SIZE 128

static CompleteCallback g_complete_cb = NULL;
static LineEdit *g_current_le = NULL;

static void lineedit_sigint_callback(void) {
    if (g_current_le) {
        g_current_le->interrupted = 1;
    }
}

static void lineedit_ensure_buf(LineEdit *le, size_t needed) {
    if (le->buf_size <= needed + 1) {
        size_t new_size = le->buf_size * 2;
        while (new_size <= needed + 1) {
            new_size *= 2;
        }
        le->buf = (char *)realloc(le->buf, new_size);
        le->buf_size = new_size;
    }
}

static int lineedit_display_cols(const LineEdit *le) {
    (void)le;
    return terminal_get_width();
}

static void lineedit_calculate_screen_pos(const LineEdit *le, 
        size_t display_pos, int *row, int *col) {
    int cols = lineedit_display_cols(le);
    size_t total = le->prompt_len + display_pos;
    *row = (int)(total / cols);
    *col = (int)(total % cols);
}

static void lineedit_clear_display(LineEdit *le) {
    int cursor_row, cursor_col;
    lineedit_calculate_screen_pos(le, le->cursor, &cursor_row, &cursor_col);
    
    if (cursor_row > 0) {
        char up[16];
        int n = snprintf(up, sizeof(up), "\x1b[%dA", cursor_row);
        terminal_write(up, n);
    }
    terminal_move_cursor_to(0);
    terminal_clear_from_cursor();
}

static void lineedit_display(LineEdit *le) {
    int cols = lineedit_display_cols(le);
    
    terminal_write(le->prompt, le->prompt_len);
    terminal_write(le->buf, le->len);
    
    int cursor_row, cursor_col;
    lineedit_calculate_screen_pos(le, le->cursor, &cursor_row, &cursor_col);
    
    int current_row = (int)((le->prompt_len + le->len + cols - 1) / cols) - 1;
    int rows_to_move = cursor_row - current_row;
    
    if (rows_to_move < 0) {
        char up[16];
        int n = snprintf(up, sizeof(up), "\x1b[%dA", -rows_to_move);
        terminal_write(up, n);
    } else if (rows_to_move > 0) {
        char down[16];
        int n = snprintf(down, sizeof(down), "\x1b[%dB", rows_to_move);
        terminal_write(down, n);
    }
    
    terminal_move_cursor_to(cursor_col);
}

static void lineedit_refresh(LineEdit *le) {
    lineedit_clear_display(le);
    lineedit_display(le);
}

void lineedit_init(LineEdit *le, History *h) {
    le->buf = (char *)malloc(INITIAL_BUF_SIZE);
    le->buf[0] = '\0';
    le->buf_size = INITIAL_BUF_SIZE;
    le->len = 0;
    le->cursor = 0;
    le->prompt = "> ";
    le->prompt_len = 2;
    le->history = h;
    le->saved_line = NULL;
    le->interrupted = 0;
}

void lineedit_free(LineEdit *le) {
    free(le->buf);
    free(le->saved_line);
    le->buf = NULL;
    le->saved_line = NULL;
    le->buf_size = 0;
    le->len = 0;
    le->cursor = 0;
}

void lineedit_set_prompt(LineEdit *le, const char *prompt) {
    le->prompt = prompt;
    le->prompt_len = strlen(prompt);
}

void lineedit_set_complete_callback(CompleteCallback cb) {
    g_complete_cb = cb;
}

static void lineedit_save_current_line(LineEdit *le) {
    if (le->saved_line) {
        free(le->saved_line);
    }
    le->saved_line = strndup(le->buf, le->len);
}

static void lineedit_restore_saved_line(LineEdit *le) {
    if (le->saved_line) {
        size_t saved_len = strlen(le->saved_line);
        lineedit_ensure_buf(le, saved_len);
        strcpy(le->buf, le->saved_line);
        le->len = saved_len;
        le->cursor = saved_len;
        free(le->saved_line);
        le->saved_line = NULL;
    }
}

static void lineedit_set_line(LineEdit *le, const char *line) {
    if (line == NULL) return;
    size_t line_len = strlen(line);
    lineedit_ensure_buf(le, line_len);
    strcpy(le->buf, line);
    le->len = line_len;
    le->cursor = line_len;
}

static void lineedit_insert_char(LineEdit *le, char c) {
    lineedit_ensure_buf(le, le->len + 1);
    
    if (le->cursor < le->len) {
        memmove(&le->buf[le->cursor + 1], &le->buf[le->cursor], 
                le->len - le->cursor);
    }
    le->buf[le->cursor] = c;
    le->len++;
    le->cursor++;
    le->buf[le->len] = '\0';
    
    lineedit_refresh(le);
}

static void lineedit_backspace(LineEdit *le) {
    if (le->cursor == 0) return;
    
    le->cursor--;
    if (le->cursor < le->len) {
        memmove(&le->buf[le->cursor], &le->buf[le->cursor + 1], 
                le->len - le->cursor - 1);
    }
    le->len--;
    le->buf[le->len] = '\0';
    
    lineedit_refresh(le);
}

static void lineedit_delete(LineEdit *le) {
    if (le->cursor >= le->len) return;
    
    if (le->cursor < le->len - 1) {
        memmove(&le->buf[le->cursor], &le->buf[le->cursor + 1], 
                le->len - le->cursor - 1);
    }
    le->len--;
    le->buf[le->len] = '\0';
    
    lineedit_refresh(le);
}

static void lineedit_move_left(LineEdit *le) {
    if (le->cursor > 0) {
        le->cursor--;
        lineedit_refresh(le);
    }
}

static void lineedit_move_right(LineEdit *le) {
    if (le->cursor < le->len) {
        le->cursor++;
        lineedit_refresh(le);
    }
}

static void lineedit_move_home(LineEdit *le) {
    if (le->cursor != 0) {
        le->cursor = 0;
        lineedit_refresh(le);
    }
}

static void lineedit_move_end(LineEdit *le) {
    if (le->cursor != le->len) {
        le->cursor = le->len;
        lineedit_refresh(le);
    }
}

static void lineedit_kill_before(LineEdit *le) {
    if (le->cursor == 0) return;
    
    memmove(le->buf, &le->buf[le->cursor], le->len - le->cursor);
    le->len -= le->cursor;
    le->cursor = 0;
    le->buf[le->len] = '\0';
    
    lineedit_refresh(le);
}

static void lineedit_kill_after(LineEdit *le) {
    if (le->cursor >= le->len) return;
    
    le->len = le->cursor;
    le->buf[le->len] = '\0';
    
    lineedit_refresh(le);
}

static void lineedit_history_prev(LineEdit *le) {
    if (!le->history) return;
    
    if (le->saved_line == NULL && le->len > 0) {
        lineedit_save_current_line(le);
    }
    
    const char *prev = history_prev(le->history);
    if (prev) {
        lineedit_set_line(le, prev);
        lineedit_refresh(le);
    }
}

static void lineedit_history_next(LineEdit *le) {
    if (!le->history) return;
    
    const char *next = history_next(le->history);
    if (next) {
        lineedit_set_line(le, next);
        lineedit_refresh(le);
    } else if (le->saved_line) {
        lineedit_restore_saved_line(le);
        lineedit_refresh(le);
    }
}

static char **get_completion_matches(const char *line, int *count) {
    if (g_complete_cb) {
        return g_complete_cb(line, count);
    }
    *count = 0;
    return NULL;
}

static void free_completions(char **completions, int count) {
    for (int i = 0; i < count; i++) {
        free(completions[i]);
    }
    free(completions);
}

static size_t find_common_prefix(const char **strs, int count) {
    if (count == 0) return 0;
    
    size_t max_len = strlen(strs[0]);
    for (int i = 1; i < count; i++) {
        size_t len = strlen(strs[i]);
        if (len < max_len) max_len = len;
    }
    
    for (size_t pos = 0; pos < max_len; pos++) {
        char c = strs[0][pos];
        for (int i = 1; i < count; i++) {
            if (strs[i][pos] != c) {
                return pos;
            }
        }
    }
    return max_len;
}

static void lineedit_complete(LineEdit *le) {
    int count = 0;
    char **completions = get_completion_matches(le->buf, &count);
    
    if (count == 0) {
        if (completions) free(completions);
        terminal_write("\a", 1);
        return;
    }
    
    if (count == 1) {
        size_t prefix_len = find_common_prefix((const char **)completions, count);
        if (prefix_len > le->len) {
            for (size_t i = le->len; i < prefix_len; i++) {
                lineedit_insert_char(le, completions[0][i]);
            }
        }
        free_completions(completions, count);
        return;
    }
    
    size_t common_prefix = find_common_prefix((const char **)completions, count);
    if (common_prefix > le->len) {
        for (size_t i = le->len; i < common_prefix; i++) {
            lineedit_insert_char(le, completions[0][i]);
        }
    }
    
    terminal_write("\r\n", 2);
    
    int cols = lineedit_display_cols(le);
    int max_len = 0;
    for (int i = 0; i < count; i++) {
        int len = (int)strlen(completions[i]);
        if (len > max_len) max_len = len;
    }
    
    int col_width = max_len + 2;
    int per_line = (cols > col_width) ? (cols / col_width) : 1;
    
    for (int i = 0; i < count; i++) {
        if (i > 0 && (i % per_line) == 0) {
            terminal_write("\r\n", 2);
        }
        terminal_write(completions[i], strlen(completions[i]));
        int pad = col_width - (int)strlen(completions[i]);
        for (int j = 0; j < pad; j++) {
            terminal_write_char(' ');
        }
    }
    terminal_write("\r\n", 2);
    
    free_completions(completions, count);
    
    lineedit_display(le);
}

char *lineedit_read(LineEdit *le) {
    le->len = 0;
    le->cursor = 0;
    le->buf[0] = '\0';
    le->interrupted = 0;
    free(le->saved_line);
    le->saved_line = NULL;
    
    if (le->history) {
        history_reset_pos(le->history);
    }
    
    terminal_enable_raw_mode();
    g_current_le = le;
    terminal_set_sigint_callback(lineedit_sigint_callback);
    
    lineedit_display(le);
    
    while (1) {
        int c = terminal_read_key();
        
        if (le->interrupted) {
            le->interrupted = 0;
            terminal_write("^C\r\n", 4);
            g_current_le = NULL;
            terminal_disable_raw_mode();
            return NULL;
        }
        
        if (c == '\r' || c == '\n') {
            terminal_write("\r\n", 2);
            break;
        }
        
        switch (c) {
            case CTRL_C:
                terminal_write("^C\r\n", 4);
                g_current_le = NULL;
                terminal_disable_raw_mode();
                return NULL;
                
            case CTRL_D:
                if (le->len == 0) {
                    g_current_le = NULL;
                    terminal_disable_raw_mode();
                    return NULL;
                }
                lineedit_delete(le);
                break;
                
            case CTRL_A:
            case HOME_KEY:
                lineedit_move_home(le);
                break;
                
            case CTRL_E:
            case END_KEY:
                lineedit_move_end(le);
                break;
                
            case CTRL_B:
                lineedit_move_left(le);
                break;
                
            case CTRL_F:
                lineedit_move_right(le);
                break;
                
            case CTRL_U:
                lineedit_kill_before(le);
                break;
                
            case CTRL_K:
                lineedit_kill_after(le);
                break;
                
            case BACKSPACE:
            case '\b':
                lineedit_backspace(le);
                break;
                
            case DELETE_KEY:
                lineedit_delete(le);
                break;
                
            case ARROW_LEFT:
                lineedit_move_left(le);
                break;
                
            case ARROW_RIGHT:
                lineedit_move_right(le);
                break;
                
            case ARROW_UP:
                lineedit_history_prev(le);
                break;
                
            case ARROW_DOWN:
                lineedit_history_next(le);
                break;
                
            case TAB_KEY:
                lineedit_complete(le);
                break;
                
            default:
                if (c >= 32 && c < 127) {
                    lineedit_insert_char(le, (char)c);
                }
                break;
        }
    }
    
    g_current_le = NULL;
    terminal_disable_raw_mode();
    
    if (le->len > 0) {
        if (le->history) {
            history_add(le->history, le->buf);
        }
        return strdup(le->buf);
    }
    
    return strdup("");
}
