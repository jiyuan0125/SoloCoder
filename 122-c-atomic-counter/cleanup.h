#ifndef RATE_LIMITER_CLEANUP_H
#define RATE_LIMITER_CLEANUP_H

#include "counter.h"

typedef struct cleanup_thread {
    rate_limiter_t *limiter;
    pthread_t thread;
    volatile int running;
} cleanup_thread_t;

cleanup_thread_t* cleanup_start(rate_limiter_t *limiter);
void cleanup_stop(cleanup_thread_t *cleanup);

void cleanup_expired_keys(rate_limiter_t *limiter, time_t now);

#endif
