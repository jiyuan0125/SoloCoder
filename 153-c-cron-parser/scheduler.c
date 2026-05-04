#define _GNU_SOURCE
#include "scheduler.h"
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>

typedef struct {
    task_func_t func;
    void* user_data;
    task_id_t task_id;
    scheduler_t* scheduler;
} task_thread_arg_t;

static void* task_thread_func(void* arg) {
    task_thread_arg_t* thread_arg = (task_thread_arg_t*)arg;
    if (!thread_arg) return NULL;

    task_func_t func = thread_arg->func;
    void* user_data = thread_arg->user_data;
    task_id_t task_id = thread_arg->task_id;
    scheduler_t* sched = thread_arg->scheduler;
    
    free(thread_arg);

    pthread_setcancelstate(PTHREAD_CANCEL_ENABLE, NULL);
    pthread_setcanceltype(PTHREAD_CANCEL_DEFERRED, NULL);

    if (func) {
        func(user_data);
    }

    pthread_mutex_lock(&sched->running_mutex);
    running_task_t** prev = &sched->running_tasks;
    running_task_t* current = sched->running_tasks;
    while (current != NULL) {
        if (current->task_id == task_id && pthread_equal(current->thread, pthread_self())) {
            *prev = current->next;
            free(current);
            break;
        }
        prev = &current->next;
        current = current->next;
    }
    pthread_mutex_unlock(&sched->running_mutex);

    return NULL;
}

static void run_task(scheduler_t* sched, task_t* task) {
    if (!sched || !task) return;

    task_thread_arg_t* arg = (task_thread_arg_t*)malloc(sizeof(task_thread_arg_t));
    if (!arg) {
        fprintf(stderr, "Failed to allocate memory for task thread\n");
        return;
    }

    arg->func = task->func;
    arg->user_data = task->user_data;
    arg->task_id = task->id;
    arg->scheduler = sched;

    pthread_t thread;
    int ret = pthread_create(&thread, NULL, task_thread_func, arg);
    if (ret != 0) {
        fprintf(stderr, "Failed to create task thread: %d\n", ret);
        free(arg);
        return;
    }

    pthread_detach(thread);

    pthread_mutex_lock(&sched->running_mutex);
    running_task_t* running = (running_task_t*)malloc(sizeof(running_task_t));
    if (running) {
        running->task_id = task->id;
        running->thread = thread;
        running->start_time = time(NULL);
        running->next = sched->running_tasks;
        sched->running_tasks = running;
    }
    pthread_mutex_unlock(&sched->running_mutex);
}

static time_t get_aligned_time(void) {
    time_t now = time(NULL);
    struct tm* tm_now = localtime(&now);
    if (!tm_now) return now;
    
    tm_now->tm_sec = 0;
    return mktime(tm_now);
}

static void* scheduler_thread(void* arg) {
    scheduler_t* sched = (scheduler_t*)arg;
    if (!sched) return NULL;

    sigset_t set;
    sigemptyset(&set);
    sigaddset(&set, SIGINT);
    sigaddset(&set, SIGTERM);
    pthread_sigmask(SIG_BLOCK, &set, NULL);

    time_t last_check = get_aligned_time();

    while (sched->running) {
        time_t now = time(NULL);
        
        if (now >= last_check + 60) {
            last_check = get_aligned_time();
            
            struct tm* tm_now = localtime(&last_check);
            if (tm_now) {
                task_list_lock(&sched->task_manager);
                task_node_t* node = task_list_get_head(&sched->task_manager);
                
                while (node != NULL) {
                    task_t* task = &node->task;
                    
                    if (task->status == TASK_STATUS_ACTIVE) {
                        if (cron_expr_matches(&task->parsed_expr, tm_now)) {
                            task->last_run = last_check;
                            task->run_count++;
                            run_task(sched, task);
                            task_update_next_run(task, last_check);
                        }
                    }
                    
                    node = node->next;
                }
                
                task_list_unlock(&sched->task_manager);
            }
        }

        if (sched->running) {
            sleep(1);
        }
    }

    sched->stopped = 1;
    return NULL;
}

int scheduler_init(scheduler_t* sched) {
    if (!sched) return -1;

    memset(sched, 0, sizeof(scheduler_t));
    
    if (task_manager_init(&sched->task_manager) != 0) {
        return -1;
    }

    if (pthread_mutex_init(&sched->running_mutex, NULL) != 0) {
        task_manager_destroy(&sched->task_manager);
        return -1;
    }

    sched->running = 0;
    sched->stopped = 1;
    sched->running_tasks = NULL;

    return 0;
}

void scheduler_destroy(scheduler_t* sched) {
    if (!sched) return;

    if (sched->running) {
        scheduler_stop(sched);
    }

    scheduler_join_all(sched);

    pthread_mutex_lock(&sched->running_mutex);
    running_task_t* current = sched->running_tasks;
    while (current != NULL) {
        running_task_t* next = current->next;
        pthread_cancel(current->thread);
        free(current);
        current = next;
    }
    sched->running_tasks = NULL;
    pthread_mutex_unlock(&sched->running_mutex);

    task_manager_destroy(&sched->task_manager);
    pthread_mutex_destroy(&sched->running_mutex);
}

task_id_t scheduler_add_task(scheduler_t* sched, const char* name, const char* cron_expr,
                               task_func_t func, void* user_data) {
    if (!sched) return 0;
    return task_add(&sched->task_manager, name, cron_expr, func, user_data);
}

int scheduler_remove_task(scheduler_t* sched, task_id_t id) {
    if (!sched) return -1;
    return task_remove(&sched->task_manager, id);
}

int scheduler_pause_task(scheduler_t* sched, task_id_t id) {
    if (!sched) return -1;
    return task_pause(&sched->task_manager, id);
}

int scheduler_resume_task(scheduler_t* sched, task_id_t id) {
    if (!sched) return -1;
    return task_resume(&sched->task_manager, id);
}

int scheduler_start(scheduler_t* sched) {
    if (!sched || sched->running) return -1;

    sched->running = 1;
    sched->stopped = 0;

    int ret = pthread_create(&sched->scheduler_thread, NULL, scheduler_thread, sched);
    if (ret != 0) {
        sched->running = 0;
        sched->stopped = 1;
        return -1;
    }

    return 0;
}

void scheduler_stop(scheduler_t* sched) {
    if (!sched || !sched->running) return;

    sched->running = 0;
    
    if (sched->scheduler_thread) {
        pthread_join(sched->scheduler_thread, NULL);
    }
}

int scheduler_list_tasks(scheduler_t* sched, char* buffer, size_t buf_size) {
    if (!sched || !buffer || buf_size == 0) return -1;

    buffer[0] = '\0';
    size_t offset = 0;

    task_list_lock(&sched->task_manager);
    task_node_t* node = task_list_get_head(&sched->task_manager);
    
    int count = 0;
    while (node != NULL) {
        task_t* task = &node->task;
        
        char next_time_str[64] = "N/A";
        if (task->next_run > 0) {
            struct tm* tm_next = localtime(&task->next_run);
            if (tm_next) {
                strftime(next_time_str, sizeof(next_time_str), "%Y-%m-%d %H:%M:%S", tm_next);
            }
        }

        const char* status_str = "UNKNOWN";
        switch (task->status) {
            case TASK_STATUS_ACTIVE: status_str = "ACTIVE"; break;
            case TASK_STATUS_PAUSED: status_str = "PAUSED"; break;
            case TASK_STATUS_DELETED: status_str = "DELETED"; break;
        }

        int len = snprintf(buffer + offset, buf_size - offset,
                           "ID: %u\n"
                           "  Name: %s\n"
                           "  Cron: %s\n"
                           "  Status: %s\n"
                           "  Next Run: %s\n"
                           "  Run Count: %d\n"
                           "------------------------\n",
                           task->id, task->name, task->cron_expr,
                           status_str, next_time_str, task->run_count);

        if (len < 0 || (size_t)len >= buf_size - offset) {
            task_list_unlock(&sched->task_manager);
            return -1;
        }
        offset += len;
        count++;
        node = node->next;
    }
    
    task_list_unlock(&sched->task_manager);

    if (count == 0) {
        snprintf(buffer, buf_size, "No tasks scheduled.\n");
    }

    return count;
}

int scheduler_is_running(scheduler_t* sched) {
    return (sched && sched->running);
}

task_manager_t* scheduler_get_task_manager(scheduler_t* sched) {
    if (!sched) return NULL;
    return &sched->task_manager;
}

void scheduler_join_all(scheduler_t* sched) {
    if (!sched) return;

    int wait_count = 0;
    const int max_wait = 50;
    
    while (wait_count < max_wait) {
        pthread_mutex_lock(&sched->running_mutex);
        int has_running = (sched->running_tasks != NULL);
        pthread_mutex_unlock(&sched->running_mutex);
        
        if (!has_running) break;
        
        usleep(100000);
        wait_count++;
    }
}
