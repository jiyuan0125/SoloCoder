#include "terminal.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>
#include <sys/ioctl.h>

static volatile sig_atomic_t g_interrupted = 0;
static void (*g_old_sigint_handler)(int) = NULL;

static void handle_sigint(int sig) {
    g_interrupted = 1;
    terminal_cleanup();
    if (g_old_sigint_handler && g_old_sigint_handler != SIG_DFL && g_old_sigint_handler != SIG_IGN) {
        g_old_sigint_handler(sig);
    }
    raise(SIGINT);
}

void terminal_init(void) {
    g_interrupted = 0;
    if (terminal_is_supported()) {
        g_old_sigint_handler = signal(SIGINT, handle_sigint);
        terminal_hide_cursor();
    }
}

void terminal_cleanup(void) {
    if (terminal_is_supported()) {
        terminal_show_cursor();
        terminal_clear_line();
        fprintf(stdout, "\n");
        fflush(stdout);
        if (g_old_sigint_handler) {
            signal(SIGINT, g_old_sigint_handler);
            g_old_sigint_handler = NULL;
        }
    }
}

int terminal_is_supported(void) {
    return isatty(STDOUT_FILENO);
}

int terminal_get_width(void) {
    struct winsize ws;
    if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &ws) == 0) {
        return ws.ws_col;
    }
    return 80;
}

void terminal_clear_line(void) {
    if (terminal_is_supported()) {
        fprintf(stdout, "\033[2K");
    }
}

void terminal_move_up(int lines) {
    if (terminal_is_supported() && lines > 0) {
        fprintf(stdout, "\033[%dA", lines);
    }
}

void terminal_move_down(int lines) {
    if (terminal_is_supported() && lines > 0) {
        fprintf(stdout, "\033[%dB", lines);
    }
}

void terminal_hide_cursor(void) {
    if (terminal_is_supported()) {
        fprintf(stdout, "\033[?25l");
        fflush(stdout);
    }
}

void terminal_show_cursor(void) {
    if (terminal_is_supported()) {
        fprintf(stdout, "\033[?25h");
        fflush(stdout);
    }
}

void bytes_format_unit(uint64_t bytes, char *buf, size_t buf_size) {
    const char *units[] = {"B", "KB", "MB", "GB", "TB"};
    int unit_index = 0;
    double size = (double)bytes;
    
    while (size >= 1024.0 && unit_index < 4) {
        size /= 1024.0;
        unit_index++;
    }
    
    if (size < 10.0 && unit_index > 0) {
        snprintf(buf, buf_size, "%.2f %s", size, units[unit_index]);
    } else if (size < 100.0 && unit_index > 0) {
        snprintf(buf, buf_size, "%.1f %s", size, units[unit_index]);
    } else {
        snprintf(buf, buf_size, "%.0f %s", size, units[unit_index]);
    }
}

void terminal_draw_progress(const ProgressState *state, int index, int total,
                             int terminal_supported, int terminal_width) {
    char current_str[32], total_str[32];
    char speed_str[32], time_str[32];
    double percent = progress_state_get_percent(state);
    double speed = speed_calculator_get_average(&state->speed_calc);
    double remaining = progress_state_get_remaining_seconds(state);
    
    bytes_format_unit(state->current_bytes, current_str, sizeof(current_str));
    bytes_format_unit(state->total_bytes, total_str, sizeof(total_str));
    
    if (speed > 0) {
        speed_format_unit(speed, speed_str, sizeof(speed_str));
    } else {
        snprintf(speed_str, sizeof(speed_str), "-- B/s");
    }
    
    time_format_remaining(remaining, time_str, sizeof(time_str));
    
    if (!terminal_supported || terminal_width < MIN_TERMINAL_WIDTH) {
        if (total > 1) {
            fprintf(stdout, "[%d] %.1f%% | %s/%s | %s | ETA %s\n",
                    index + 1, percent, current_str, total_str, speed_str, time_str);
        } else {
            fprintf(stdout, "%.1f%% | %s/%s | %s | ETA %s\n",
                    percent, current_str, total_str, speed_str, time_str);
        }
        fflush(stdout);
        return;
    }
    
    int bar_width = state->bar_width;
    int max_content_width = terminal_width - 2;
    
    char bar_buffer[256];
    int bar_filled = (int)(percent / 100.0 * bar_width);
    
    bar_buffer[0] = '[';
    int pos = 1;
    
    for (int i = 0; i < bar_filled - 1 && i < bar_width - 1; i++) {
        bar_buffer[pos++] = '=';
    }
    
    if (bar_filled > 0 && bar_filled <= bar_width) {
        bar_buffer[pos++] = (percent >= 100.0) ? '=' : '>';
    }
    
    for (int i = bar_filled; i < bar_width; i++) {
        bar_buffer[pos++] = ' ';
    }
    
    bar_buffer[pos++] = ']';
    bar_buffer[pos] = '\0';
    
    char info_buffer[256];
    if (total > 1) {
        snprintf(info_buffer, sizeof(info_buffer), "[%d] %.1f%% %s %s/%s | %s | ETA %s",
                 index + 1, percent, bar_buffer, current_str, total_str, speed_str, time_str);
    } else {
        snprintf(info_buffer, sizeof(info_buffer), "%.1f%% %s %s/%s | %s | ETA %s",
                 percent, bar_buffer, current_str, total_str, speed_str, time_str);
    }
    
    int info_len = strlen(info_buffer);
    if (info_len > max_content_width) {
        info_buffer[max_content_width] = '\0';
    }
    
    fprintf(stdout, "\r%s", info_buffer);
    fflush(stdout);
}

void terminal_draw_all(const ProgressManager *manager) {
    if (manager->count == 0) return;
    
    static int first_draw = 1;
    
    if (!first_draw && manager->terminal_supported && manager->count > 1) {
        terminal_move_up(manager->count);
    }
    
    for (size_t i = 0; i < manager->count; i++) {
        if (i > 0 && manager->terminal_supported) {
            fprintf(stdout, "\n");
        }
        terminal_draw_progress(&manager->states[i], (int)i, (int)manager->count,
                               manager->terminal_supported, manager->terminal_width);
    }
    
    first_draw = 0;
}
