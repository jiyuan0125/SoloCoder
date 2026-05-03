#ifndef COMMON_H
#define COMMON_H

#include <stddef.h>

typedef enum {
    TP_OK = 0,
    TP_ERROR = -1,
    TP_FULL = -2,
    TP_CLOSED = -3,
    TP_TIMEOUT = -4,
    TP_INVALID_ARG = -5
} tp_status_t;

typedef enum {
    TP_REJECT_BLOCK = 0,
    TP_REJECT_DISCARD
} tp_reject_policy_t;

typedef enum {
    TP_CALLBACK_IN_WORKER = 0,
    TP_CALLBACK_DEFERRED
} tp_callback_exec_t;

typedef enum {
    TP_GRACEFUL_SHUTDOWN = 0,
    TP_FORCE_SHUTDOWN
} tp_shutdown_mode_t;

typedef enum {
    TP_STOPPED = 0,
    TP_RUNNING,
    TP_SHUTDOWN
} tp_state_t;

typedef void (*tp_task_func_t)(void *arg);
typedef void (*tp_completion_func_t)(void *arg, int success);

typedef struct {
    tp_task_func_t func;
    void *arg;
    tp_completion_func_t completion;
    tp_callback_exec_t callback_mode;
} tp_task_t;

typedef struct {
    void *arg;
    tp_completion_func_t completion;
    int success;
} tp_completion_t;

#define TP_TASK_INIT(func_, arg_, comp_, mode_) \
    { (func_), (arg_), (comp_), (mode_) }

#endif
