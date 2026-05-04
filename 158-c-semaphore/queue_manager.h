#ifndef QUEUE_MANAGER_H
#define QUEUE_MANAGER_H

#include "common.h"
#include <semaphore.h>

#define QUEUE_SUCCESS 0
#define QUEUE_FULL -1
#define QUEUE_EMPTY -2
#define CUSTOMER_NOT_FOUND -3
#define INVALID_PARAM -4

typedef struct {
    Customer customers[MAX_QUEUE];
    int front;
    int rear;
    int count;
    pthread_mutex_t mutex;
    sem_t available_slots;
    sem_t waiting_customers;
    int next_customer_id;
} WaitQueue;

int queue_init(WaitQueue *queue);
int queue_destroy(WaitQueue *queue);
int queue_add_customer(WaitQueue *queue, const char *name, CustomerType type, int *assigned_id);
int queue_add_vip_customer(WaitQueue *queue, const char *name, int *assigned_id);
int queue_cancel_customer(WaitQueue *queue, int customer_id);
int queue_cancel_customer_by_name(WaitQueue *queue, const char *name);
int queue_get_next_customer(WaitQueue *queue, Customer *customer);
int queue_peek_next_customer(WaitQueue *queue, Customer *customer);
int queue_get_count(WaitQueue *queue);
int queue_get_vip_count(WaitQueue *queue);
int queue_find_customer(WaitQueue *queue, int customer_id, Customer *customer);
int queue_list_all_customers(WaitQueue *queue, Customer *list, int max_list);
double queue_estimate_wait_time(WaitQueue *queue, int position);
int queue_get_customer_position(WaitQueue *queue, int customer_id);
int queue_get_max_capacity(void);

#endif
