#ifndef TASK_H
#define TASK_H

#include "cron_expr.h"
#include <pthread.h>
#include <time.h>

#define TASK_NAME_MAX_LEN 64
#define TASK_EXPR_MAX_LEN 128

typedef unsigned int task_id_t;
typedef void (*task_func_t)(void* user_data);

typedef enum {
    TASK_STATUS_ACTIVE = 0,
    TASK_STATUS_PAUSED,
    TASK_STATUS_DELETED
} task_status_t;

typedef struct {
    task_id_t id;
    char name[TASK_NAME_MAX_LEN];
    char cron_expr[TASK_EXPR_MAX_LEN];
    cron_expr_t parsed_expr;
    task_status_t status;
    task_func_t func;
    void* user_data;
    time_t last_run;
    time_t next_run;
    int run_count;
} task_t;

typedef struct task_node {
    task_t task;
    struct task_node* next;
} task_node_t;

typedef struct {
    task_node_t* head;
    pthread_mutex_t mutex;
    task_id_t next_id;
} task_manager_t;

int task_manager_init(task_manager_t* manager);

void task_manager_destroy(task_manager_t* manager);

task_id_t task_add(task_manager_t* manager, const char* name, const char* cron_expr, 
                     task_func_t func, void* user_data);

int task_remove(task_manager_t* manager, task_id_t id);

int task_pause(task_manager_t* manager, task_id_t id);

int task_resume(task_manager_t* manager, task_id_t id);

task_t* task_find(task_manager_t* manager, task_id_t id);

void task_list_lock(task_manager_t* manager);

void task_list_unlock(task_manager_t* manager);

task_node_t* task_list_get_head(task_manager_t* manager);

int task_count(task_manager_t* manager);

int task_update_next_run(task_t* task, time_t from_time);

void task_free(task_t* task);

#endif
