#include "conn_pool.h"
#include "conn_health.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <time.h>

static Connection* create_connection(ConnPool *pool)
{
    Connection *conn = (Connection*)malloc(sizeof(Connection));
    if (!conn) {
        return NULL;
    }

    conn->sockfd = create_tcp_connection(pool->config.host, pool->config.port);
    if (conn->sockfd < 0) {
        free(conn);
        return NULL;
    }

    gettimeofday(&conn->last_used, NULL);
    conn->next = NULL;

    return conn;
}

static void destroy_all_connections(ConnPool *pool)
{
    Connection *curr, *next;

    curr = pool->idle_list;
    while (curr != NULL) {
        next = curr->next;
        close_connection(curr);
        free(curr);
        curr = next;
    }
    pool->idle_list = NULL;

    curr = pool->busy_list;
    while (curr != NULL) {
        next = curr->next;
        close_connection(curr);
        free(curr);
        curr = next;
    }
    pool->busy_list = NULL;

    pool->total_connections = 0;
    pool->idle_count = 0;
    pool->busy_count = 0;
}

ConnPool* conn_pool_create(const ConnPoolConfig *config)
{
    if (!config) {
        return NULL;
    }

    if (config->min_connections < 0 ||
        config->max_connections <= 0 ||
        config->min_connections > config->max_connections) {
        return NULL;
    }

    ConnPool *pool = (ConnPool*)malloc(sizeof(ConnPool));
    if (!pool) {
        return NULL;
    }

    memset(pool, 0, sizeof(ConnPool));

    pool->config = *config;

    if (pthread_mutex_init(&pool->mutex, NULL) != 0) {
        free(pool);
        return NULL;
    }

    if (pthread_cond_init(&pool->cond, NULL) != 0) {
        pthread_mutex_destroy(&pool->mutex);
        free(pool);
        return NULL;
    }

    for (int i = 0; i < config->min_connections; i++) {
        Connection *conn = create_connection(pool);
        if (conn) {
            conn->next = pool->idle_list;
            pool->idle_list = conn;
            pool->idle_count++;
            pool->total_connections++;
        }
    }

    if (config->idle_timeout_sec > 0) {
        start_reaper_thread(pool);
    }

    return pool;
}

void conn_pool_destroy(ConnPool *pool)
{
    if (!pool) {
        return;
    }

    if (pool->config.idle_timeout_sec > 0) {
        stop_reaper_thread(pool);
    }

    pthread_mutex_lock(&pool->mutex);

    destroy_all_connections(pool);

    pthread_mutex_unlock(&pool->mutex);

    pthread_cond_destroy(&pool->cond);
    pthread_mutex_destroy(&pool->mutex);

    free(pool);
}

static Connection* pop_idle_connection(ConnPool *pool)
{
    if (!pool->idle_list) {
        return NULL;
    }

    Connection *conn = pool->idle_list;
    pool->idle_list = conn->next;
    conn->next = NULL;

    pool->idle_count--;

    return conn;
}

static void push_busy_connection(ConnPool *pool, Connection *conn)
{
    conn->next = pool->busy_list;
    pool->busy_list = conn;
    pool->busy_count++;
}

static Connection* get_valid_idle_connection(ConnPool *pool)
{
    while (pool->idle_list != NULL) {
        Connection *conn = pop_idle_connection(pool);

        if (pool->config.enable_health_check) {
            if (!health_check_connection(conn)) {
                close_connection(conn);
                free(conn);
                pool->total_connections--;
                continue;
            }
        }

        return conn;
    }

    return NULL;
}

Connection* conn_pool_acquire(ConnPool *pool)
{
    if (!pool) {
        return NULL;
    }

    pthread_mutex_lock(&pool->mutex);

    Connection *conn = NULL;
    struct timespec abs_timeout;

    while (1) {
        conn = get_valid_idle_connection(pool);
        if (conn) {
            break;
        }

        if (pool->total_connections < pool->config.max_connections) {
            conn = create_connection(pool);
            if (conn) {
                pool->total_connections++;
                break;
            }
        }

        if (pool->config.acquire_timeout_ms <= 0) {
            pthread_cond_wait(&pool->cond, &pool->mutex);
        } else {
            struct timeval now;
            gettimeofday(&now, NULL);

            abs_timeout.tv_sec = now.tv_sec + pool->config.acquire_timeout_ms / 1000;
            abs_timeout.tv_nsec = now.tv_usec * 1000 + 
                                   (pool->config.acquire_timeout_ms % 1000) * 1000000;

            if (abs_timeout.tv_nsec >= 1000000000) {
                abs_timeout.tv_sec += abs_timeout.tv_nsec / 1000000000;
                abs_timeout.tv_nsec = abs_timeout.tv_nsec % 1000000000;
            }

            int ret = pthread_cond_timedwait(&pool->cond, &pool->mutex, &abs_timeout);
            if (ret == ETIMEDOUT) {
                pthread_mutex_unlock(&pool->mutex);
                return NULL;
            }
        }
    }

    push_busy_connection(pool, conn);
    gettimeofday(&conn->last_used, NULL);

    pthread_mutex_unlock(&pool->mutex);

    return conn;
}

static void remove_from_busy_list(ConnPool *pool, Connection *conn)
{
    Connection **pp = &pool->busy_list;
    while (*pp != NULL) {
        if (*pp == conn) {
            *pp = conn->next;
            pool->busy_count--;
            return;
        }
        pp = &(*pp)->next;
    }
}

void conn_pool_release(ConnPool *pool, Connection *conn)
{
    if (!pool || !conn) {
        return;
    }

    cleanup_connection(conn);

    pthread_mutex_lock(&pool->mutex);

    remove_from_busy_list(pool, conn);

    conn->next = pool->idle_list;
    pool->idle_list = conn;
    pool->idle_count++;

    gettimeofday(&conn->last_used, NULL);

    pthread_cond_signal(&pool->cond);

    pthread_mutex_unlock(&pool->mutex);
}

int connection_socket(const Connection *conn)
{
    return conn ? conn->sockfd : -1;
}
