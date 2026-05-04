#include "stats.h"

void stats_init(PoolStats *stats) {
    stats->total_allocations = 0;
    stats->total_deallocations = 0;
    stats->leak_recoveries = 0;
    pthread_mutex_init(&stats->mutex, NULL);
}

void stats_destroy(PoolStats *stats) {
    pthread_mutex_destroy(&stats->mutex);
}

void stats_record_allocation(PoolStats *stats) {
    pthread_mutex_lock(&stats->mutex);
    stats->total_allocations++;
    pthread_mutex_unlock(&stats->mutex);
}

void stats_record_deallocation(PoolStats *stats) {
    pthread_mutex_lock(&stats->mutex);
    stats->total_deallocations++;
    pthread_mutex_unlock(&stats->mutex);
}

void stats_record_leak_recovery(PoolStats *stats) {
    pthread_mutex_lock(&stats->mutex);
    stats->leak_recoveries++;
    pthread_mutex_unlock(&stats->mutex);
}

unsigned long long stats_get_total_allocations(PoolStats *stats) {
    unsigned long long result;
    pthread_mutex_lock(&stats->mutex);
    result = stats->total_allocations;
    pthread_mutex_unlock(&stats->mutex);
    return result;
}

unsigned long long stats_get_total_deallocations(PoolStats *stats) {
    unsigned long long result;
    pthread_mutex_lock(&stats->mutex);
    result = stats->total_deallocations;
    pthread_mutex_unlock(&stats->mutex);
    return result;
}

unsigned long long stats_get_leak_recoveries(PoolStats *stats) {
    unsigned long long result;
    pthread_mutex_lock(&stats->mutex);
    result = stats->leak_recoveries;
    pthread_mutex_unlock(&stats->mutex);
    return result;
}
