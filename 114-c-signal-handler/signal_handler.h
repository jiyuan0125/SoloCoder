#ifndef SIGNAL_HANDLER_H
#define SIGNAL_HANDLER_H

#include <signal.h>
#include <stdbool.h>

typedef enum {
    SIGNAL_FLAG_NONE = 0,
    SIGNAL_FLAG_TERMINATE = 1 << 0,
    SIGNAL_FLAG_INTERRUPT = 1 << 1,
    SIGNAL_FLAG_RELOAD = 1 << 2,
    SIGNAL_FLAG_ROTATE_LOG = 1 << 3
} SignalFlags;

void signal_handler_init(void);
void signal_handler_cleanup(void);

bool signal_handler_is_flag_set(SignalFlags flag);
void signal_handler_clear_flag(SignalFlags flag);
SignalFlags signal_handler_get_flags(void);
void signal_handler_set_stop_requested(bool requested);
bool signal_handler_is_stop_requested(void);

#endif
