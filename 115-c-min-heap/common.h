#ifndef COMMON_H
#define COMMON_H

#include <time.h>

#define MAX_PRIORITY_LEVELS 5
#define MAX_PATIENTS 1000

#define PRIORITY_1_TIME_LIMIT 0
#define PRIORITY_2_TIME_LIMIT 10
#define PRIORITY_3_TIME_LIMIT 30
#define PRIORITY_4_TIME_LIMIT 60
#define PRIORITY_5_TIME_LIMIT 120

typedef struct {
    int id;
    int priority;
    time_t arrival_time;
    time_t queue_time;
    int is_active;
} Patient;

typedef enum {
    PRIORITY_1 = 1,
    PRIORITY_2 = 2,
    PRIORITY_3 = 3,
    PRIORITY_4 = 4,
    PRIORITY_5 = 5
} PriorityLevel;

typedef struct {
    int processed_count;
    double total_wait_time;
    double max_wait_time;
} PriorityStats;

typedef struct {
    PriorityStats priority_stats[MAX_PRIORITY_LEVELS];
    int total_processed;
    int total_patients;
} SystemStats;

int get_priority_time_limit(int priority);
const char* get_priority_description(int priority);

#endif
