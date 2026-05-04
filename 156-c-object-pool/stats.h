#ifndef STATS_H
#define STATS_H

#include <pthread.h>

typedef struct PoolStats {
    unsigned long long total_allocations;
    unsigned long long total_deallocations;
    unsigned long long leak_recoveries;
    pthread_mutex_t mutex;
} PoolStats;

void stats_init(PoolStats *stats);
void stats_destroy(PoolStats *stats);
void stats_record_allocation(PoolStats *stats);
void stats_record_deallocation(PoolStats *stats);
void stats_record_leak_recovery(PoolStats *stats);
unsigned long long stats_get_total_allocations(PoolStats *stats);
unsigned long long stats_get_total_deallocations(PoolStats *stats);
unsigned long long stats_get_leak_recoveries(PoolStats *stats);

#endif
