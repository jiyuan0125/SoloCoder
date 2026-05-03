#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <signal.h>
#include "timeout_checker.h"
#include "client_manager.h"
#include "heartbeat_protocol.h"

static pthread_t g_checker_thread;
static int g_is_running = 0;
static int g_scan_interval = HEARTBEAT_SCAN_INTERVAL_SEC;
static int g_timeout_sec = HEARTBEAT_TIMEOUT_SEC;
static timeout_notify_callback_t g_notify_callback = NULL;

static void *timeout_checker_thread(void *arg) {
    (void)arg;
    
    printf("Timeout checker thread started (scan: %d sec, timeout: %d sec)\n",
           g_scan_interval, g_timeout_sec);
    
    while (g_is_running) {
        sleep(g_scan_interval);
        
        if (!g_is_running) {
            break;
        }
        
        time_t now = time(NULL);
        int timeout_count = client_manager_check_timeouts(now, g_timeout_sec);
        
        if (timeout_count > 0) {
            printf("Timeout checker: removed %d timed out clients\n", timeout_count);
            
            if (g_notify_callback != NULL) {
                g_notify_callback();
            }
        }
    }
    
    printf("Timeout checker thread stopped\n");
    return NULL;
}

int timeout_checker_init(int scan_interval_sec, int timeout_sec) {
    if (scan_interval_sec <= 0 || timeout_sec <= 0) {
        fprintf(stderr, "Invalid timeout parameters\n");
        return -1;
    }
    
    g_scan_interval = scan_interval_sec;
    g_timeout_sec = timeout_sec;
    g_is_running = 0;
    g_notify_callback = NULL;
    
    return 0;
}

void timeout_checker_cleanup(void) {
    if (g_is_running) {
        timeout_checker_stop();
    }
    g_notify_callback = NULL;
}

int timeout_checker_start(void) {
    if (g_is_running) {
        printf("Timeout checker is already running\n");
        return 0;
    }
    
    g_is_running = 1;
    
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);
    
    int ret = pthread_create(&g_checker_thread, &attr, timeout_checker_thread, NULL);
    pthread_attr_destroy(&attr);
    
    if (ret != 0) {
        fprintf(stderr, "Failed to create timeout checker thread: %d\n", ret);
        g_is_running = 0;
        return -1;
    }
    
    return 0;
}

int timeout_checker_stop(void) {
    if (!g_is_running) {
        return 0;
    }
    
    g_is_running = 0;
    
    int ret = pthread_join(g_checker_thread, NULL);
    if (ret != 0) {
        fprintf(stderr, "Failed to join timeout checker thread: %d\n", ret);
        return -1;
    }
    
    return 0;
}

void timeout_checker_set_notify_callback(timeout_notify_callback_t callback) {
    g_notify_callback = callback;
}

int timeout_checker_is_running(void) {
    return g_is_running;
}
