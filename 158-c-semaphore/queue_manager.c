#include "queue_manager.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

int queue_init(WaitQueue *queue) {
    if (queue == NULL) return INVALID_PARAM;
    
    memset(queue, 0, sizeof(WaitQueue));
    queue->front = 0;
    queue->rear = -1;
    queue->count = 0;
    queue->next_customer_id = 1;
    
    if (pthread_mutex_init(&queue->mutex, NULL) != 0) {
        return -1;
    }
    
    if (sem_init(&queue->available_slots, 0, MAX_QUEUE) != 0) {
        pthread_mutex_destroy(&queue->mutex);
        return -1;
    }
    
    if (sem_init(&queue->waiting_customers, 0, 0) != 0) {
        sem_destroy(&queue->available_slots);
        pthread_mutex_destroy(&queue->mutex);
        return -1;
    }
    
    return QUEUE_SUCCESS;
}

int queue_destroy(WaitQueue *queue) {
    if (queue == NULL) return INVALID_PARAM;
    
    pthread_mutex_destroy(&queue->mutex);
    sem_destroy(&queue->available_slots);
    sem_destroy(&queue->waiting_customers);
    
    return QUEUE_SUCCESS;
}

static int queue_full(WaitQueue *queue) {
    return queue->count >= MAX_QUEUE;
}

static int queue_empty(WaitQueue *queue) {
    return queue->count <= 0;
}

int queue_add_customer(WaitQueue *queue, const char *name, CustomerType type, int *assigned_id) {
    if (queue == NULL || name == NULL || assigned_id == NULL) {
        return INVALID_PARAM;
    }
    
    if (sem_trywait(&queue->available_slots) != 0) {
        return QUEUE_FULL;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    if (queue_full(queue)) {
        pthread_mutex_unlock(&queue->mutex);
        sem_post(&queue->available_slots);
        return QUEUE_FULL;
    }
    
    int id = queue->next_customer_id++;
    int insert_pos;
    
    if (type == CUSTOMER_VIP) {
        insert_pos = queue->front;
        for (int i = 0; i < queue->count; i++) {
            int idx = (queue->front + i) % MAX_QUEUE;
            if (queue->customers[idx].type == CUSTOMER_NORMAL) {
                insert_pos = idx;
                break;
            }
        }
        
        if (insert_pos == queue->front && queue->count > 0) {
            for (int i = queue->count; i > 0; i--) {
                int dst = (queue->front + i) % MAX_QUEUE;
                int src = (queue->front + i - 1) % MAX_QUEUE;
                memcpy(&queue->customers[dst], &queue->customers[src], sizeof(Customer));
            }
            insert_pos = queue->front;
        } else if (insert_pos != queue->front) {
            for (int i = queue->count; i > (insert_pos - queue->front + MAX_QUEUE) % MAX_QUEUE; i--) {
                int dst = (queue->front + i) % MAX_QUEUE;
                int src = (queue->front + i - 1) % MAX_QUEUE;
                memcpy(&queue->customers[dst], &queue->customers[src], sizeof(Customer));
            }
        } else {
            insert_pos = (queue->rear + 1) % MAX_QUEUE;
            queue->rear = (queue->rear + 1) % MAX_QUEUE;
        }
    } else {
        insert_pos = (queue->rear + 1) % MAX_QUEUE;
        queue->rear = (queue->rear + 1) % MAX_QUEUE;
    }
    
    Customer *c = &queue->customers[insert_pos];
    c->id = id;
    strncpy(c->name, name, sizeof(c->name) - 1);
    c->name[sizeof(c->name) - 1] = '\0';
    c->type = type;
    c->status = CUSTOMER_WAITING;
    c->queue_time = time(NULL);
    c->seat_time = 0;
    c->table_number = -1;
    
    queue->count++;
    
    *assigned_id = id;
    
    pthread_mutex_unlock(&queue->mutex);
    sem_post(&queue->waiting_customers);
    
    return QUEUE_SUCCESS;
}

int queue_add_vip_customer(WaitQueue *queue, const char *name, int *assigned_id) {
    return queue_add_customer(queue, name, CUSTOMER_VIP, assigned_id);
}

int queue_cancel_customer(WaitQueue *queue, int customer_id) {
    if (queue == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    int found = -1;
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].id == customer_id && 
            queue->customers[idx].status == CUSTOMER_WAITING) {
            found = i;
            queue->customers[idx].status = CUSTOMER_CANCELLED;
            break;
        }
    }
    
    if (found == -1) {
        pthread_mutex_unlock(&queue->mutex);
        return CUSTOMER_NOT_FOUND;
    }
    
    for (int i = found; i < queue->count - 1; i++) {
        int dst = (queue->front + i) % MAX_QUEUE;
        int src = (queue->front + i + 1) % MAX_QUEUE;
        memcpy(&queue->customers[dst], &queue->customers[src], sizeof(Customer));
    }
    
    queue->rear = (queue->rear - 1 + MAX_QUEUE) % MAX_QUEUE;
    queue->count--;
    
    pthread_mutex_unlock(&queue->mutex);
    sem_post(&queue->available_slots);
    
    return QUEUE_SUCCESS;
}

int queue_cancel_customer_by_name(WaitQueue *queue, const char *name) {
    if (queue == NULL || name == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    int found = -1;
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (strcmp(queue->customers[idx].name, name) == 0 && 
            queue->customers[idx].status == CUSTOMER_WAITING) {
            found = i;
            queue->customers[idx].status = CUSTOMER_CANCELLED;
            break;
        }
    }
    
    if (found == -1) {
        pthread_mutex_unlock(&queue->mutex);
        return CUSTOMER_NOT_FOUND;
    }
    
    for (int i = found; i < queue->count - 1; i++) {
        int dst = (queue->front + i) % MAX_QUEUE;
        int src = (queue->front + i + 1) % MAX_QUEUE;
        memcpy(&queue->customers[dst], &queue->customers[src], sizeof(Customer));
    }
    
    queue->rear = (queue->rear - 1 + MAX_QUEUE) % MAX_QUEUE;
    queue->count--;
    
    pthread_mutex_unlock(&queue->mutex);
    sem_post(&queue->available_slots);
    
    return QUEUE_SUCCESS;
}

int queue_get_next_customer(WaitQueue *queue, Customer *customer) {
    if (queue == NULL || customer == NULL) return INVALID_PARAM;
    
    if (sem_trywait(&queue->waiting_customers) != 0) {
        return QUEUE_EMPTY;
    }
    
    pthread_mutex_lock(&queue->mutex);
    
    while (!queue_empty(queue)) {
        int idx = queue->front;
        Customer *c = &queue->customers[idx];
        
        if (c->status == CUSTOMER_WAITING) {
            memcpy(customer, c, sizeof(Customer));
            queue->front = (queue->front + 1) % MAX_QUEUE;
            queue->count--;
            pthread_mutex_unlock(&queue->mutex);
            sem_post(&queue->available_slots);
            return QUEUE_SUCCESS;
        } else {
            queue->front = (queue->front + 1) % MAX_QUEUE;
            queue->count--;
            sem_post(&queue->available_slots);
        }
    }
    
    pthread_mutex_unlock(&queue->mutex);
    return QUEUE_EMPTY;
}

int queue_peek_next_customer(WaitQueue *queue, Customer *customer) {
    if (queue == NULL || customer == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    if (queue_empty(queue)) {
        pthread_mutex_unlock(&queue->mutex);
        return QUEUE_EMPTY;
    }
    
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].status == CUSTOMER_WAITING) {
            memcpy(customer, &queue->customers[idx], sizeof(Customer));
            pthread_mutex_unlock(&queue->mutex);
            return QUEUE_SUCCESS;
        }
    }
    
    pthread_mutex_unlock(&queue->mutex);
    return QUEUE_EMPTY;
}

int queue_get_count(WaitQueue *queue) {
    if (queue == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    int count = queue->count;
    pthread_mutex_unlock(&queue->mutex);
    
    return count;
}

int queue_get_vip_count(WaitQueue *queue) {
    if (queue == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    int vip_count = 0;
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].type == CUSTOMER_VIP && 
            queue->customers[idx].status == CUSTOMER_WAITING) {
            vip_count++;
        }
    }
    pthread_mutex_unlock(&queue->mutex);
    
    return vip_count;
}

int queue_find_customer(WaitQueue *queue, int customer_id, Customer *customer) {
    if (queue == NULL || customer == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].id == customer_id) {
            memcpy(customer, &queue->customers[idx], sizeof(Customer));
            pthread_mutex_unlock(&queue->mutex);
            return QUEUE_SUCCESS;
        }
    }
    
    pthread_mutex_unlock(&queue->mutex);
    return CUSTOMER_NOT_FOUND;
}

int queue_list_all_customers(WaitQueue *queue, Customer *list, int max_list) {
    if (queue == NULL || list == NULL || max_list <= 0) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    int count = 0;
    for (int i = 0; i < queue->count && count < max_list; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].status == CUSTOMER_WAITING) {
            memcpy(&list[count], &queue->customers[idx], sizeof(Customer));
            count++;
        }
    }
    
    pthread_mutex_unlock(&queue->mutex);
    return count;
}

double queue_estimate_wait_time(WaitQueue *queue, int position) {
    if (queue == NULL || position < 0) return -1.0;
    
    double avg_eating_time = statistics_get_average_eating_time();
    int occupied_tables = 0;
    
    for (int i = 0; i < MAX_TABLES; i++) {
        extern Table g_tables[MAX_TABLES];
        if (g_tables[i].status == TABLE_OCCUPIED || g_tables[i].status == TABLE_EATING) {
            occupied_tables++;
        }
    }
    
    int free_tables = MAX_TABLES - occupied_tables;
    if (position < free_tables) {
        return 0.0;
    }
    
    int effective_position = position - free_tables;
    double estimated_wait = ((double)(effective_position + 1) / MAX_TABLES) * avg_eating_time;
    
    return estimated_wait;
}

int queue_get_customer_position(WaitQueue *queue, int customer_id) {
    if (queue == NULL) return INVALID_PARAM;
    
    pthread_mutex_lock(&queue->mutex);
    
    int position = 0;
    for (int i = 0; i < queue->count; i++) {
        int idx = (queue->front + i) % MAX_QUEUE;
        if (queue->customers[idx].status == CUSTOMER_WAITING) {
            if (queue->customers[idx].id == customer_id) {
                pthread_mutex_unlock(&queue->mutex);
                return position;
            }
            position++;
        }
    }
    
    pthread_mutex_unlock(&queue->mutex);
    return CUSTOMER_NOT_FOUND;
}

int queue_get_max_capacity(void) {
    return MAX_QUEUE;
}
