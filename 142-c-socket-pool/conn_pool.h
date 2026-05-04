#ifndef CONN_POOL_H
#define CONN_POOL_H

#include <pthread.h>
#include <sys/time.h>
#include <stdbool.h>

typedef struct Connection {
    int sockfd;
    struct timeval last_used;
    struct Connection *next;
} Connection;

typedef struct ConnPoolConfig {
    const char *host;
    int port;
    int min_connections;
    int max_connections;
    int idle_timeout_sec;
    int acquire_timeout_ms;
    bool enable_health_check;
} ConnPoolConfig;

typedef struct ConnPool {
    Connection *idle_list;
    Connection *busy_list;
    int total_connections;
    int idle_count;
    int busy_count;
    
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    
    ConnPoolConfig config;
    
    pthread_t reaper_thread;
    bool shutdown_flag;
} ConnPool;

ConnPool* conn_pool_create(const ConnPoolConfig *config);
void conn_pool_destroy(ConnPool *pool);

Connection* conn_pool_acquire(ConnPool *pool);
void conn_pool_release(ConnPool *pool, Connection *conn);

int connection_socket(const Connection *conn);

#endif
