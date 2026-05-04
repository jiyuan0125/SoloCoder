#ifndef TABLE_MANAGER_H
#define TABLE_MANAGER_H

#include "common.h"
#include <pthread.h>
#include <semaphore.h>

#define TABLE_SUCCESS 0
#define TABLE_NOT_FOUND -1
#define TABLE_ALREADY_OCCUPIED -2
#define TABLE_ALREADY_FREE -3
#define TABLE_INVALID_PARAM -4

typedef struct {
    Table tables[MAX_TABLES];
    pthread_mutex_t mutex;
    sem_t free_tables;
    int table_count;
} TableManager;

extern TableManager g_table_manager;
extern Table g_tables[MAX_TABLES];

int table_manager_init(TableManager *manager);
int table_manager_destroy(TableManager *manager);
int table_occupy(TableManager *manager, int table_number, int customer_id);
int table_start_eating(TableManager *manager, int table_number, int eating_time_estimate);
int table_release(TableManager *manager, int table_number);
int table_force_release(TableManager *manager, int table_number);
int table_get_status(TableManager *manager, int table_number, TableStatus *status);
int table_find_free(TableManager *manager);
int table_get_customer_table(TableManager *manager, int customer_id);
int table_check_timeout(TableManager *manager, int timeout_threshold);
int table_get_count(TableManager *manager);
int table_get_free_count(TableManager *manager);
int table_get_occupied_count(TableManager *manager);
int table_list_all(TableManager *manager, Table *list, int max_list);
int table_get_info(TableManager *manager, int table_number, Table *info);

#endif
