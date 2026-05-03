#define _GNU_SOURCE
#define _POSIX_C_SOURCE 200809L

#include "signal_handler.h"
#include <signal.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <errno.h>

static volatile sig_atomic_t g_signal_flags = SIGNAL_FLAG_NONE;
static volatile sig_atomic_t g_stop_requested = 0;

static void handle_signal(int sig) {
    int saved_errno = errno;
    
    switch (sig) {
        case SIGTERM:
            g_signal_flags |= SIGNAL_FLAG_TERMINATE;
            g_stop_requested = 1;
            break;
        case SIGINT:
            g_signal_flags |= SIGNAL_FLAG_INTERRUPT;
            g_stop_requested = 1;
            break;
        case SIGHUP:
            g_signal_flags |= SIGNAL_FLAG_RELOAD;
            break;
        case SIGUSR1:
            g_signal_flags |= SIGNAL_FLAG_ROTATE_LOG;
            break;
        default:
            break;
    }
    
    errno = saved_errno;
}

void signal_handler_init(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = handle_signal;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART;
    
    sigaction(SIGTERM, &sa, NULL);
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGHUP, &sa, NULL);
    sigaction(SIGUSR1, &sa, NULL);
    
    signal(SIGPIPE, SIG_IGN);
}

void signal_handler_cleanup(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = SIG_DFL;
    sigemptyset(&sa.sa_mask);
    
    sigaction(SIGTERM, &sa, NULL);
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGHUP, &sa, NULL);
    sigaction(SIGUSR1, &sa, NULL);
}

bool signal_handler_is_flag_set(SignalFlags flag) {
    return (g_signal_flags & flag) != 0;
}

void signal_handler_clear_flag(SignalFlags flag) {
    g_signal_flags &= ~flag;
}

SignalFlags signal_handler_get_flags(void) {
    return (SignalFlags)g_signal_flags;
}

void signal_handler_set_stop_requested(bool requested) {
    g_stop_requested = requested ? 1 : 0;
}

bool signal_handler_is_stop_requested(void) {
    return g_stop_requested != 0;
}
