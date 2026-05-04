#include "terminal.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <signal.h>
#include <sys/ioctl.h>
#include <string.h>
#include <errno.h>

static TerminalState g_terminal;
static void (*g_sigint_callback)(void) = NULL;

void terminal_sigint_handler(int sig) {
    (void)sig;
    if (g_sigint_callback) {
        g_sigint_callback();
    }
}

void terminal_set_sigint_callback(void (*callback)(void)) {
    g_sigint_callback = callback;
}

static void handle_exit(void) {
    terminal_disable_raw_mode();
}

static void handle_sigterm(int sig) {
    (void)sig;
    terminal_disable_raw_mode();
    exit(128 + sig);
}

void terminal_enable_raw_mode(void) {
    if (g_terminal.is_raw) return;
    
    if (tcgetattr(STDIN_FILENO, &g_terminal.orig_termios) == -1) {
        perror("tcgetattr");
        exit(1);
    }
    
    atexit(handle_exit);
    
    struct termios raw = g_terminal.orig_termios;
    
    raw.c_iflag &= ~(BRKINT | ICRNL | INPCK | ISTRIP | IXON);
    raw.c_oflag &= ~(OPOST);
    raw.c_cflag |= (CS8);
    raw.c_lflag &= ~(ECHO | ICANON | IEXTEN | ISIG);
    raw.c_cc[VMIN] = 1;
    raw.c_cc[VTIME] = 0;
    
    if (tcsetattr(STDIN_FILENO, TCSAFLUSH, &raw) == -1) {
        perror("tcsetattr");
        exit(1);
    }
    
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = handle_sigterm;
    sigemptyset(&sa.sa_mask);
    if (sigaction(SIGTERM, &sa, NULL) == -1) {
        perror("sigaction SIGTERM");
    }
    if (sigaction(SIGHUP, &sa, NULL) == -1) {
        perror("sigaction SIGHUP");
    }
    
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = terminal_sigint_handler;
    sa.sa_flags = SA_RESTART;
    sigemptyset(&sa.sa_mask);
    if (sigaction(SIGINT, &sa, NULL) == -1) {
        perror("sigaction SIGINT");
    }
    
    g_terminal.is_raw = 1;
}

void terminal_disable_raw_mode(void) {
    if (!g_terminal.is_raw) return;
    
    if (tcsetattr(STDIN_FILENO, TCSAFLUSH, &g_terminal.orig_termios) == -1) {
        perror("tcsetattr");
    }
    g_terminal.is_raw = 0;
}

int terminal_get_width(void) {
    struct winsize ws;
    if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &ws) == -1 || ws.ws_col == 0) {
        return 80;
    }
    return ws.ws_col;
}

int terminal_get_cursor_row(void) {
    char buf[32];
    int i = 0;
    
    if (write(STDOUT_FILENO, "\x1b[6n", 4) != 4) {
        return -1;
    }
    
    while (i < (int)sizeof(buf) - 1) {
        if (read(STDIN_FILENO, &buf[i], 1) != 1) break;
        if (buf[i] == 'R') break;
        i++;
    }
    buf[i] = '\0';
    
    int row = 0;
    if (buf[0] == '\x1b' && buf[1] == '[') {
        sscanf(&buf[2], "%d", &row);
    }
    
    return row;
}

void terminal_move_cursor(int cols) {
    if (cols > 0) {
        char seq[16];
        int n = snprintf(seq, sizeof(seq), "\x1b[%dC", cols);
        write(STDOUT_FILENO, seq, n);
    } else if (cols < 0) {
        char seq[16];
        int n = snprintf(seq, sizeof(seq), "\x1b[%dD", -cols);
        write(STDOUT_FILENO, seq, n);
    }
}

void terminal_move_cursor_to(int col) {
    char seq[16];
    int n = snprintf(seq, sizeof(seq), "\x1b[%dG", col + 1);
    write(STDOUT_FILENO, seq, n);
}

void terminal_clear_from_cursor(void) {
    write(STDOUT_FILENO, "\x1b[0J", 4);
}

void terminal_clear_screen(void) {
    write(STDOUT_FILENO, "\x1b[2J\x1b[H", 7);
}

void terminal_write(const char *s, size_t len) {
    write(STDOUT_FILENO, s, len);
}

void terminal_write_char(char c) {
    write(STDOUT_FILENO, &c, 1);
}

int terminal_read_key(void) {
    int nread;
    char c;
    
    while ((nread = read(STDIN_FILENO, &c, 1)) != 1) {
        if (nread == -1 && errno != EAGAIN) {
            return -1;
        }
    }
    
    if (c == '\x1b') {
        char seq[3];
        
        if (read(STDIN_FILENO, &seq[0], 1) != 1) return '\x1b';
        if (read(STDIN_FILENO, &seq[1], 1) != 1) return '\x1b';
        
        if (seq[0] == '[') {
            switch (seq[1]) {
                case 'A': return ARROW_UP;
                case 'B': return ARROW_DOWN;
                case 'C': return ARROW_RIGHT;
                case 'D': return ARROW_LEFT;
                case 'F': return END_KEY;
                case 'H': return HOME_KEY;
            }
            
            if (seq[1] >= '0' && seq[1] <= '9') {
                char seq2;
                if (read(STDIN_FILENO, &seq2, 1) != 1) return '\x1b';
                if (seq2 == '~') {
                    switch (seq[1]) {
                        case '1': return HOME_KEY;
                        case '3': return DELETE_KEY;
                        case '4': return END_KEY;
                        case '7': return HOME_KEY;
                        case '8': return END_KEY;
                    }
                }
            }
        }
        
        return '\x1b';
    }
    
    return c;
}
