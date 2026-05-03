#define _GNU_SOURCE
#define _POSIX_C_SOURCE 200809L

#include "shutdown.h"
#include <pthread.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/time.h>
#include <errno.h>
#include <signal.h>

static ShutdownConfig g_shutdown_config;
static volatile sig_atomic_t g_shutdown_in_progress = 0;
static volatile sig_atomic_t g_shutdown_completed = 0;
static pthread_mutex_t g_shutdown_mutex = PTHREAD_MUTEX_INITIALIZER;
static uint64_t g_active_request_count = 0;

static uint64_t get_current_time_ms(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000 + (uint64_t)tv.tv_usec / 1000;
}

void shutdown_init(const ShutdownConfig *config) {
    memset(&g_shutdown_config, 0, sizeof(g_shutdown_config));
    
    if (config != NULL) {
        g_shutdown_config = *config;
    }
    
    if (g_shutdown_config.timeout_sec == 0) {
        g_shutdown_config.timeout_sec = DEFAULT_SHUTDOWN_TIMEOUT_SEC;
    }
    
    g_shutdown_in_progress = 0;
    g_shutdown_completed = 0;
}

void shutdown_cleanup(void) {
    pthread_mutex_destroy(&g_shutdown_mutex);
    memset(&g_shutdown_config, 0, sizeof(g_shutdown_config));
}

bool shutdown_is_in_progress(void) {
    return g_shutdown_in_progress != 0;
}

bool shutdown_is_completed(void) {
    return g_shutdown_completed != 0;
}

void shutdown_set_timeout(uint32_t timeout_sec) {
    g_shutdown_config.timeout_sec = timeout_sec;
}

uint32_t shutdown_get_timeout(void) {
    return g_shutdown_config.timeout_sec;
}

ShutdownResult shutdown_execute(void) {
    int lock_result = pthread_mutex_trylock(&g_shutdown_mutex);
    
    if (lock_result != 0) {
        return SHUTDOWN_ALREADY_IN_PROGRESS;
    }
    
    if (g_shutdown_completed) {
        pthread_mutex_unlock(&g_shutdown_mutex);
        return SHUTDOWN_ALREADY_IN_PROGRESS;
    }
    
    g_shutdown_in_progress = 1;
    
    uint64_t start_time = get_current_time_ms();
    uint64_t timeout_ms = (uint64_t)g_shutdown_config.timeout_sec * 1000;
    ShutdownResult result = SHUTDOWN_OK;
    
    fprintf(stderr, "[INFO] Starting graceful shutdown...\n");
    fflush(stderr);
    
    if (g_shutdown_config.stop_accepting != NULL) {
        fprintf(stderr, "[INFO] Step 1: Stopping to accept new requests...\n");
        fflush(stderr);
        g_shutdown_config.stop_accepting(g_shutdown_config.user_data);
    }
    
    fprintf(stderr, "[INFO] Step 2: Waiting for active requests to complete (timeout: %u sec)...\n",
            g_shutdown_config.timeout_sec);
    fflush(stderr);
    
    while (g_active_request_count > 0) {
        uint64_t elapsed = get_current_time_ms() - start_time;
        if (elapsed >= timeout_ms) {
            fprintf(stderr, "[WARNING] Shutdown timeout! Forcing exit with %lu active requests.\n",
                    (unsigned long)g_active_request_count);
            fflush(stderr);
            result = SHUTDOWN_TIMEOUT;
            break;
        }
        
        if (g_shutdown_config.wait_requests_complete != NULL) {
            g_shutdown_config.wait_requests_complete(g_shutdown_config.user_data);
        }
        
        usleep(100000);
    }
    
    if (result == SHUTDOWN_OK) {
        fprintf(stderr, "[INFO] All requests completed. Continuing shutdown...\n");
        fflush(stderr);
    }
    
    if (g_shutdown_config.flush_logs != NULL) {
        fprintf(stderr, "[INFO] Step 3: Flushing log buffers...\n");
        fflush(stderr);
        g_shutdown_config.flush_logs(g_shutdown_config.user_data);
    }
    
    if (g_shutdown_config.close_sockets != NULL) {
        fprintf(stderr, "[INFO] Step 4: Closing network sockets...\n");
        fflush(stderr);
        g_shutdown_config.close_sockets(g_shutdown_config.user_data);
    }
    
    if (g_shutdown_config.free_resources != NULL) {
        fprintf(stderr, "[INFO] Step 5: Releasing resources...\n");
        fflush(stderr);
        g_shutdown_config.free_resources(g_shutdown_config.user_data);
    }
    
    fprintf(stderr, "[INFO] Step 6: Shutdown complete. Result: %s\n",
            result == SHUTDOWN_OK ? "SUCCESS" : "TIMEOUT/FORCED");
    fflush(stderr);
    
    g_shutdown_completed = 1;
    g_shutdown_in_progress = 0;
    pthread_mutex_unlock(&g_shutdown_mutex);
    
    return result;
}

void shutdown_increment_active_requests(void) {
    pthread_mutex_lock(&g_shutdown_mutex);
    g_active_request_count++;
    pthread_mutex_unlock(&g_shutdown_mutex);
}

void shutdown_decrement_active_requests(void) {
    pthread_mutex_lock(&g_shutdown_mutex);
    if (g_active_request_count > 0) {
        g_active_request_count--;
    }
    pthread_mutex_unlock(&g_shutdown_mutex);
}

uint64_t shutdown_get_active_requests(void) {
    uint64_t count;
    pthread_mutex_lock(&g_shutdown_mutex);
    count = g_active_request_count;
    pthread_mutex_unlock(&g_shutdown_mutex);
    return count;
}
