#include "process_manager.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <stdio.h>
#include <errno.h>

static process_manager_t *g_signal_pm = NULL;

static void sigchld_handler(int sig) {
    (void)sig;
    if (g_signal_pm != NULL) {
        g_signal_pm->sigchld_received = 1;
    }
}

int process_manager_init(process_manager_t *pm) {
    memset(pm, 0, sizeof(process_manager_t));
    pm->worker_count = 0;
    pm->sigchld_received = 0;
    pm->event_fd = -1;
    g_signal_pm = pm;
    return 0;
}

void process_manager_destroy(process_manager_t *pm) {
    for (size_t i = 0; i < pm->worker_count; i++) {
        if (pm->workers[i].pid > 0) {
            kill(pm->workers[i].pid, SIGTERM);
            waitpid(pm->workers[i].pid, NULL, 0);
        }
        process_manager_close_worker_pipes(&pm->workers[i]);
    }
    pm->worker_count = 0;
    g_signal_pm = NULL;
}

void process_manager_close_worker_pipes(worker_process_t *worker) {
    pipe_pair_close(&worker->master_to_worker);
    pipe_pair_close(&worker->worker_to_master);
    message_buffer_destroy(&worker->recv_buffer);
}

int process_manager_create_worker(process_manager_t *pm, 
                                    void (*worker_main)(int read_fd, int write_fd)) {
    if (pm->worker_count >= MAX_WORKERS) {
        return -1;
    }
    
    worker_process_t *worker = &pm->workers[pm->worker_count];
    memset(worker, 0, sizeof(worker_process_t));
    
    worker->pid = -1;
    worker->state = WORKER_INIT;
    worker->master_to_worker.read_fd = -1;
    worker->master_to_worker.write_fd = -1;
    worker->worker_to_master.read_fd = -1;
    worker->worker_to_master.write_fd = -1;
    
    if (pipe_pair_create(&worker->master_to_worker) != 0) {
        goto cleanup;
    }
    
    if (pipe_pair_create(&worker->worker_to_master) != 0) {
        goto cleanup;
    }
    
    if (message_buffer_init(&worker->recv_buffer, 4096) != 0) {
        goto cleanup;
    }
    
    pid_t pid = fork();
    
    if (pid < 0) {
        goto cleanup;
    }
    
    if (pid == 0) {
        pipe_pair_close_write(&worker->master_to_worker);
        pipe_pair_close_read(&worker->worker_to_master);
        
        int worker_read_fd = worker->master_to_worker.read_fd;
        int worker_write_fd = worker->worker_to_master.write_fd;
        
        for (size_t i = 0; i < pm->worker_count; i++) {
            process_manager_close_worker_pipes(&pm->workers[i]);
        }
        
        worker_main(worker_read_fd, worker_write_fd);
        
        close(worker_read_fd);
        close(worker_write_fd);
        _exit(0);
    }
    
    pipe_pair_close_read(&worker->master_to_worker);
    pipe_pair_close_write(&worker->worker_to_master);
    
    pipe_set_nonblocking(worker->master_to_worker.write_fd);
    pipe_set_nonblocking(worker->worker_to_master.read_fd);
    
    worker->pid = pid;
    worker->state = WORKER_RUNNING;
    pm->worker_count++;
    
    return (int)(pm->worker_count - 1);
    
cleanup:
    process_manager_close_worker_pipes(worker);
    return -1;
}

worker_process_t *process_manager_get_worker(process_manager_t *pm, size_t index) {
    if (index >= pm->worker_count) {
        return NULL;
    }
    return &pm->workers[index];
}

worker_process_t *process_manager_get_worker_by_pid(process_manager_t *pm, pid_t pid) {
    for (size_t i = 0; i < pm->worker_count; i++) {
        if (pm->workers[i].pid == pid) {
            return &pm->workers[i];
        }
    }
    return NULL;
}

int process_manager_setup_sigchld(process_manager_t *pm) {
    struct sigaction sa;
    
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = sigchld_handler;
    sa.sa_flags = SA_RESTART;
    
    if (sigaction(SIGCHLD, &sa, NULL) == -1) {
        return -1;
    }
    
    g_signal_pm = pm;
    return 0;
}

void process_manager_check_children(process_manager_t *pm) {
    int status;
    pid_t pid;
    
    while ((pid = waitpid(-1, &status, WNOHANG)) > 0) {
        worker_process_t *worker = process_manager_get_worker_by_pid(pm, pid);
        if (worker != NULL) {
            worker->exit_status = status;
            if (WIFEXITED(status)) {
                worker->state = WORKER_EXITED;
            } else if (WIFSIGNALED(status)) {
                worker->state = WORKER_CRASHED;
            }
        }
    }
}

int process_manager_cleanup_exited(process_manager_t *pm) {
    size_t write_idx = 0;
    int cleaned = 0;
    
    for (size_t i = 0; i < pm->worker_count; i++) {
        worker_process_t *worker = &pm->workers[i];
        
        if (worker->state == WORKER_EXITED || worker->state == WORKER_CRASHED) {
            process_manager_close_worker_pipes(worker);
            cleaned++;
        } else {
            if (i != write_idx) {
                pm->workers[write_idx] = *worker;
            }
            write_idx++;
        }
    }
    
    pm->worker_count = write_idx;
    return cleaned;
}

ssize_t worker_send_message(worker_process_t *worker, const uint8_t *data, size_t len) {
    if (worker->state != WORKER_RUNNING) {
        return PIPE_ERROR;
    }
    
    if (worker->master_to_worker.write_fd < 0) {
        return PIPE_ERROR;
    }
    
    return pipe_write_message(worker->master_to_worker.write_fd, data, len);
}

ssize_t worker_recv_message(worker_process_t *worker, uint8_t **data, size_t *len) {
    if (worker->state != WORKER_RUNNING) {
        return PIPE_ERROR;
    }
    
    if (worker->worker_to_master.read_fd < 0) {
        return PIPE_ERROR;
    }
    
    if (message_buffer_has_complete_message(&worker->recv_buffer)) {
        if (message_buffer_extract_message(&worker->recv_buffer, data, len) != 0) {
            return PIPE_ERROR;
        }
        return (ssize_t)(*len);
    }
    
    ssize_t result = pipe_read_partial(worker->worker_to_master.read_fd, 
                                      &worker->recv_buffer);
    
    if (result > 0) {
        if (message_buffer_has_complete_message(&worker->recv_buffer)) {
            if (message_buffer_extract_message(&worker->recv_buffer, data, len) != 0) {
                return PIPE_ERROR;
            }
            return (ssize_t)(*len);
        }
        return PIPE_AGAIN;
    }
    
    return result;
}
