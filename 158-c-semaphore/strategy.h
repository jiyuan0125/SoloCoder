#ifndef STRATEGY_H
#define STRATEGY_H

#include "common.h"
#include "queue_manager.h"
#include "table_manager.h"
#include <pthread.h>

#define STRATEGY_SUCCESS 0
#define STRATEGY_ERROR -1
#define STRATEGY_NO_CUSTOMER -2
#define STRATEGY_NO_TABLE -3
#define STRATEGY_CUSTOMER_LEFT -4

typedef enum {
    CALL_POLICY_NORMAL = 0,
    CALL_POLICY_SKIP_LEFT,
    CALL_POLICY_VIP_FIRST
} CallPolicy;

typedef struct {
    WaitQueue *queue;
    TableManager *table_manager;
    CallPolicy call_policy;
    int vip_enabled;
    pthread_mutex_t mutex;
    pthread_cond_t table_available;
    pthread_cond_t customer_arrived;
} SeatingStrategy;

int strategy_init(SeatingStrategy *strategy, WaitQueue *queue, TableManager *table_manager);
int strategy_destroy(SeatingStrategy *strategy);
int strategy_set_call_policy(SeatingStrategy *strategy, CallPolicy policy);
int strategy_enable_vip(SeatingStrategy *strategy, int enable);
int strategy_is_vip_enabled(SeatingStrategy *strategy);
int strategy_try_seat_next_customer(SeatingStrategy *strategy, Customer *seated_customer, int *table_number);
int strategy_try_seat_with_table(SeatingStrategy *strategy, int table_number, Customer *seated_customer);
int strategy_call_and_check(SeatingStrategy *strategy, Customer *customer, int *table_number);
int strategy_auto_seat_loop(SeatingStrategy *strategy);
int strategy_get_system_status(SeatingStrategy *strategy, int *queue_count, int *free_tables, 
                                int *occupied_tables, double *avg_eating_time);
double strategy_estimate_wait_for_customer(SeatingStrategy *strategy, int customer_id);
int strategy_force_timeout_check(SeatingStrategy *strategy, int timeout_seconds);
int strategy_get_vip_count(SeatingStrategy *strategy);

#endif
