#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <poll.h>
#include <errno.h>
#include <time.h>

#include "scheduler.h"
#include "cmd_exec.h"
#include "common.h"

#define MAX_TASKS 1024

typedef struct task_node_s {
    cmd_config_t config;
    cmd_result_t result;
    bool in_use;
    bool completed;
} task_node_t;

struct scheduler_s {
    int max_concurrent;
    int task_count;
    int running_count;
    int next_task_idx;
    task_node_t tasks[MAX_TASKS];
};

scheduler_t *scheduler_create(int max_concurrent) {
    scheduler_t *sched = calloc(1, sizeof(scheduler_t));
    if (!sched) return NULL;
    
    sched->max_concurrent = max_concurrent > 0 ? max_concurrent : DEFAULT_MAX_CONCURRENT;
    sched->task_count = 0;
    sched->running_count = 0;
    sched->next_task_idx = 0;
    
    for (int i = 0; i < MAX_TASKS; i++) {
        sched->tasks[i].in_use = false;
        sched->tasks[i].completed = false;
    }
    
    return sched;
}

void scheduler_destroy(scheduler_t *sched) {
    if (!sched) return;
    
    for (int i = 0; i < sched->task_count; i++) {
        if (sched->tasks[i].in_use) {
            cmd_result_free(&sched->tasks[i].result);
            cmd_config_free(&sched->tasks[i].config);
        }
    }
    
    free(sched);
}

int scheduler_add_task(scheduler_t *sched, cmd_config_t *config) {
    if (!sched || !config) return -1;
    if (sched->task_count >= MAX_TASKS) return -1;
    
    task_node_t *node = &sched->tasks[sched->task_count];
    
    node->config = *config;
    node->in_use = true;
    node->completed = false;
    
    if (cmd_result_init(&node->result, &node->config) != 0) {
        node->in_use = false;
        return -1;
    }
    
    sched->task_count++;
    return 0;
}

static int start_next_task(scheduler_t *sched) {
    if (sched->next_task_idx >= sched->task_count) return -1;
    
    task_node_t *node = &sched->tasks[sched->next_task_idx];
    
    if (cmd_start(&node->result) != 0) {
        return -1;
    }
    
    sched->next_task_idx++;
    sched->running_count++;
    return 0;
}

static int wait_for_any(scheduler_t *sched) {
    struct pollfd *pfds = NULL;
    int nfds = 0;
    int max_possible = sched->running_count * 2;
    
    if (max_possible > 0) {
        pfds = malloc(max_possible * sizeof(struct pollfd));
        if (!pfds) {
            usleep(100000);
            return 0;
        }
    }
    
    for (int i = 0; i < sched->task_count; i++) {
        task_node_t *node = &sched->tasks[i];
        if (!node->in_use || node->completed) continue;
        if (node->result.status != CMD_STATUS_RUNNING) continue;
        
        if (node->result.stdout_pipe[0] >= 0 && nfds < max_possible) {
            pfds[nfds].fd = node->result.stdout_pipe[0];
            pfds[nfds].events = POLLIN;
            nfds++;
        }
        if (node->result.stderr_pipe[0] >= 0 && nfds < max_possible) {
            pfds[nfds].fd = node->result.stderr_pipe[0];
            pfds[nfds].events = POLLIN;
            nfds++;
        }
    }
    
    if (nfds == 0) {
        free(pfds);
        usleep(100000);
        return 0;
    }
    
    int ret = poll(pfds, nfds, 100);
    if (ret <= 0) {
        free(pfds);
        return ret;
    }
    
    for (int i = 0; i < sched->task_count; i++) {
        task_node_t *node = &sched->tasks[i];
        if (!node->in_use || node->completed) continue;
        if (node->result.status != CMD_STATUS_RUNNING) continue;
        
        cmd_read_pipes(&node->result);
    }
    
    free(pfds);
    return ret;
}

static int check_completed_tasks(scheduler_t *sched) {
    int completed = 0;
    
    for (int i = 0; i < sched->task_count; i++) {
        task_node_t *node = &sched->tasks[i];
        if (!node->in_use || node->completed) continue;
        if (node->result.status != CMD_STATUS_RUNNING) continue;
        
        time_t now = time(NULL);
        double elapsed = difftime(now, node->result.start_time);
        int timeout = node->config.timeout_sec > 0 ? node->config.timeout_sec : DEFAULT_TIMEOUT_SEC;
        
        if (elapsed >= timeout && !node->result.timed_out) {
            node->result.timed_out = true;
            cmd_kill_process_group(node->result.pid);
        }
        
        int status;
        pid_t w = waitpid(node->result.pid, &status, WNOHANG);
        
        if (w < 0) {
            if (errno != ECHILD) continue;
            w = node->result.pid;
        }
        
        if (w > 0) {
            while (cmd_read_pipes(&node->result) > 0);
            
            node->result.end_time = time(NULL);
            node->result.elapsed_sec = difftime(node->result.end_time, node->result.start_time);
            
            if (node->result.timed_out) {
                node->result.status = CMD_STATUS_TIMEOUT;
                node->result.exit_code = CMD_EXIT_TIMEOUT;
            } else if (WIFEXITED(status)) {
                node->result.exit_code = WEXITSTATUS(status);
                node->result.status = (node->result.exit_code == 0) ? CMD_STATUS_SUCCESS : CMD_STATUS_FAILED;
            } else if (WIFSIGNALED(status)) {
                node->result.exit_code = 128 + WTERMSIG(status);
                node->result.status = CMD_STATUS_KILLED;
            }
            
            node->completed = true;
            sched->running_count--;
            completed++;
        }
    }
    
    return completed;
}

int scheduler_run(scheduler_t *sched) {
    if (!sched) return -1;
    
    while (sched->running_count > 0 || sched->next_task_idx < sched->task_count) {
        while (sched->running_count < sched->max_concurrent && 
               sched->next_task_idx < sched->task_count) {
            if (start_next_task(sched) != 0) {
                sched->tasks[sched->next_task_idx].result.status = CMD_STATUS_FAILED;
                sched->tasks[sched->next_task_idx].completed = true;
                sched->next_task_idx++;
            }
        }
        
        wait_for_any(sched);
        
        check_completed_tasks(sched);
    }
    
    return 0;
}

cmd_result_t *scheduler_get_results(scheduler_t *sched, size_t *count) {
    if (!sched || !count) {
        if (count) *count = 0;
        return NULL;
    }
    
    *count = sched->task_count;
    
    cmd_result_t *results = malloc(sched->task_count * sizeof(cmd_result_t));
    if (!results) {
        *count = 0;
        return NULL;
    }
    
    for (int i = 0; i < sched->task_count; i++) {
        results[i] = sched->tasks[i].result;
    }
    
    return results;
}

int scheduler_get_task_count(scheduler_t *sched) {
    if (!sched) return 0;
    return sched->task_count;
}
