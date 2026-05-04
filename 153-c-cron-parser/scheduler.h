#ifndef SCHEDULER_H
#define SCHEDULER_H

#include "task.h"

typedef struct running_task {
    task_id_t task_id;
    pthread_t thread;
    time_t start_time;
    struct running_task* next;
} running_task_t;

typedef struct {
    task_manager_t task_manager;
    pthread_t scheduler_thread;
    pthread_mutex_t running_mutex;
    running_task_t* running_tasks;
    volatile int running;
    volatile int stopped;
} scheduler_t;

int scheduler_init(scheduler_t* sched);

void scheduler_destroy(scheduler_t* sched);

task_id_t scheduler_add_task(scheduler_t* sched, const char* name, const char* cron_expr,
                               task_func_t func, void* user_data);

int scheduler_remove_task(scheduler_t* sched, task_id_t id);

int scheduler_pause_task(scheduler_t* sched, task_id_t id);

int scheduler_resume_task(scheduler_t* sched, task_id_t id);

int scheduler_start(scheduler_t* sched);

void scheduler_stop(scheduler_t* sched);

int scheduler_list_tasks(scheduler_t* sched, char* buffer, size_t buf_size);

int scheduler_is_running(scheduler_t* sched);

task_manager_t* scheduler_get_task_manager(scheduler_t* sched);

void scheduler_join_all(scheduler_t* sched);

#endif
