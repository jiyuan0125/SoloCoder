#ifndef PROCESS_MANAGER_H
#define PROCESS_MANAGER_H

#include <sys/types.h>
#include <signal.h>
#include "pipe_manager.h"

#define MAX_WORKERS 64

typedef enum {
    WORKER_INIT = 0,
    WORKER_RUNNING,
    WORKER_EXITED,
    WORKER_CRASHED
} worker_state_t;

typedef struct {
    pid_t pid;
    worker_state_t state;
    int exit_status;
    
    pipe_pair_t master_to_worker;
    pipe_pair_t worker_to_master;
    
    message_buffer_t recv_buffer;
} worker_process_t;

typedef struct {
    worker_process_t workers[MAX_WORKERS];
    size_t worker_count;
    
    volatile sig_atomic_t sigchld_received;
    int event_fd;
} process_manager_t;

int process_manager_init(process_manager_t *pm);
void process_manager_destroy(process_manager_t *pm);

int process_manager_create_worker(process_manager_t *pm, 
                                    void (*worker_main)(int read_fd, int write_fd));

worker_process_t *process_manager_get_worker(process_manager_t *pm, size_t index);
worker_process_t *process_manager_get_worker_by_pid(process_manager_t *pm, pid_t pid);

int process_manager_cleanup_exited(process_manager_t *pm);
void process_manager_close_worker_pipes(worker_process_t *worker);

int process_manager_setup_sigchld(process_manager_t *pm);
void process_manager_check_children(process_manager_t *pm);

ssize_t worker_send_message(worker_process_t *worker, const uint8_t *data, size_t len);
ssize_t worker_recv_message(worker_process_t *worker, uint8_t **data, size_t *len);

#endif
