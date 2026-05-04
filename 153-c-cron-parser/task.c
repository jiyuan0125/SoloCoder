#define _GNU_SOURCE
#include "task.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

int task_manager_init(task_manager_t* manager) {
    if (!manager) return -1;
    manager->head = NULL;
    manager->next_id = 1;
    if (pthread_mutex_init(&manager->mutex, NULL) != 0) {
        return -1;
    }
    return 0;
}

void task_manager_destroy(task_manager_t* manager) {
    if (!manager) return;
    
    pthread_mutex_lock(&manager->mutex);
    
    task_node_t* current = manager->head;
    while (current != NULL) {
        task_node_t* next = current->next;
        task_free(&current->task);
        free(current);
        current = next;
    }
    manager->head = NULL;
    
    pthread_mutex_unlock(&manager->mutex);
    pthread_mutex_destroy(&manager->mutex);
}

task_id_t task_add(task_manager_t* manager, const char* name, const char* cron_expr, 
                     task_func_t func, void* user_data) {
    if (!manager || !name || !cron_expr || !func) return 0;

    task_node_t* node = (task_node_t*)malloc(sizeof(task_node_t));
    if (!node) return 0;
    memset(node, 0, sizeof(task_node_t));

    pthread_mutex_lock(&manager->mutex);

    task_t* task = &node->task;
    task->id = manager->next_id++;
    
    strncpy(task->name, name, TASK_NAME_MAX_LEN - 1);
    task->name[TASK_NAME_MAX_LEN - 1] = '\0';
    
    strncpy(task->cron_expr, cron_expr, TASK_EXPR_MAX_LEN - 1);
    task->cron_expr[TASK_EXPR_MAX_LEN - 1] = '\0';

    cron_parse_error_t err = cron_expr_parse(cron_expr, &task->parsed_expr);
    if (err != CRON_PARSE_OK) {
        pthread_mutex_unlock(&manager->mutex);
        free(node);
        fprintf(stderr, "Failed to parse cron expression: %s\n", cron_parse_strerror(err));
        return 0;
    }

    task->func = func;
    task->user_data = user_data;
    task->status = TASK_STATUS_ACTIVE;
    task->last_run = 0;
    task->run_count = 0;

    time_t now = time(NULL);
    if (task_update_next_run(task, now) != 0) {
        pthread_mutex_unlock(&manager->mutex);
        free(node);
        fprintf(stderr, "Failed to calculate next run time\n");
        return 0;
    }

    node->next = manager->head;
    manager->head = node;

    pthread_mutex_unlock(&manager->mutex);
    
    return task->id;
}

int task_remove(task_manager_t* manager, task_id_t id) {
    if (!manager || id == 0) return -1;

    pthread_mutex_lock(&manager->mutex);

    task_node_t* current = manager->head;
    task_node_t* prev = NULL;

    while (current != NULL) {
        if (current->task.id == id) {
            if (prev == NULL) {
                manager->head = current->next;
            } else {
                prev->next = current->next;
            }
            task_free(&current->task);
            free(current);
            pthread_mutex_unlock(&manager->mutex);
            return 0;
        }
        prev = current;
        current = current->next;
    }

    pthread_mutex_unlock(&manager->mutex);
    return -1;
}

int task_pause(task_manager_t* manager, task_id_t id) {
    if (!manager || id == 0) return -1;

    pthread_mutex_lock(&manager->mutex);

    task_node_t* current = manager->head;
    while (current != NULL) {
        if (current->task.id == id) {
            current->task.status = TASK_STATUS_PAUSED;
            pthread_mutex_unlock(&manager->mutex);
            return 0;
        }
        current = current->next;
    }

    pthread_mutex_unlock(&manager->mutex);
    return -1;
}

int task_resume(task_manager_t* manager, task_id_t id) {
    if (!manager || id == 0) return -1;

    pthread_mutex_lock(&manager->mutex);

    task_node_t* current = manager->head;
    while (current != NULL) {
        if (current->task.id == id) {
            if (current->task.status == TASK_STATUS_PAUSED) {
                current->task.status = TASK_STATUS_ACTIVE;
                time_t now = time(NULL);
                task_update_next_run(&current->task, now);
            }
            pthread_mutex_unlock(&manager->mutex);
            return 0;
        }
        current = current->next;
    }

    pthread_mutex_unlock(&manager->mutex);
    return -1;
}

task_t* task_find(task_manager_t* manager, task_id_t id) {
    if (!manager || id == 0) return NULL;

    pthread_mutex_lock(&manager->mutex);

    task_node_t* current = manager->head;
    while (current != NULL) {
        if (current->task.id == id) {
            pthread_mutex_unlock(&manager->mutex);
            return &current->task;
        }
        current = current->next;
    }

    pthread_mutex_unlock(&manager->mutex);
    return NULL;
}

void task_list_lock(task_manager_t* manager) {
    if (manager) pthread_mutex_lock(&manager->mutex);
}

void task_list_unlock(task_manager_t* manager) {
    if (manager) pthread_mutex_unlock(&manager->mutex);
}

task_node_t* task_list_get_head(task_manager_t* manager) {
    if (!manager) return NULL;
    return manager->head;
}

int task_count(task_manager_t* manager) {
    if (!manager) return 0;

    pthread_mutex_lock(&manager->mutex);

    int count = 0;
    task_node_t* current = manager->head;
    while (current != NULL) {
        count++;
        current = current->next;
    }

    pthread_mutex_unlock(&manager->mutex);
    return count;
}

int task_update_next_run(task_t* task, time_t from_time) {
    if (!task) return -1;
    
    time_t check_time = from_time + 60;
    return cron_calculate_next_run(&task->parsed_expr, &check_time, &task->next_run);
}

void task_free(task_t* task) {
    if (!task) return;
    cron_expr_clear(&task->parsed_expr);
}
