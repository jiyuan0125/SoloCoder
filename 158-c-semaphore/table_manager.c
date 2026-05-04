#include "table_manager.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

TableManager g_table_manager;
Table g_tables[MAX_TABLES];

int table_manager_init(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    memset(manager, 0, sizeof(TableManager));
    manager->table_count = MAX_TABLES;
    
    for (int i = 0; i < MAX_TABLES; i++) {
        manager->tables[i].number = i + 1;
        manager->tables[i].status = TABLE_FREE;
        manager->tables[i].customer_id = -1;
        manager->tables[i].occupied_time = 0;
        manager->tables[i].eating_start_time = 0;
        manager->tables[i].eating_time_estimate = DEFAULT_EATING_TIME;
        memcpy(&g_tables[i], &manager->tables[i], sizeof(Table));
    }
    
    if (pthread_mutex_init(&manager->mutex, NULL) != 0) {
        return -1;
    }
    
    if (sem_init(&manager->free_tables, 0, MAX_TABLES) != 0) {
        pthread_mutex_destroy(&manager->mutex);
        return -1;
    }
    
    return TABLE_SUCCESS;
}

int table_manager_destroy(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    pthread_mutex_destroy(&manager->mutex);
    sem_destroy(&manager->free_tables);
    
    return TABLE_SUCCESS;
}

int table_occupy(TableManager *manager, int table_number, int customer_id) {
    if (manager == NULL || table_number < 1 || table_number > MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    int idx = table_number - 1;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (manager->tables[idx].status != TABLE_FREE) {
        pthread_mutex_unlock(&manager->mutex);
        return TABLE_ALREADY_OCCUPIED;
    }
    
    if (sem_trywait(&manager->free_tables) != 0) {
        pthread_mutex_unlock(&manager->mutex);
        return -1;
    }
    
    manager->tables[idx].status = TABLE_OCCUPIED;
    manager->tables[idx].customer_id = customer_id;
    manager->tables[idx].occupied_time = time(NULL);
    manager->tables[idx].eating_start_time = 0;
    manager->tables[idx].eating_time_estimate = DEFAULT_EATING_TIME;
    
    memcpy(&g_tables[idx], &manager->tables[idx], sizeof(Table));
    
    pthread_mutex_unlock(&manager->mutex);
    
    return TABLE_SUCCESS;
}

int table_start_eating(TableManager *manager, int table_number, int eating_time_estimate) {
    if (manager == NULL || table_number < 1 || table_number > MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    int idx = table_number - 1;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (manager->tables[idx].status != TABLE_OCCUPIED) {
        pthread_mutex_unlock(&manager->mutex);
        return TABLE_NOT_FOUND;
    }
    
    manager->tables[idx].status = TABLE_EATING;
    manager->tables[idx].eating_start_time = time(NULL);
    manager->tables[idx].eating_time_estimate = eating_time_estimate;
    
    memcpy(&g_tables[idx], &manager->tables[idx], sizeof(Table));
    
    pthread_mutex_unlock(&manager->mutex);
    
    return TABLE_SUCCESS;
}

int table_release(TableManager *manager, int table_number) {
    if (manager == NULL || table_number < 1 || table_number > MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    int idx = table_number - 1;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (manager->tables[idx].status == TABLE_FREE) {
        pthread_mutex_unlock(&manager->mutex);
        return TABLE_ALREADY_FREE;
    }
    
    if (manager->tables[idx].status == TABLE_EATING && 
        manager->tables[idx].eating_start_time > 0) {
        time_t now = time(NULL);
        double actual_eating_time = difftime(now, manager->tables[idx].eating_start_time);
        statistics_update(actual_eating_time);
    }
    
    manager->tables[idx].status = TABLE_FREE;
    manager->tables[idx].customer_id = -1;
    manager->tables[idx].occupied_time = 0;
    manager->tables[idx].eating_start_time = 0;
    
    memcpy(&g_tables[idx], &manager->tables[idx], sizeof(Table));
    
    pthread_mutex_unlock(&manager->mutex);
    sem_post(&manager->free_tables);
    
    return TABLE_SUCCESS;
}

int table_force_release(TableManager *manager, int table_number) {
    return table_release(manager, table_number);
}

int table_get_status(TableManager *manager, int table_number, TableStatus *status) {
    if (manager == NULL || status == NULL || table_number < 1 || table_number > MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    int idx = table_number - 1;
    
    pthread_mutex_lock(&manager->mutex);
    *status = manager->tables[idx].status;
    pthread_mutex_unlock(&manager->mutex);
    
    return TABLE_SUCCESS;
}

int table_find_free(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        if (manager->tables[i].status == TABLE_FREE) {
            int table_num = manager->tables[i].number;
            pthread_mutex_unlock(&manager->mutex);
            return table_num;
        }
    }
    
    pthread_mutex_unlock(&manager->mutex);
    return -1;
}

int table_get_customer_table(TableManager *manager, int customer_id) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        if (manager->tables[i].customer_id == customer_id &&
            (manager->tables[i].status == TABLE_OCCUPIED || 
             manager->tables[i].status == TABLE_EATING)) {
            int table_num = manager->tables[i].number;
            pthread_mutex_unlock(&manager->mutex);
            return table_num;
        }
    }
    
    pthread_mutex_unlock(&manager->mutex);
    return -1;
}

int table_check_timeout(TableManager *manager, int timeout_threshold) {
    if (manager == NULL || timeout_threshold <= 0) return TABLE_INVALID_PARAM;
    
    time_t now = time(NULL);
    int released_count = 0;
    
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        if (manager->tables[i].status == TABLE_OCCUPIED && 
            manager->tables[i].occupied_time > 0) {
            
            double elapsed = difftime(now, manager->tables[i].occupied_time);
            if (elapsed >= timeout_threshold) {
                manager->tables[i].status = TABLE_FREE;
                manager->tables[i].customer_id = -1;
                manager->tables[i].occupied_time = 0;
                manager->tables[i].eating_start_time = 0;
                memcpy(&g_tables[i], &manager->tables[i], sizeof(Table));
                released_count++;
                sem_post(&manager->free_tables);
            }
        }
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return released_count;
}

int table_get_count(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    return MAX_TABLES;
}

int table_get_free_count(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    int count = 0;
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        if (manager->tables[i].status == TABLE_FREE) {
            count++;
        }
    }
    
    pthread_mutex_unlock(&manager->mutex);
    return count;
}

int table_get_occupied_count(TableManager *manager) {
    if (manager == NULL) return TABLE_INVALID_PARAM;
    
    int count = 0;
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        if (manager->tables[i].status == TABLE_OCCUPIED || 
            manager->tables[i].status == TABLE_EATING) {
            count++;
        }
    }
    
    pthread_mutex_unlock(&manager->mutex);
    return count;
}

int table_list_all(TableManager *manager, Table *list, int max_list) {
    if (manager == NULL || list == NULL || max_list < MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    pthread_mutex_lock(&manager->mutex);
    
    for (int i = 0; i < MAX_TABLES; i++) {
        memcpy(&list[i], &manager->tables[i], sizeof(Table));
    }
    
    pthread_mutex_unlock(&manager->mutex);
    return MAX_TABLES;
}

int table_get_info(TableManager *manager, int table_number, Table *info) {
    if (manager == NULL || info == NULL || table_number < 1 || table_number > MAX_TABLES) {
        return TABLE_INVALID_PARAM;
    }
    
    int idx = table_number - 1;
    
    pthread_mutex_lock(&manager->mutex);
    memcpy(info, &manager->tables[idx], sizeof(Table));
    pthread_mutex_unlock(&manager->mutex);
    
    return TABLE_SUCCESS;
}
