#ifndef CONN_HEALTH_H
#define CONN_HEALTH_H

#include "conn_pool.h"
#include <stdbool.h>

bool health_check_connection(Connection *conn);
void cleanup_connection(Connection *conn);
bool is_connection_idle(const Connection *conn, int idle_timeout_sec);

void start_reaper_thread(ConnPool *pool);
void stop_reaper_thread(ConnPool *pool);
void* reaper_thread_func(void *arg);

int create_tcp_connection(const char *host, int port);
void close_connection(Connection *conn);

#endif
