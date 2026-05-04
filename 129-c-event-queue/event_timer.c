#include "event_timer.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/time.h>
#include <time.h>
#include <unistd.h>

#define TIMER_HEAP_INITIAL_CAPACITY 64

extern event_id_t generate_event_id(void);

static uint64_t get_current_time_ms(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000 + (uint64_t)tv.tv_usec / 1000;
}

static timer_heap_t* timer_heap_create(size_t capacity) {
    timer_heap_t *h = (timer_heap_t*)calloc(1, sizeof(timer_heap_t));
    if (!h) return NULL;
    
    h->heap = (timer_event_t**)calloc(capacity, sizeof(timer_event_t*));
    if (!h->heap) {
        free(h);
        return NULL;
    }
    
    h->capacity = capacity;
    h->size = 0;
    return h;
}

static void timer_heap_destroy(timer_heap_t *h) {
    if (!h) return;
    for (size_t i = 0; i < h->size; i++) {
        if (h->heap[i]->data) {
            free(h->heap[i]->data);
        }
        free(h->heap[i]);
    }
    free(h->heap);
    free(h);
}

static int timer_heap_resize(timer_heap_t *h) {
    size_t new_capacity = h->capacity * 2;
    timer_event_t **new_heap = (timer_event_t**)realloc(h->heap, new_capacity * sizeof(timer_event_t*));
    if (!new_heap) return -1;
    
    h->heap = new_heap;
    h->capacity = new_capacity;
    return 0;
}

static void timer_heap_heapify_up(timer_heap_t *h, size_t idx) {
    while (idx > 0) {
        size_t parent = (idx - 1) / 2;
        if (h->heap[idx]->expire_time_ms >= h->heap[parent]->expire_time_ms) {
            break;
        }
        timer_event_t *tmp = h->heap[idx];
        h->heap[idx] = h->heap[parent];
        h->heap[parent] = tmp;
        idx = parent;
    }
}

static void timer_heap_heapify_down(timer_heap_t *h, size_t idx) {
    for (;;) {
        size_t left = 2 * idx + 1;
        size_t right = 2 * idx + 2;
        size_t smallest = idx;
        
        if (left < h->size && h->heap[left]->expire_time_ms < h->heap[smallest]->expire_time_ms) {
            smallest = left;
        }
        if (right < h->size && h->heap[right]->expire_time_ms < h->heap[smallest]->expire_time_ms) {
            smallest = right;
        }
        if (smallest == idx) {
            break;
        }
        
        timer_event_t *tmp = h->heap[idx];
        h->heap[idx] = h->heap[smallest];
        h->heap[smallest] = tmp;
        idx = smallest;
    }
}

static int timer_heap_push(timer_heap_t *h, timer_event_t *event) {
    if (h->size >= h->capacity) {
        if (timer_heap_resize(h) != 0) {
            return -1;
        }
    }
    
    h->heap[h->size] = event;
    timer_heap_heapify_up(h, h->size);
    h->size++;
    return 0;
}

static timer_event_t* timer_heap_pop(timer_heap_t *h) {
    if (h->size == 0) return NULL;
    
    timer_event_t *result = h->heap[0];
    h->heap[0] = h->heap[h->size - 1];
    h->size--;
    h->heap[h->size] = NULL;
    
    if (h->size > 0) {
        timer_heap_heapify_down(h, 0);
    }
    return result;
}

static timer_event_t* timer_heap_peek(timer_heap_t *h) {
    if (h->size == 0) return NULL;
    return h->heap[0];
}

static int timer_heap_erase(timer_heap_t *h, timer_id_t timer_id) {
    for (size_t i = 0; i < h->size; i++) {
        if (h->heap[i]->timer_id == timer_id) {
            if (h->heap[i]->data) {
                free(h->heap[i]->data);
            }
            free(h->heap[i]);
            
            h->heap[i] = h->heap[h->size - 1];
            h->size--;
            h->heap[h->size] = NULL;
            
            if (i < h->size) {
                timer_heap_heapify_down(h, i);
            }
            return 0;
        }
    }
    return -1;
}

static void* timer_thread_func(void *arg) {
    event_timer_t *timer = (event_timer_t*)arg;
    
    pthread_mutex_lock(&timer->mutex);
    
    while (timer->running) {
        uint64_t now = get_current_time_ms();
        uint64_t wait_time_ms = (uint64_t)-1;
        
        while (timer->heap->size > 0) {
            timer_event_t *earliest = timer_heap_peek(timer->heap);
            
            if (earliest->expire_time_ms <= now) {
                timer_event_t *te = timer_heap_pop(timer->heap);
                
                pthread_mutex_unlock(&timer->mutex);
                
                event_channel_post(timer->channel, te->type, te->priority,
                                    te->data, te->data_size, NULL);
                
                free(te->data);
                free(te);
                
                pthread_mutex_lock(&timer->mutex);
                now = get_current_time_ms();
            } else {
                wait_time_ms = earliest->expire_time_ms - now;
                break;
            }
        }
        
        if (timer->running) {
            if (wait_time_ms == (uint64_t)-1) {
                pthread_cond_wait(&timer->cond, &timer->mutex);
            } else {
                struct timespec ts;
                ts.tv_sec = wait_time_ms / 1000;
                ts.tv_nsec = (wait_time_ms % 1000) * 1000000UL;
                pthread_cond_timedwait(&timer->cond, &timer->mutex, &ts);
            }
        }
    }
    
    pthread_mutex_unlock(&timer->mutex);
    return NULL;
}

event_timer_t* event_timer_create(event_channel_t *channel) {
    if (!channel) return NULL;
    
    event_timer_t *timer = (event_timer_t*)calloc(1, sizeof(event_timer_t));
    if (!timer) return NULL;
    
    timer->channel = channel;
    timer->heap = timer_heap_create(TIMER_HEAP_INITIAL_CAPACITY);
    if (!timer->heap) {
        free(timer);
        return NULL;
    }
    
    pthread_mutex_init(&timer->mutex, NULL);
    pthread_cond_init(&timer->cond, NULL);
    timer->running = 0;
    timer->next_timer_id = 1;
    
    return timer;
}

void event_timer_destroy(event_timer_t *timer) {
    if (!timer) return;
    
    if (timer->running) {
        pthread_mutex_lock(&timer->mutex);
        timer->running = 0;
        pthread_cond_signal(&timer->cond);
        pthread_mutex_unlock(&timer->mutex);
        
        pthread_join(timer->timer_thread, NULL);
    }
    
    timer_heap_destroy(timer->heap);
    pthread_mutex_destroy(&timer->mutex);
    pthread_cond_destroy(&timer->cond);
    free(timer);
}

timer_id_t event_timer_post_delayed(event_timer_t *timer, event_type_t type,
                                      event_priority_t priority, void *data, size_t data_size,
                                      uint64_t delay_ms, event_id_t *out_event_id) {
    if (!timer) return 0;
    
    timer_event_t *te = (timer_event_t*)calloc(1, sizeof(timer_event_t));
    if (!te) return 0;
    
    te->timer_id = timer->next_timer_id++;
    te->event_id = generate_event_id();
    te->type = type;
    te->priority = priority;
    te->expire_time_ms = get_current_time_ms() + delay_ms;
    
    if (data && data_size > 0) {
        te->data = malloc(data_size);
        if (!te->data) {
            free(te);
            return 0;
        }
        memcpy(te->data, data, data_size);
        te->data_size = data_size;
    } else {
        te->data = NULL;
        te->data_size = 0;
    }
    
    pthread_mutex_lock(&timer->mutex);
    
    timer_heap_push(timer->heap, te);
    
    if (timer->running) {
        pthread_cond_signal(&timer->cond);
    }
    
    if (out_event_id) {
        *out_event_id = te->event_id;
    }
    
    pthread_mutex_unlock(&timer->mutex);
    
    if (!timer->running) {
        timer->running = 1;
        pthread_create(&timer->timer_thread, NULL, timer_thread_func, timer);
    }
    
    return te->timer_id;
}

int event_timer_cancel(event_timer_t *timer, timer_id_t timer_id) {
    if (!timer || timer_id == 0) return -EINVAL;
    
    pthread_mutex_lock(&timer->mutex);
    int result = timer_heap_erase(timer->heap, timer_id);
    pthread_mutex_unlock(&timer->mutex);
    
    return result;
}
