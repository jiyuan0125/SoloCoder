#ifndef BLOCKING_QUEUE_H
#define BLOCKING_QUEUE_H

#include <pthread.h>

typedef struct blocking_queue blocking_queue_t;

blocking_queue_t* bq_create(int capacity);
int bq_put(blocking_queue_t* queue, void* item);
int bq_take(blocking_queue_t* queue, void** out_item);
void bq_close(blocking_queue_t* queue);
int bq_size(blocking_queue_t* queue);
int bq_batch_put(blocking_queue_t* queue, void** items, int count);
int bq_batch_take(blocking_queue_t* queue, void** items, int max_count);
void bq_destroy(blocking_queue_t* queue);

#endif
