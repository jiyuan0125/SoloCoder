#include "common.h"
#include <pthread.h>

Statistics g_statistics = {0};

void statistics_init(void) {
    pthread_mutex_init(&g_statistics.mutex, NULL);
    g_statistics.total_eating_time = 0.0;
    g_statistics.total_customers_served = 0;
}

void statistics_update(double eating_time) {
    pthread_mutex_lock(&g_statistics.mutex);
    g_statistics.total_eating_time += eating_time;
    g_statistics.total_customers_served++;
    pthread_mutex_unlock(&g_statistics.mutex);
}

double statistics_get_average_eating_time(void) {
    double avg = DEFAULT_EATING_TIME;
    pthread_mutex_lock(&g_statistics.mutex);
    if (g_statistics.total_customers_served > 0) {
        avg = g_statistics.total_eating_time / g_statistics.total_customers_served;
    }
    pthread_mutex_unlock(&g_statistics.mutex);
    return avg;
}
