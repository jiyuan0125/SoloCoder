#ifndef TIMEOUT_CHECKER_H
#define TIMEOUT_CHECKER_H

#include <pthread.h>

typedef void (*timeout_notify_callback_t)(void);

int timeout_checker_init(int scan_interval_sec, int timeout_sec);
void timeout_checker_cleanup(void);

int timeout_checker_start(void);
int timeout_checker_stop(void);

void timeout_checker_set_notify_callback(timeout_notify_callback_t callback);
int timeout_checker_is_running(void);

#endif
