#ifndef COMMON_H
#define COMMON_H

#include <time.h>
#include <pthread.h>

#define MAX_TABLES 20
#define MAX_QUEUE 100
#define DEFAULT_EATING_TIME 60
#define TIMEOUT_THRESHOLD 120

typedef enum {
    CUSTOMER_WAITING = 0,
    CUSTOMER_SEATED,
    CUSTOMER_LEFT,
    CUSTOMER_CANCELLED
} CustomerStatus;

typedef enum {
    TABLE_FREE = 0,
    TABLE_OCCUPIED,
    TABLE_EATING,
    TABLE_DIRTY
} TableStatus;

typedef enum {
    CUSTOMER_NORMAL = 0,
    CUSTOMER_VIP
} CustomerType;

typedef struct {
    int id;
    char name[50];
    CustomerType type;
    CustomerStatus status;
    time_t queue_time;
    time_t seat_time;
    int table_number;
} Customer;

typedef struct {
    int number;
    TableStatus status;
    int customer_id;
    time_t occupied_time;
    time_t eating_start_time;
    int eating_time_estimate;
} Table;

typedef struct {
    double total_eating_time;
    int total_customers_served;
    pthread_mutex_t mutex;
} Statistics;

extern Statistics g_statistics;

void statistics_init(void);
void statistics_update(double eating_time);
double statistics_get_average_eating_time(void);

#endif
