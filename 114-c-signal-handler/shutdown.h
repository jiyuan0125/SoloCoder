#ifndef SHUTDOWN_H
#define SHUTDOWN_H

#include <stdbool.h>
#include <stdint.h>

#define DEFAULT_SHUTDOWN_TIMEOUT_SEC 30

typedef enum {
    SHUTDOWN_OK = 0,
    SHUTDOWN_TIMEOUT = -1,
    SHUTDOWN_FORCED = -2,
    SHUTDOWN_ALREADY_IN_PROGRESS = -3
} ShutdownResult;

typedef void (*shutdown_callback_fn)(void *user_data);

typedef struct {
    shutdown_callback_fn stop_accepting;
    shutdown_callback_fn wait_requests_complete;
    shutdown_callback_fn flush_logs;
    shutdown_callback_fn close_sockets;
    shutdown_callback_fn free_resources;
    void *user_data;
    uint32_t timeout_sec;
} ShutdownConfig;

void shutdown_init(const ShutdownConfig *config);
void shutdown_cleanup(void);

ShutdownResult shutdown_execute(void);
bool shutdown_is_in_progress(void);
bool shutdown_is_completed(void);

void shutdown_set_timeout(uint32_t timeout_sec);
uint32_t shutdown_get_timeout(void);

void shutdown_increment_active_requests(void);
void shutdown_decrement_active_requests(void);
uint64_t shutdown_get_active_requests(void);

#endif
