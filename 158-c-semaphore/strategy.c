#include "strategy.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

int strategy_init(SeatingStrategy *strategy, WaitQueue *queue, TableManager *table_manager) {
    if (strategy == NULL || queue == NULL || table_manager == NULL) {
        return STRATEGY_ERROR;
    }
    
    memset(strategy, 0, sizeof(SeatingStrategy));
    strategy->queue = queue;
    strategy->table_manager = table_manager;
    strategy->call_policy = CALL_POLICY_SKIP_LEFT;
    strategy->vip_enabled = 1;
    
    if (pthread_mutex_init(&strategy->mutex, NULL) != 0) {
        return STRATEGY_ERROR;
    }
    
    if (pthread_cond_init(&strategy->table_available, NULL) != 0) {
        pthread_mutex_destroy(&strategy->mutex);
        return STRATEGY_ERROR;
    }
    
    if (pthread_cond_init(&strategy->customer_arrived, NULL) != 0) {
        pthread_cond_destroy(&strategy->table_available);
        pthread_mutex_destroy(&strategy->mutex);
        return STRATEGY_ERROR;
    }
    
    return STRATEGY_SUCCESS;
}

int strategy_destroy(SeatingStrategy *strategy) {
    if (strategy == NULL) return STRATEGY_ERROR;
    
    pthread_cond_destroy(&strategy->customer_arrived);
    pthread_cond_destroy(&strategy->table_available);
    pthread_mutex_destroy(&strategy->mutex);
    
    return STRATEGY_SUCCESS;
}

int strategy_set_call_policy(SeatingStrategy *strategy, CallPolicy policy) {
    if (strategy == NULL) return STRATEGY_ERROR;
    
    pthread_mutex_lock(&strategy->mutex);
    strategy->call_policy = policy;
    pthread_mutex_unlock(&strategy->mutex);
    
    return STRATEGY_SUCCESS;
}

int strategy_enable_vip(SeatingStrategy *strategy, int enable) {
    if (strategy == NULL) return STRATEGY_ERROR;
    
    pthread_mutex_lock(&strategy->mutex);
    strategy->vip_enabled = enable;
    pthread_mutex_unlock(&strategy->mutex);
    
    return STRATEGY_SUCCESS;
}

int strategy_is_vip_enabled(SeatingStrategy *strategy) {
    if (strategy == NULL) return 0;
    
    pthread_mutex_lock(&strategy->mutex);
    int enabled = strategy->vip_enabled;
    pthread_mutex_unlock(&strategy->mutex);
    
    return enabled;
}

int strategy_try_seat_next_customer(SeatingStrategy *strategy, Customer *seated_customer, int *table_number) {
    if (strategy == NULL || seated_customer == NULL || table_number == NULL) {
        return STRATEGY_ERROR;
    }
    
    WaitQueue *queue = strategy->queue;
    TableManager *table_manager = strategy->table_manager;
    
    int free_table = table_find_free(table_manager);
    if (free_table < 0) {
        return STRATEGY_NO_TABLE;
    }
    
    Customer customer;
    int ret = queue_get_next_customer(queue, &customer);
    
    if (ret == QUEUE_EMPTY) {
        return STRATEGY_NO_CUSTOMER;
    }
    
    if (ret != QUEUE_SUCCESS) {
        return STRATEGY_ERROR;
    }
    
    if (customer.status != CUSTOMER_WAITING) {
        return STRATEGY_CUSTOMER_LEFT;
    }
    
    ret = table_occupy(table_manager, free_table, customer.id);
    if (ret != TABLE_SUCCESS) {
        return STRATEGY_ERROR;
    }
    
    memcpy(seated_customer, &customer, sizeof(Customer));
    *table_number = free_table;
    
    return STRATEGY_SUCCESS;
}

int strategy_try_seat_with_table(SeatingStrategy *strategy, int table_number, Customer *seated_customer) {
    if (strategy == NULL || seated_customer == NULL) {
        return STRATEGY_ERROR;
    }
    
    WaitQueue *queue = strategy->queue;
    TableManager *table_manager = strategy->table_manager;
    
    TableStatus status;
    if (table_get_status(table_manager, table_number, &status) != TABLE_SUCCESS) {
        return STRATEGY_ERROR;
    }
    
    if (status != TABLE_FREE) {
        return STRATEGY_NO_TABLE;
    }
    
    Customer customer;
    int ret = queue_get_next_customer(queue, &customer);
    
    if (ret == QUEUE_EMPTY) {
        return STRATEGY_NO_CUSTOMER;
    }
    
    if (ret != QUEUE_SUCCESS) {
        return STRATEGY_ERROR;
    }
    
    if (customer.status != CUSTOMER_WAITING) {
        return STRATEGY_CUSTOMER_LEFT;
    }
    
    ret = table_occupy(table_manager, table_number, customer.id);
    if (ret != TABLE_SUCCESS) {
        return STRATEGY_ERROR;
    }
    
    memcpy(seated_customer, &customer, sizeof(Customer));
    
    return STRATEGY_SUCCESS;
}

int strategy_call_and_check(SeatingStrategy *strategy, Customer *customer, int *table_number) {
    if (strategy == NULL || customer == NULL || table_number == NULL) {
        return STRATEGY_ERROR;
    }
    
    WaitQueue *queue = strategy->queue;
    TableManager *table_manager = strategy->table_manager;
    
    int free_table = table_find_free(table_manager);
    if (free_table < 0) {
        return STRATEGY_NO_TABLE;
    }
    
    Customer peeked;
    int ret = queue_peek_next_customer(queue, &peeked);
    
    if (ret == QUEUE_EMPTY) {
        return STRATEGY_NO_CUSTOMER;
    }
    
    int attempts = 0;
    int max_attempts = 10;
    
    while (attempts < max_attempts) {
        Customer actual;
        ret = queue_get_next_customer(queue, &actual);
        
        if (ret == QUEUE_EMPTY) {
            return STRATEGY_NO_CUSTOMER;
        }
        
        if (actual.status == CUSTOMER_WAITING) {
            ret = table_occupy(table_manager, free_table, actual.id);
            if (ret != TABLE_SUCCESS) {
                return STRATEGY_ERROR;
            }
            
            memcpy(customer, &actual, sizeof(Customer));
            *table_number = free_table;
            return STRATEGY_SUCCESS;
        }
        
        attempts++;
    }
    
    return STRATEGY_CUSTOMER_LEFT;
}

int strategy_auto_seat_loop(SeatingStrategy *strategy) {
    if (strategy == NULL) return STRATEGY_ERROR;
    
    int seated = 0;
    
    while (1) {
        Customer customer;
        int table_num;
        
        int ret = strategy_call_and_check(strategy, &customer, &table_num);
        
        if (ret == STRATEGY_SUCCESS) {
            seated++;
        } else if (ret == STRATEGY_NO_TABLE || ret == STRATEGY_NO_CUSTOMER) {
            break;
        }
    }
    
    return seated;
}

int strategy_get_system_status(SeatingStrategy *strategy, int *queue_count, int *free_tables, 
                                int *occupied_tables, double *avg_eating_time) {
    if (strategy == NULL) return STRATEGY_ERROR;
    
    if (queue_count != NULL) {
        *queue_count = queue_get_count(strategy->queue);
    }
    
    if (free_tables != NULL) {
        *free_tables = table_get_free_count(strategy->table_manager);
    }
    
    if (occupied_tables != NULL) {
        *occupied_tables = table_get_occupied_count(strategy->table_manager);
    }
    
    if (avg_eating_time != NULL) {
        *avg_eating_time = statistics_get_average_eating_time();
    }
    
    return STRATEGY_SUCCESS;
}

double strategy_estimate_wait_for_customer(SeatingStrategy *strategy, int customer_id) {
    if (strategy == NULL) return -1.0;
    
    int position = queue_get_customer_position(strategy->queue, customer_id);
    if (position < 0) {
        return -1.0;
    }
    
    return queue_estimate_wait_time(strategy->queue, position);
}

int strategy_force_timeout_check(SeatingStrategy *strategy, int timeout_seconds) {
    if (strategy == NULL || timeout_seconds <= 0) return STRATEGY_ERROR;
    
    int released = table_check_timeout(strategy->table_manager, timeout_seconds);
    
    if (released > 0) {
        strategy_auto_seat_loop(strategy);
    }
    
    return released;
}

int strategy_get_vip_count(SeatingStrategy *strategy) {
    if (strategy == NULL) return -1;
    
    return queue_get_vip_count(strategy->queue);
}
