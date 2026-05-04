#include "conn_pool.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <signal.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>

#define TEST_HOST "127.0.0.1"
#define TEST_PORT 55555
#define NUM_WORKER_THREADS 10
#define OPERATIONS_PER_THREAD 20
#define MAX_CLIENTS 100

static int server_running = 1;
static int server_sockfd = -1;

static int client_fds[MAX_CLIENTS];
static int client_count = 0;

void* mock_server_thread(void *arg)
{
    int port = *(int*)arg;
    int sockfd;
    struct sockaddr_in addr;
    int i;

    for (i = 0; i < MAX_CLIENTS; i++) {
        client_fds[i] = -1;
    }
    client_count = 0;

    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        perror("socket");
        return NULL;
    }

    int opt = 1;
    setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = inet_addr("127.0.0.1");
    addr.sin_port = htons(port);

    if (bind(sockfd, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
        perror("bind");
        close(sockfd);
        return NULL;
    }

    if (listen(sockfd, 10) < 0) {
        perror("listen");
        close(sockfd);
        return NULL;
    }

    server_sockfd = sockfd;
    printf("[Mock Server] Listening on 127.0.0.1:%d (keeping connections alive)\n", port);

    while (server_running) {
        struct timeval tv;
        fd_set readfds;
        int max_fd = sockfd;

        FD_ZERO(&readfds);
        FD_SET(sockfd, &readfds);

        for (i = 0; i < MAX_CLIENTS; i++) {
            if (client_fds[i] != -1) {
                FD_SET(client_fds[i], &readfds);
                if (client_fds[i] > max_fd) {
                    max_fd = client_fds[i];
                }
            }
        }

        tv.tv_sec = 1;
        tv.tv_usec = 0;

        int ret = select(max_fd + 1, &readfds, NULL, NULL, &tv);
        if (ret < 0) {
            if (errno == EINTR) continue;
            perror("select");
            break;
        }
        if (ret == 0) continue;

        if (FD_ISSET(sockfd, &readfds)) {
            struct sockaddr_in client_addr;
            socklen_t client_len = sizeof(client_addr);
            int client_fd = accept(sockfd, (struct sockaddr*)&client_addr, &client_len);

            if (client_fd < 0) {
                perror("accept");
                continue;
            }

            int added = 0;
            for (i = 0; i < MAX_CLIENTS; i++) {
                if (client_fds[i] == -1) {
                    client_fds[i] = client_fd;
                    client_count++;
                    added = 1;
                    break;
                }
            }

            if (!added) {
                printf("[Mock Server] Too many clients, rejecting connection\n");
                close(client_fd);
            } else {
                printf("[Mock Server] New connection accepted (fd=%d, total=%d)\n", 
                       client_fd, client_count);
            }
        }

        for (i = 0; i < MAX_CLIENTS; i++) {
            int fd = client_fds[i];
            if (fd != -1 && FD_ISSET(fd, &readfds)) {
                char buf[1024];
                ssize_t n = recv(fd, buf, sizeof(buf), 0);
                if (n <= 0) {
                    if (n < 0 && errno != EINTR) {
                        printf("[Mock Server] Connection closed by peer (fd=%d)\n", fd);
                    } else if (n == 0) {
                        printf("[Mock Server] Connection closed by peer (fd=%d)\n", fd);
                    }
                    close(fd);
                    client_fds[i] = -1;
                    client_count--;
                }
            }
        }
    }

    printf("[Mock Server] Cleaning up %d client connections...\n", client_count);
    for (i = 0; i < MAX_CLIENTS; i++) {
        if (client_fds[i] != -1) {
            close(client_fds[i]);
            client_fds[i] = -1;
        }
    }
    client_count = 0;

    close(sockfd);
    server_sockfd = -1;
    printf("[Mock Server] Stopped\n");
    return NULL;
}

static int success_count = 0;
static int timeout_count = 0;
static int total_conn_created = 0;
static pthread_mutex_t stat_mutex = PTHREAD_MUTEX_INITIALIZER;

typedef struct {
    ConnPool *pool;
    int thread_id;
} WorkerArg;

void* worker_thread(void *arg)
{
    WorkerArg *warg = (WorkerArg*)arg;
    ConnPool *pool = warg->pool;
    int tid = warg->thread_id;

    for (int i = 0; i < OPERATIONS_PER_THREAD; i++) {
        Connection *conn = conn_pool_acquire(pool);
        if (conn == NULL) {
            printf("[Thread %d] Connection acquire TIMEOUT!\n", tid);
            pthread_mutex_lock(&stat_mutex);
            timeout_count++;
            pthread_mutex_unlock(&stat_mutex);
            continue;
        }

        int sock = connection_socket(conn);
        printf("[Thread %d] Acquired connection (sockfd=%d)\n", tid, sock);

        usleep(100000);

        printf("[Thread %d] Releasing connection (sockfd=%d)\n", tid, sock);
        conn_pool_release(pool, conn);

        pthread_mutex_lock(&stat_mutex);
        success_count++;
        pthread_mutex_unlock(&stat_mutex);

        usleep(50000);
    }

    printf("[Thread %d] Finished all operations\n", tid);
    return NULL;
}

void print_pool_stats(const char *msg, ConnPool *pool)
{
    printf("\n=== %s ===\n", msg);
    printf("  Total connections: %d\n", pool->total_connections);
    printf("  Idle connections:  %d\n", pool->idle_count);
    printf("  Busy connections:  %d\n", pool->busy_count);
    printf("  Min connections:   %d\n", pool->config.min_connections);
    printf("  Max connections:   %d\n", pool->config.max_connections);
    printf("========================================\n\n");
}

int main(int argc, char *argv[])
{
    int port = TEST_PORT;
    pthread_t server_tid;

    printf("Starting Mock Server (keeping connections alive)...\n");
    pthread_create(&server_tid, NULL, mock_server_thread, &port);
    sleep(1);

    printf("\n=== Connection Pool Demo ===\n\n");

    ConnPoolConfig config;
    memset(&config, 0, sizeof(config));
    config.host = TEST_HOST;
    config.port = port;
    config.min_connections = 2;
    config.max_connections = 5;
    config.idle_timeout_sec = 30;
    config.acquire_timeout_ms = 3000;
    config.enable_health_check = true;

    printf("Creating connection pool with:\n");
    printf("  Host: %s:%d\n", config.host, config.port);
    printf("  Min connections: %d\n", config.min_connections);
    printf("  Max connections: %d\n", config.max_connections);
    printf("  Idle timeout: %d seconds\n", config.idle_timeout_sec);
    printf("  Acquire timeout: %d ms\n", config.acquire_timeout_ms);
    printf("  Health check: enabled\n");
    printf("\n");

    ConnPool *pool = conn_pool_create(&config);
    if (!pool) {
        fprintf(stderr, "Failed to create connection pool!\n");
        server_running = 0;
        pthread_join(server_tid, NULL);
        return 1;
    }

    printf("[NOTE] Check if Mock Server shows only %d 'New connection accepted' messages\n",
           config.min_connections);
    printf("[NOTE] Then we'll see if connections are REUSED (same sockfd numbers) during test\n\n");

    print_pool_stats("Pool after creation", pool);

    printf("Starting %d worker threads, each performing %d operations...\n",
           NUM_WORKER_THREADS, OPERATIONS_PER_THREAD);
    printf("Total operations: %d, max connections: %d\n\n",
           NUM_WORKER_THREADS * OPERATIONS_PER_THREAD,
           config.max_connections);

    pthread_t workers[NUM_WORKER_THREADS];
    WorkerArg args[NUM_WORKER_THREADS];

    for (int i = 0; i < NUM_WORKER_THREADS; i++) {
        args[i].pool = pool;
        args[i].thread_id = i;
        pthread_create(&workers[i], NULL, worker_thread, &args[i]);
    }

    for (int i = 0; i < NUM_WORKER_THREADS; i++) {
        pthread_join(workers[i], NULL);
    }

    printf("\nAll worker threads finished.\n\n");

    print_pool_stats("Pool after all operations", pool);

    printf("=== Test Statistics ===\n");
    printf("  Successful acquisitions: %d\n", success_count);
    printf("  Timeout acquisitions:    %d\n", timeout_count);
    printf("  Total:                   %d\n", success_count + timeout_count);
    printf("  Success rate:            %.1f%%\n", 
           (success_count + timeout_count > 0) ? 
           (100.0 * success_count / (success_count + timeout_count)) : 0.0);
    printf("========================\n");
    printf("\n[VERIFICATION CHECK]\n");
    printf("  - If connections are being REUSED: Mock Server should show only ~5 'New connection accepted'\n");
    printf("  - If NOT reused (buggy): Mock Server would show 200+ 'New connection accepted'\n");
    printf("========================\n\n");

    printf("Destroying connection pool...\n");
    conn_pool_destroy(pool);
    printf("Pool destroyed.\n");

    printf("Stopping mock server...\n");
    server_running = 0;
    pthread_join(server_tid, NULL);

    printf("\n=== Demo Complete ===\n");
    return 0;
}
