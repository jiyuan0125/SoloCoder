#ifndef PIDFILE_H
#define PIDFILE_H

#include <sys/types.h>
#include <stdbool.h>

#define PIDFILE_DEFAULT_PATH "/var/run/daemon.pid"

typedef enum {
    PIDFILE_OK = 0,
    PIDFILE_ERROR_OPEN = -1,
    PIDFILE_ERROR_LOCK = -2,
    PIDFILE_ERROR_WRITE = -3,
    PIDFILE_ERROR_ALREADY_RUNNING = -4,
    PIDFILE_ERROR_INVALID_PATH = -5
} PidFileResult;

PidFileResult pidfile_create(const char *path);
PidFileResult pidfile_remove(void);
PidFileResult pidfile_check(const char *path, bool *is_running);

pid_t pidfile_get_current_pid(void);
const char *pidfile_get_current_path(void);

#endif
