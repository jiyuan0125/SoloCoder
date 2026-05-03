#define _GNU_SOURCE
#define _POSIX_C_SOURCE 200809L

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <pthread.h>
#include <errno.h>
#include <stdbool.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>

#include "signal_handler.h"
#include "shutdown.h"
#include "pidfile.h"

#define PIDFILE_PATH "/tmp/daemon_demo.pid"
#define SHUTDOWN_TIMEOUT_SEC 10

static volatile sig_atomic_t g_accepting_requests = 1;
static pthread_mutex_t g_request_mutex = PTHREAD_MUTEX_INITIALIZER;

typedef struct {
    int request_id;
    int process_time_ms;
} RequestInfo;

static void simulate_log_flush(void *user_data) {
    (void)user_data;
    fprintf(stderr, "  [DEMO] Flushing all log buffers...\n");
    fflush(stderr);
    usleep(200000);
    fprintf(stderr, "  [DEMO] Log buffers flushed.\n");
    fflush(stderr);
}

static void simulate_socket_close(void *user_data) {
    (void)user_data;
    fprintf(stderr, "  [DEMO] Closing listening sockets...\n");
    fflush(stderr);
    usleep(100000);
    fprintf(stderr, "  [DEMO] All sockets closed.\n");
    fflush(stderr);
}

static void simulate_free_resources(void *user_data) {
    (void)user_data;
    fprintf(stderr, "  [DEMO] Freeing allocated resources...\n");
    fflush(stderr);
    usleep(100000);
    fprintf(stderr, "  [DEMO] Resources released.\n");
    fflush(stderr);
}

static void simulate_stop_accepting(void *user_data) {
    (void)user_data;
    g_accepting_requests = 0;
    fprintf(stderr, "  [DEMO] Stopped accepting new requests.\n");
    fflush(stderr);
}

static void simulate_wait_requests(void *user_data) {
    (void)user_data;
    usleep(50000);
}

static void *request_worker(void *arg) {
    RequestInfo *info = (RequestInfo *)arg;
    
    shutdown_increment_active_requests();
    
    fprintf(stderr, "  [REQUEST #%d] Starting processing (will take %d ms)...\n",
            info->request_id, info->process_time_ms);
    fflush(stderr);
    
    usleep(info->process_time_ms * 1000);
    
    fprintf(stderr, "  [REQUEST #%d] Processing completed.\n", info->request_id);
    fflush(stderr);
    
    shutdown_decrement_active_requests();
    
    free(info);
    return NULL;
}

static void simulate_new_request(int request_id, int process_time_ms) {
    if (!g_accepting_requests) {
        return;
    }
    
    RequestInfo *info = (RequestInfo *)malloc(sizeof(RequestInfo));
    if (info == NULL) {
        return;
    }
    
    info->request_id = request_id;
    info->process_time_ms = process_time_ms;
    
    pthread_t thread;
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);
    
    if (pthread_create(&thread, &attr, request_worker, info) != 0) {
        free(info);
    }
    
    pthread_attr_destroy(&attr);
}

static void handle_reload_config(void) {
    fprintf(stderr, "[INFO] Received SIGHUP - Reloading configuration...\n");
    fflush(stderr);
    usleep(300000);
    fprintf(stderr, "[INFO] Configuration reloaded successfully.\n");
    fflush(stderr);
    signal_handler_clear_flag(SIGNAL_FLAG_RELOAD);
}

static void handle_rotate_log(void) {
    fprintf(stderr, "[INFO] Received SIGUSR1 - Rotating log files...\n");
    fflush(stderr);
    usleep(200000);
    fprintf(stderr, "[INFO] Log rotation completed.\n");
    fflush(stderr);
    signal_handler_clear_flag(SIGNAL_FLAG_ROTATE_LOG);
}

static void print_usage(const char *program_name) {
    fprintf(stderr, "Usage: %s [OPTIONS]\n", program_name);
    fprintf(stderr, "Options:\n");
    fprintf(stderr, "  -h, --help          Show this help message\n");
    fprintf(stderr, "  -f, --foreground    Run in foreground (not as daemon)\n");
    fprintf(stderr, "  -p, --pidfile PATH  Specify PID file path (default: %s)\n", PIDFILE_PATH);
    fprintf(stderr, "\nSignals handled:\n");
    fprintf(stderr, "  SIGTERM/SIGINT  - Graceful shutdown\n");
    fprintf(stderr, "  SIGHUP          - Reload configuration\n");
    fprintf(stderr, "  SIGUSR1         - Rotate log files\n");
}

int main(int argc, char *argv[]) {
    const char *pidfile_path = PIDFILE_PATH;
    bool run_as_daemon = true;
    int request_counter = 0;
    
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-h") == 0 || strcmp(argv[i], "--help") == 0) {
            print_usage(argv[0]);
            return EXIT_SUCCESS;
        } else if (strcmp(argv[i], "-f") == 0 || strcmp(argv[i], "--foreground") == 0) {
            run_as_daemon = false;
        } else if ((strcmp(argv[i], "-p") == 0 || strcmp(argv[i], "--pidfile") == 0) && i + 1 < argc) {
            pidfile_path = argv[++i];
        }
    }
    
    PidFileResult pid_result = pidfile_create(pidfile_path);
    if (pid_result == PIDFILE_ERROR_ALREADY_RUNNING) {
        fprintf(stderr, "[ERROR] Another instance is already running. PID file: %s\n", pidfile_path);
        return EXIT_FAILURE;
    } else if (pid_result != PIDFILE_OK) {
        fprintf(stderr, "[ERROR] Failed to create PID file %s (error: %d)\n", pidfile_path, pid_result);
        return EXIT_FAILURE;
    }
    
    if (run_as_daemon) {
        pid_t pid = fork();
        if (pid < 0) {
            perror("[ERROR] fork failed");
            pidfile_remove();
            return EXIT_FAILURE;
        }
        if (pid > 0) {
            fprintf(stdout, "Daemon started with PID: %d\n", pid);
            return EXIT_SUCCESS;
        }
        
        if (setsid() < 0) {
            perror("[ERROR] setsid failed");
            pidfile_remove();
            return EXIT_FAILURE;
        }
        
        pid = fork();
        if (pid < 0) {
            perror("[ERROR] second fork failed");
            pidfile_remove();
            return EXIT_FAILURE;
        }
        if (pid > 0) {
            return EXIT_SUCCESS;
        }
        
        umask(0);
        chdir("/");
        
        close(STDIN_FILENO);
        close(STDOUT_FILENO);
        close(STDERR_FILENO);
        
        open("/dev/null", O_RDONLY);
        open("/dev/null", O_WRONLY);
        open("/dev/null", O_WRONLY);
    }
    
    ShutdownConfig shutdown_config = {
        .stop_accepting = simulate_stop_accepting,
        .wait_requests_complete = simulate_wait_requests,
        .flush_logs = simulate_log_flush,
        .close_sockets = simulate_socket_close,
        .free_resources = simulate_free_resources,
        .user_data = NULL,
        .timeout_sec = SHUTDOWN_TIMEOUT_SEC
    };
    
    shutdown_init(&shutdown_config);
    signal_handler_init();
    
    fprintf(stderr, "[INFO] Daemon started. PID: %d, PID file: %s\n",
            (int)getpid(), pidfile_path);
    fprintf(stderr, "[INFO] Press Ctrl+C or send SIGTERM to stop gracefully.\n");
    fprintf(stderr, "[INFO] Send SIGHUP to reload config, SIGUSR1 to rotate logs.\n");
    fflush(stderr);
    
    int tick_count = 0;
    while (!signal_handler_is_stop_requested() && !shutdown_is_completed()) {
        SignalFlags flags = signal_handler_get_flags();
        
        if ((flags & SIGNAL_FLAG_RELOAD) != 0) {
            handle_reload_config();
        }
        
        if ((flags & SIGNAL_FLAG_ROTATE_LOG) != 0) {
            handle_rotate_log();
        }
        
        if ((flags & (SIGNAL_FLAG_TERMINATE | SIGNAL_FLAG_INTERRUPT)) != 0) {
            fprintf(stderr, "\n[INFO] Stop signal received. Initiating graceful shutdown...\n");
            fflush(stderr);
            
            ShutdownResult result = shutdown_execute();
            
            if (result == SHUTDOWN_ALREADY_IN_PROGRESS) {
                fprintf(stderr, "[INFO] Shutdown already in progress. Waiting...\n");
                fflush(stderr);
            }
            
            break;
        }
        
        tick_count++;
        if (tick_count % 5 == 0 && g_accepting_requests) {
            request_counter++;
            int process_time = 1000 + (rand() % 3000);
            fprintf(stderr, "[MAIN] Tick %d: Simulating new request #%d (processing time: %d ms)\n",
                    tick_count, request_counter, process_time);
            fflush(stderr);
            simulate_new_request(request_counter, process_time);
        }
        
        usleep(500000);
    }
    
    fprintf(stderr, "[INFO] Service stopping. Cleaning up...\n");
    fflush(stderr);
    
    signal_handler_cleanup();
    shutdown_cleanup();
    pidfile_remove();
    
    fprintf(stderr, "[INFO] Service stopped. Exit.\n");
    fflush(stderr);
    
    pthread_mutex_destroy(&g_request_mutex);
    
    return EXIT_SUCCESS;
}
