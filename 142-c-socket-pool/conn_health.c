#include "conn_health.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <errno.h>
#include <fcntl.h>
#include <poll.h>

int create_tcp_connection(const char *host, int port)
{
    int sockfd;
    struct addrinfo hints, *res, *p;
    char port_str[16];
    int rv;

    memset(&hints, 0, sizeof hints);
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    snprintf(port_str, sizeof(port_str), "%d", port);

    if ((rv = getaddrinfo(host, port_str, &hints, &res)) != 0) {
        fprintf(stderr, "getaddrinfo: %s\n", gai_strerror(rv));
        return -1;
    }

    for (p = res; p != NULL; p = p->ai_next) {
        if ((sockfd = socket(p->ai_family, p->ai_socktype, p->ai_protocol)) == -1) {
            continue;
        }

        if (connect(sockfd, p->ai_addr, p->ai_addrlen) == -1) {
            close(sockfd);
            continue;
        }

        break;
    }

    freeaddrinfo(res);

    if (p == NULL) {
        return -1;
    }

    int flags = fcntl(sockfd, F_GETFL, 0);
    if (flags != -1) {
        fcntl(sockfd, F_SETFL, flags);
    }

    return sockfd;
}

void close_connection(Connection *conn)
{
    if (conn && conn->sockfd >= 0) {
        close(conn->sockfd);
        conn->sockfd = -1;
    }
}

bool health_check_connection(Connection *conn)
{
    if (!conn || conn->sockfd < 0) {
        return false;
    }

    struct pollfd pfd;
    pfd.fd = conn->sockfd;
    pfd.events = POLLIN;

    int ret = poll(&pfd, 1, 100);

    if (ret < 0) {
        return false;
    }

    if (ret == 0) {
        return true;
    }

    if (pfd.revents & POLLIN) {
        char buf[1];
        ssize_t n = recv(conn->sockfd, buf, sizeof(buf), MSG_PEEK);

        if (n <= 0) {
            return false;
        }
    }

    return true;
}

void cleanup_connection(Connection *conn)
{
    if (!conn || conn->sockfd < 0) {
        return;
    }

    struct pollfd pfd;
    pfd.fd = conn->sockfd;
    pfd.events = POLLIN;

    while (1) {
        int ret = poll(&pfd, 1, 0);
        if (ret <= 0) {
            break;
        }

        char buf[4096];
        ssize_t n = recv(conn->sockfd, buf, sizeof(buf), 0);
        if (n <= 0) {
            break;
        }
    }
}

bool is_connection_idle(const Connection *conn, int idle_timeout_sec)
{
    struct timeval now;
    gettimeofday(&now, NULL);

    long elapsed_sec = now.tv_sec - conn->last_used.tv_sec;
    if (elapsed_sec > idle_timeout_sec) {
        return true;
    }

    return false;
}

void start_reaper_thread(ConnPool *pool)
{
    pool->shutdown_flag = false;
    pthread_create(&pool->reaper_thread, NULL, reaper_thread_func, pool);
}

void stop_reaper_thread(ConnPool *pool)
{
    pool->shutdown_flag = true;
    pthread_join(pool->reaper_thread, NULL);
}

void* reaper_thread_func(void *arg)
{
    ConnPool *pool = (ConnPool*)arg;

    while (!pool->shutdown_flag) {
        sleep(10);

        if (pool->shutdown_flag) {
            break;
        }

        pthread_mutex_lock(&pool->mutex);

        Connection *prev = NULL;
        Connection *curr = pool->idle_list;

        while (curr != NULL && pool->total_connections > pool->config.min_connections) {
            if (is_connection_idle(curr, pool->config.idle_timeout_sec)) {
                Connection *to_remove = curr;

                if (prev == NULL) {
                    pool->idle_list = curr->next;
                } else {
                    prev->next = curr->next;
                }
                curr = curr->next;

                close_connection(to_remove);
                free(to_remove);

                pool->total_connections--;
                pool->idle_count--;
            } else {
                prev = curr;
                curr = curr->next;
            }
        }

        pthread_mutex_unlock(&pool->mutex);
    }

    return NULL;
}
